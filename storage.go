package main

import (
	"database/sql"
	"time"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) InsertNewPayment(sheetID *string, categoryID string, id string, amount int64, comment string, time time.Time) error {
	_, err := s.db.Exec("INSERT INTO `payment` (`payment_id`, `sheet_id`, `category_id`, `amount`, `comment`, `payment_made_time`) VALUES (?, ?, ?, ?, ?, ?)",
		id, sheetID, categoryID, amount, comment, time)
	return err
}

func (s *Storage) FindCategory(sheetID *string, categoryName string) (string, error) {
	var categoryID string

	err := s.db.QueryRow("SELECT `category_id` FROM `category` WHERE `sheet_id` = ? AND `name` = ?", sheetID, categoryName).
		Scan(&categoryID)
	if err == sql.ErrNoRows {
		err = nil
	}

	return categoryID, err
}

func (s *Storage) InsertNewCategory(sheetID string, id string, name string) error {
	_, err := s.db.Exec("INSERT INTO `category` (`category_id`, `sheet_id`, `name`) VALUES (?, ?, ?)", id, sheetID, name)
	return err
}

func (s *Storage) ListCategories(sheetID string) ([]string, error) {
	rows, err := s.db.Query("SELECT `name` FROM `category` WHERE `sheet_id` = ?", sheetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (s *Storage) CheckPassword(sheetID string, password string) bool {
	var unused int

	err := s.db.QueryRow("SELECT 1 FROM `sheet` WHERE `sheet_id` = ? AND `password` = PASSWORD(?)", sheetID, password).Scan(&unused)

	return err == nil
}

func (s *Storage) InsertNewSheet(chatID int64, id string, name string, password string) error {
	_, err := s.db.Exec("INSERT INTO `sheet` (`sheet_id`, `owner_chat_id`, `name`, `password`) VALUES (?, ?, ?, PASSWORD(?))",
		id, chatID, name, password)
	return err
}

func (s *Storage) ConnectToSheet(chatID int64, sheetID string) error {
	_, err := s.db.Exec("INSERT INTO `current_sheet` (`chat_id`, `sheet_id`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `sheet_id` = ?", chatID, sheetID, sheetID)
	return err
}

func (s *Storage) FetchCurrentSheetFromDB(chatID int64) (*string, error) {
	var currentSheet string

	err := s.db.QueryRow("SELECT `sheet_id` FROM `current_sheet` WHERE `chat_id` = ?", chatID).Scan(&currentSheet)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	} else {
		return &currentSheet, nil
	}
}

func (s *Storage) DisconnectFromSheet(chatID int64) error {
	_, err := s.db.Exec("DELETE FROM `current_sheet` WHERE `chat_id` = ?", chatID)

	return err
}

type Sheet struct {
	id   string
	name string
}

func (s *Storage) ListSheets(chatID int64) ([]Sheet, error) {
	rows, err := s.db.Query("SELECT `sheet_id`, `name` FROM `sheet` WHERE `owner_chat_id` = ?", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sheets []Sheet
	for rows.Next() {
		var sheet Sheet
		if err := rows.Scan(&sheet.id, &sheet.name); err != nil {
			return nil, err
		}
		sheets = append(sheets, sheet)
	}
	return sheets, nil
}
