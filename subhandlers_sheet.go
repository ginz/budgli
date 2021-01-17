package main

import (
	"strings"

	"github.com/google/uuid"
)

func getSheetSubhandlers(h *Handler) []Subhandler {
	return []Subhandler{
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
	}
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
