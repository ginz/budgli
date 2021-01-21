package main

func getInfoSubhandlers(h *Handler) []Subhandler {
	return []Subhandler{
		Subhandler{
			expectedText:  "/start",
			sheetOptional: true,
			handle: func(_ string, chatStatus *ChatStatus, _ *ReplyExtras) string {
				if chatStatus.sheetID != nil {
					return MESSAGE_START_GREETING + "\n\n" + MESSAGE_START_FULL_HELP
				}

				return MESSAGE_START_GREETING + "\n\n" + MESSAGE_START_CONNECT + "\n\n" + MESSAGE_START_FULL_HELP
			},
		},
		Subhandler{
			expectedText:  "/help",
			sheetOptional: true,
			handle: func(_ string, _ *ChatStatus, _ *ReplyExtras) string {
				return MESSAGE_HELP
			},
		},
	}
}
