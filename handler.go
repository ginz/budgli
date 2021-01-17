package main

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	subhandlers := append(getSheetSubhandlers(&h), getCategorySubhandlers(&h)...)
	subhandlers = append(subhandlers, getDefaultSubhandler(&h))
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
		return MESSAGE_UNEXPECTED_SERVER_ERROR
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

type Subhandler struct {
	expectedText  string
	expectedStage ChatStage

	sheetOptional bool

	handle func(text string, status *ChatStatus) string
}
