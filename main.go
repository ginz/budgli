package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"

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

	handler := CreateHandler(storage, bot)
	for update := range updates {
		handler.ProcessUpdate(&update)
	}
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
