package main

import (
	"fmt"
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

				return MESSAGE_INPUT_NEW_SHEET_NAME
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

				return MESSAGE_INPUT_NEW_SHEET_PASSWORD
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
					return MESSAGE_UNEXPECTED_SERVER_ERROR
				}

				chatStatus.sheetID = &newSheetID
				chatStatus.stage = None

				return fmt.Sprintf(MESSAGE_CREATED_NEW_SHEET, chatStatus.newSheetName, newSheetID)
			},
		},
		Subhandler{
			expectedText:  "/connectSheet",
			sheetOptional: true,
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = ConnectToSheetInputID

				return MESSAGE_INPUT_SHEET_ID
			},
		},
		Subhandler{
			expectedStage: ConnectToSheetInputID,
			sheetOptional: true,
			handle: func(connectToSheetID string, chatStatus *ChatStatus) string {
				chatStatus.connectToSheetID = connectToSheetID
				chatStatus.stage = ConnectToSheetInputPassword

				return MESSAGE_INPUT_SHEET_PASSWORD
			},
		},
		Subhandler{
			expectedStage: ConnectToSheetInputPassword,
			sheetOptional: true,
			handle: func(password string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				if h.storage.CheckPassword(chatStatus.connectToSheetID, password) {
					chatStatus.sheetID = &chatStatus.connectToSheetID
					return MESSAGE_SUCCESS_CONNECT_TO_SHEET
				}

				return MESSAGE_INCORRECT_PASSWORD
			},
		},
		Subhandler{
			expectedText:  "/disconnectSheet",
			sheetOptional: true,
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.sheetID = nil
				chatStatus.stage = None

				if err := h.storage.DisconnectFromSheet(chatStatus.chatID); err != nil {
					return MESSAGE_UNEXPECTED_SERVER_ERROR
				}

				return MESSAGE_SUCCESS_DISCONNECT_SHEET
			},
		},
		Subhandler{
			expectedText:  "/listSheets",
			sheetOptional: true,
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				sheets, err := h.storage.ListSheets(chatStatus.chatID)
				if err != nil {
					return MESSAGE_UNEXPECTED_SERVER_ERROR
				}

				var reply strings.Builder
				fmt.Fprintf(&reply, MESSAGE_LIST_SHEETS_INTRO, len(sheets))
				reply.WriteString("\n\n")
				for i, sheet := range sheets {
					fmt.Fprintf(&reply, "%2d. Name: %s\n    ID: %s\n\n", i+1, sheet.name, sheet.id)
				}
				reply.WriteString("\n\n")
				reply.WriteString(MESSAGE_LIST_SHEETS_OUTRO)
				return reply.String()
			},
		},
	}
}

func validateNewSheetPassword(password string) string {
	if strings.HasPrefix(password, "/") {
		return MESSAGE_INCORRECT_NEW_PASSWORD_SLASH
	}

	if len(password) < 3 {
		return MESSAGE_INCORRECT_NEW_PASSWORD_TOO_SHORT
	}

	return ""
}

func validateNewSheetName(name string) string {
	if strings.HasPrefix(name, "/") {
		return MESSAGE_INCORRECT_NEW_SHEET_NAME_SLASH
	}

	if len(name) < 3 {
		return MESSAGE_INCORRECT_NEW_SHEET_NAME_TOO_SHORT
	}

	return ""
}
