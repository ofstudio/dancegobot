package telegock

const base = "https://api.telegram.org/bot.*/"

const (
	GetUpdates          = base + "getUpdates"
	GetMe               = base + "getMe"
	SetMyCommands       = base + "setMyCommands"
	DeleteWebhook       = base + "deleteWebhook"
	SendMessage         = base + "sendMessage"
	EditMessageText     = base + "editMessageText"
	EditMessageMarkup   = base + "editMessageReplyMarkup"
	AnswerInlineQuery   = base + "answerInlineQuery"
	AnswerCallbackQuery = base + "answerCallbackQuery"
)
