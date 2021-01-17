package main

import "github.com/google/uuid"

func getCategorySubhandlers(h *Handler) []Subhandler {
	return []Subhandler{
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
	}
}
