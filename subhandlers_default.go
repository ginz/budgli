package main

import (
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

func getDefaultSubhandler(h *Handler) Subhandler {
	return Subhandler{
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
	}
}
