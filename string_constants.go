package main

const (
	MESSAGE_UNEXPECTED_SERVER_ERROR = "Unexpected server error"

	MESSAGE_START_GREETING = "Hi, BudgliBot for your service!"
	MESSAGE_START_CONNECT  = `To start using the bot you need to either
- Create your own sheet with /createSheet
- Or connect to an existing one (that someone else created) with /connectSheet`
	MESSAGE_START_FULL_HELP = "Click /help to get the list of all the commands"

	MESSAGE_HELP = `
- To add a new payment record, simply type "<amount> <category>", e.g "42 groceries". The category with this name must exist prior to this. TBD: It will soon be possible to add a category if it doesn't exist.

Categories:
- To add a new category, click /createCategory
- To list your categories, click /listCategories

Sheets:
- To add a new sheet, click /createSheet, but you are very likely to only need one
- To connect to a sheet, click /connectSheet
- To disconnect from the current sheet, click /disconnectSheet
- To list all your sheets, click /listSheets`

	MESSAGE_INPUT_NEW_SHEET_NAME               = "Create and enter a name for the new sheet"
	MESSAGE_INPUT_NEW_SHEET_PASSWORD           = "Please enter new sheet password"
	MESSAGE_CREATED_NEW_SHEET                  = "New sheet is created!\nName: %s\nID: %s"
	MESSAGE_INPUT_SHEET_ID                     = "Please enter sheet ID (it is shown when a sheet is created)"
	MESSAGE_INPUT_SHEET_PASSWORD               = "Please enter sheet password"
	MESSAGE_INCORRECT_SHEET_ID_FORMAT          = "Incorrect sheet ID format, expected to be xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx, e.g e72e1f4c-fb53-4455-9f0e-a1e9d0e1bc4d"
	MESSAGE_SUCCESS_CONNECT_TO_SHEET           = "Successfully connected to the sheet"
	MESSAGE_INCORRECT_PASSWORD                 = "Incorrect password, please try again"
	MESSAGE_INCORRECT_NEW_PASSWORD_SLASH       = "Sheet password shouldn't start with /"
	MESSAGE_INCORRECT_NEW_PASSWORD_TOO_SHORT   = "Sheet password should be at least 3 characters long"
	MESSAGE_INCORRECT_NEW_SHEET_NAME_SLASH     = "Sheet name shouldn't start with /"
	MESSAGE_INCORRECT_NEW_SHEET_NAME_TOO_SHORT = "Sheet name should be at least 3 characters long"
	MESSAGE_SUCCESS_DISCONNECT_SHEET           = "Successfully disconnected from the sheet"
	MESSAGE_LIST_SHEETS_INTRO                  = "Your user owns the following %d sheets:"
	MESSAGE_LIST_SHEETS_OUTRO                  = `To add new sheets (if you have one, you are very unlikely to need more), click /createSheet
To connect to one of these or other sheets, click /connectSheet`

	MESSAGE_FAILURE_UNKNOWN_CATEGORY_NAME = "Could not find category with this name"
	MESAGE_SUCCESS_CREATE_PAYMENT         = "Successfully created payment record"
	MESSAGE_FAILURE_PARSING               = "Failed to parse\n\n" + MESSAGE_START_FULL_HELP

	MESSAGE_INPUT_CATEGORY_NAME     = "Please enter new category name"
	MESSAGE_SUCCESS_CREATE_CATEGORY = "New category is created!"
	MESSAGE_LIST_CATEGORIES_INTRO   = "This sheet has the following %d categories:"
	MESSAGE_LIST_CATEGORIES_OUTRO   = "To add new categories, click /createCategory"
)
