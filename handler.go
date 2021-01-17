package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
)

type Handler struct {
	storage *Storage
	bot     *tgbotapi.BotAPI

	subhandlersByText  map[string]Subhandler
	subhandlersByStage map[ChatStage]Subhandler
	defaultSubhandler  Subhandler
}

type ChatStatus struct {
	chatID int64

	stage   ChatStage
	sheetID *string

	// For CreateSheet* flow
	newSheetName string

	// For ConnectToSheet* flow
	connectToSheetID string
}

type ChatStage int

const (
	None ChatStage = iota

	CreateSheetInputName
	CreateSheetInputPassword
	ConnectToSheetInputID
	ConnectToSheetInputPassword

	CreateCategoryInputName
)

func CreateHandler(storage *Storage, bot *tgbotapi.BotAPI) *Handler {
	h := Handler{storage: storage, bot: bot}

	subhandlers := []Subhandler{
		Subhandler{
			expectedText:  "/createSheet",
			sheetOptional: true,
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = CreateSheetInputName

				return "Create and enter a name for the new sheet"
			},
		},
		Subhandler{
			expectedStage: CreateSheetInputName,
			sheetOptional: true,
			handle: func(name string, chatStatus *ChatStatus) string {
				if errMsg := validateNewSheetName(name); errMsg != "" {
					return errMsg
				}

				chatStatus.newSheetName = name
				chatStatus.stage = CreateSheetInputPassword

				return "Please enter new sheet password"
			},
		},
		Subhandler{
			expectedStage: CreateSheetInputPassword,
			sheetOptional: true,
			handle: func(password string, chatStatus *ChatStatus) string {
				if errMsg := validateNewSheetPassword(password); errMsg != "" {
					return errMsg
				}

				newSheetID := uuid.New().String()

				err := h.storage.InsertNewSheet(chatStatus.chatID, newSheetID, chatStatus.newSheetName, password)
				if err != nil {
					return serverErrorMessage
				}

				chatStatus.sheetID = &newSheetID
				chatStatus.stage = None

				return "New sheet is created!"
			},
		},
		Subhandler{
			expectedText:  "/connectSheet",
			sheetOptional: true,
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = ConnectToSheetInputID

				return "Please enter sheet ID (it is shown when a sheet is created)"
			},
		},
		Subhandler{
			expectedStage: ConnectToSheetInputID,
			sheetOptional: true,
			handle: func(connectToSheetID string, chatStatus *ChatStatus) string {
				chatStatus.connectToSheetID = connectToSheetID
				chatStatus.stage = ConnectToSheetInputPassword

				return "Please enter sheet password"
			},
		},
		Subhandler{
			expectedStage: ConnectToSheetInputPassword,
			sheetOptional: true,
			handle: func(password string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				if h.storage.CheckPassword(chatStatus.connectToSheetID, password) {
					chatStatus.sheetID = &chatStatus.connectToSheetID
					return "Successfully connected to the sheet"
				}

				return "Incorrect password, please try again"
			},
		},
		Subhandler{
			expectedText: "/createCategory",
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = CreateCategoryInputName

				return "Please enter new category name"
			},
		},
		Subhandler{
			expectedStage: CreateCategoryInputName,
			handle: func(name string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				newCategoryID := uuid.New().String()
				err := h.storage.InsertNewCategory(*chatStatus.sheetID, newCategoryID, name)
				if err != nil {
					return serverErrorMessage
				}

				return "New category is created!"
			},
		},
		// default subhandler
		Subhandler{
			handle: func(text string, chatStatus *ChatStatus) string {
				re := regexp.MustCompile(`^(-?\d+(\.\d+)?)\s(.*)$`)
				if matches := re.FindAllStringSubmatch(text, 1); matches != nil {
					amount, _ := strconv.ParseFloat(matches[0][1], 64)
					categoryName := matches[0][3]

					categoryID, err := h.storage.FindCategory(chatStatus.sheetID, categoryName)
					if err != nil {
						return serverErrorMessage
					}
					if len(categoryID) == 0 {
						return "Could not find category with this name"
					}

					newPaymentID := uuid.New().String()
					err = h.storage.InsertNewPayment(chatStatus.sheetID, categoryID, newPaymentID, int64(amount*100), categoryName)
					if err != nil {
						return serverErrorMessage
					}

					return "Successfully created payment record"
				}

				return "Failed to parse"
			},
		},
	}
	h.subhandlersByText = make(map[string]Subhandler)
	h.subhandlersByStage = make(map[ChatStage]Subhandler)
	defaultSubhandlerDefined := false
	for _, subhandler := range subhandlers {
		normalizedExpectedText := normalizeText(subhandler.expectedText)
		if normalizedExpectedText != "" {
			h.subhandlersByText[normalizedExpectedText] = subhandler
		} else if subhandler.expectedStage != None {
			h.subhandlersByStage[subhandler.expectedStage] = subhandler
		} else {
			h.defaultSubhandler = subhandler
			defaultSubhandlerDefined = true
		}
	}
	if !defaultSubhandlerDefined {
		log.Panic("Default subhandler should always be defined")
	}
	return &h
}

func (h *Handler) ProcessUpdate(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	chatID := update.Message.Chat.ID

	replyText := h.replyToMessage(chatID, update.Message.Text)
	h.bot.Send(tgbotapi.NewMessage(chatID, replyText))
}

func (h *Handler) replyToMessage(chatID int64, text string) string {
	chatStatus, err := h.getChatStatus(chatID)
	if err != nil {
		return serverErrorMessage
	}

	var sh Subhandler
	sh, ok := h.subhandlersByText[normalizeText(text)]
	if !ok {
		sh, ok = h.subhandlersByStage[chatStatus.stage]
		if !ok {
			sh = h.defaultSubhandler
		}
	}
	if !sh.sheetOptional && chatStatus.sheetID == nil {
		return "You are not yet connected to a sheet. Please either create a new one or connect to an existing one.\n" +
			"/createSheet /connectSheet"
	}
	return sh.handle(text, chatStatus)
}

func validateNewSheetPassword(password string) string {
	if strings.HasPrefix(password, "/") {
		return "Sheet password shouldn't start with /"
	}

	if len(password) < 3 {
		return "Sheet password should be at least 3 characters long"
	}

	return ""
}

func validateNewSheetName(name string) string {
	if strings.HasPrefix(name, "/") {
		return "Sheet name shouldn't start with /"
	}

	if len(name) < 3 {
		return "Sheet name should be at least 3 characters long"
	}

	return ""
}

func normalizeText(text string) string {
	return strings.TrimSpace(strings.ToLower(text))
}

var chatStatuses = make(map[int64]*ChatStatus)

func (h *Handler) getChatStatus(chatID int64) (*ChatStatus, error) {
	if status, ok := chatStatuses[chatID]; ok {
		return status, nil
	}

	currentSheetID, err := h.storage.FetchCurrentSheetFromDB(chatID)
	if err != nil {
		return nil, err
	}

	status := &ChatStatus{chatID: chatID, sheetID: currentSheetID}
	chatStatuses[chatID] = status
	return status, nil
}

const serverErrorMessage = "Unexpected server error"

type Subhandler struct {
	expectedText  string
	expectedStage ChatStage

	sheetOptional bool

	handle func(text string, status *ChatStatus) string
}
