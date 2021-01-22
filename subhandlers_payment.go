package main

import (
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func getPaymentSubhandlers(h *Handler) []Subhandler {
	return []Subhandler{
		// default subhandler
		Subhandler{
			handle: func(text string, chatStatus *ChatStatus, _ *ReplyExtras) string {
				re := regexp.MustCompile(`^(-?\d+(\.\d+)?)\s(.*)$`)
				if matches := re.FindAllStringSubmatch(text, 1); matches != nil {
					amount, _ := strconv.ParseFloat(matches[0][1], 64)
					categoryName := matches[0][3]

					categoryID, err := h.storage.FindCategory(chatStatus.sheetID, categoryName)
					if err != nil {
						return MESSAGE_UNEXPECTED_SERVER_ERROR
					}
					if len(categoryID) == 0 {
						return MESSAGE_FAILURE_UNKNOWN_CATEGORY_NAME
					}

					newPaymentID := uuid.New().String()
					err = h.storage.InsertNewPayment(chatStatus.sheetID, categoryID, newPaymentID, int64(amount*100), categoryName, time.Now())
					if err != nil {
						return MESSAGE_UNEXPECTED_SERVER_ERROR
					}

					return MESAGE_SUCCESS_CREATE_PAYMENT
				}

				return MESSAGE_FAILURE_PARSING
			},
		},
	}
}
