package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	conf, err := readConfiguration()
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotKey)
	if err != nil {
		log.Panic(err)
	}

	storage, err := initStorage(conf)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = conf.Debug

	if conf.Debug {
		log.Printf("Authorized on account %s", bot.Self.UserName)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		processUpdate(storage, bot, &update)
	}
}

type ChatStatus struct {
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

var chatStatuses = make(map[int64]*ChatStatus)

func processUpdate(storage *Storage, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	chatID := update.Message.Chat.ID

	chatStatus, err := getChatStatus(storage, chatID)
	if err != nil {
		sendErrorMessage(bot, chatID)
		return
	}

	if update.Message.Text == "/createSheet" {
		chatStatus.stage = CreateSheetInputName

		msg := tgbotapi.NewMessage(chatID, "Create and enter a name for the new sheet")
		bot.Send(msg)
		return
	}

	if update.Message.Text == "/connectSheet" {
		chatStatus.stage = ConnectToSheetInputID

		msg := tgbotapi.NewMessage(chatID, "Please enter sheet ID (it is shown when a sheet is created)")
		bot.Send(msg)
		return
	}

	if chatStatus.stage == CreateSheetInputName {
		name := update.Message.Text

		if errMsg := validateNewSheetName(name); errMsg != "" {
			msg := tgbotapi.NewMessage(chatID, errMsg)
			bot.Send(msg)
			return
		}

		chatStatus.newSheetName = name
		chatStatus.stage = CreateSheetInputPassword

		msg := tgbotapi.NewMessage(chatID, "Please enter new sheet password")
		bot.Send(msg)
		return
	}

	if chatStatus.stage == CreateSheetInputPassword {
		password := update.Message.Text

		if errMsg := validateNewSheetPassword(password); errMsg != "" {
			msg := tgbotapi.NewMessage(chatID, errMsg)
			bot.Send(msg)
			return
		}

		newSheetID := uuid.New().String()

		err := storage.InsertNewSheet(chatID, newSheetID, chatStatus.newSheetName, password)
		if err != nil {
			sendErrorMessage(bot, chatID)
			return
		}

		chatStatus.sheetID = &newSheetID
		chatStatus.stage = None

		msg := tgbotapi.NewMessage(chatID, "New sheet is created!")
		bot.Send(msg)
		return
	}

	if chatStatus.stage == ConnectToSheetInputID {
		connectToSheetID := update.Message.Text

		chatStatus.connectToSheetID = connectToSheetID
		chatStatus.stage = ConnectToSheetInputPassword

		msg := tgbotapi.NewMessage(chatID, "Please enter sheet password")
		bot.Send(msg)
		return
	}

	if chatStatus.stage == ConnectToSheetInputPassword {
		password := update.Message.Text

		chatStatus.stage = None

		if storage.CheckPassword(chatStatus.connectToSheetID, password) {
			chatStatus.sheetID = &chatStatus.connectToSheetID
			msg := tgbotapi.NewMessage(chatID, "Successfully connected to the sheet")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Incorrect password, please try again")
		bot.Send(msg)
		return
	}

	if chatStatus.sheetID == nil {
		msg := tgbotapi.NewMessage(chatID, "You are not yet connected to a sheet. Please either create a new one or connect to an existing one.\n"+
			"/createSheet /connectSheet")

		bot.Send(msg)
		return
	}

	if update.Message.Text == "/createCategory" {
		chatStatus.stage = CreateCategoryInputName

		msg := tgbotapi.NewMessage(chatID, "Please enter new category name")

		bot.Send(msg)
		return
	}

	if chatStatus.stage == CreateCategoryInputName {
		name := update.Message.Text

		chatStatus.stage = None

		newCategoryID := uuid.New().String()
		err := storage.InsertNewCategory(*chatStatus.sheetID, newCategoryID, name)
		if err != nil {
			sendErrorMessage(bot, chatID)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "New category is created!")
		bot.Send(msg)
		return
	}

	// TODO: compile once
	re := regexp.MustCompile(`^(-?\d+(\.\d+)?)\s(.*)$`)
	if matches := re.FindAllStringSubmatch(update.Message.Text, 1); matches != nil {
		amount, _ := strconv.ParseFloat(matches[0][1], 64)
		categoryName := matches[0][3]

		categoryID, err := storage.FindCategory(chatStatus.sheetID, categoryName)
		if err != nil {
			sendErrorMessage(bot, chatID)
			return
		}
		if len(categoryID) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Could not find category with this name")
			bot.Send(msg)
			return
		}

		newPaymentID := uuid.New().String()
		err = storage.InsertNewPayment(chatStatus.sheetID, categoryID, newPaymentID, int64(amount*100), categoryName)
		if err != nil {
			sendErrorMessage(bot, chatID)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Successfully created payment record")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Failed to parse")
	bot.Send(msg)
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

func getChatStatus(storage *Storage, chatID int64) (*ChatStatus, error) {
	if status, ok := chatStatuses[chatID]; ok {
		return status, nil
	}

	currentSheetID, err := storage.FetchCurrentSheetFromDB(chatID)
	if err != nil {
		return nil, err
	}

	status := &ChatStatus{sheetID: currentSheetID}
	chatStatuses[chatID] = status
	return status, nil
}

func sendErrorMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Uexpected server error")

	bot.Send(msg)
}

type Conf struct {
	SQLConnection  string
	TelegramBotKey string
	Debug          bool
}

func readConfiguration() (*Conf, error) {
	var err error

	raw, err := ioutil.ReadFile("conf.json")
	if err != nil {
		return nil, err
	}
	var conf Conf
	err = json.Unmarshal(raw, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func initStorage(conf *Conf) (*Storage, error) {
	db, err := sql.Open("mysql", conf.SQLConnection)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}
