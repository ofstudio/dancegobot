package telegock

const base = "https://api.telegram.org/bot.*/"

const (
	GetUpdates          = base + "getUpdates"
	GetMe               = base + "getMe"
	SetMyCommands       = base + "setMyCommands"
	SendMessage         = base + "sendMessage"
	EditMessageText     = base + "editMessageText"
	AnswerInlineQuery   = base + "answerInlineQuery"
	AnswerCallbackQuery = base + "answerCallbackQuery"
)
