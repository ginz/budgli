package main

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func getCategorySubhandlers(h *Handler) []Subhandler {
	return []Subhandler{
		Subhandler{
			expectedText: "/createCategory",
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = CreateCategoryInputName

				return MESSAGE_INPUT_CATEGORY_NAME
			},
		},
		Subhandler{
			expectedStage: CreateCategoryInputName,
			handle: func(name string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				newCategoryID := uuid.New().String()
				err := h.storage.InsertNewCategory(*chatStatus.sheetID, newCategoryID, name)
				if err != nil {
					return MESSAGE_UNEXPECTED_SERVER_ERROR
				}

				return MESSAGE_SUCCESS_CREATE_CATEGORY
			},
		},
		Subhandler{
			expectedText: "/listCategories",
			handle: func(text string, chatStatus *ChatStatus) string {
				chatStatus.stage = None

				categories, err := h.storage.ListCategories(*chatStatus.sheetID)
				if err != nil {
					return MESSAGE_UNEXPECTED_SERVER_ERROR
				}

				var reply strings.Builder
				fmt.Fprintf(&reply, MESSAGE_LIST_CATEGORIES_INTRO, len(categories))
				reply.WriteString("\n\n")
				for i, category := range categories {
					fmt.Fprintf(&reply, "%2d. %s\n", i+1, category)
				}
				reply.WriteString("\n\n")
				reply.WriteString(MESSAGE_LIST_CATEGORIES_OUTRO)

				return reply.String()
			},
		},
	}
}
