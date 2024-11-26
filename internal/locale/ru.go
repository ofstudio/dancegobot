package locale

import (
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/numerals"
)

const (
	BtnClose = "✖️Закрыть"
	CloseOK  = "ок"

	Start = "Привет! Я бот для танцевального клуба. Чем могу помочь?" // todo

	ErrNotImplemented    = "Пока в разработке 🚧"
	ErrSomethingWrong    = "Что-то пошло не так 👾"
	ErrStartPayload      = "Некорректные параметры 👾"
	ErrDancerNameTooLong = "Имя партнера слишком длинное 🤔"
	ErrSingleNotFound    = "Такой танцор не найден 🤷‍♀️"

	AnnouncementCouples = "👫 <b>Пары</b>\n"

	SignupPlaceholder   = "Введи имя партнера…"
	SignupNotRegistered = "Отправь мне имя партнера или выбери из списка..."
	SignupSingle        = "%s Ты в поиске пары. Если пара уже нашлась, отправь мне имя партнера или выбери из списка..."
	SignupInCouple      = "👫Вы записаны в паре с %s"
	SignupForbidden     = "Тебе запрещено записываться на это мероприятие 😔\n\nОбратись к организатору, чтобы уточнить причину."
	BtnSignupContact    = "👥 Из списка контактов"
	BtnRemove           = "🗑️ Удалить регистрацию"

	ResultSuccessCouple         = "👫 Вы зарегистрировались в паре с %s"
	ResultSuccessSingle         = "%s Добавил тебя в список ищущих пару.\n\nЕсли кто-то зарегистрируется вместе с тобой, я об этом сообщу 🤗"
	ResultSuccessDeleted        = "Регистрация удалена 🗑"
	ResultAlreadyAsSingle       = "%s Ты в поиске пары. Если пара уже нашлась, отправь мне имя партнера или выбери из списка..."
	ResultAlreadyInCouple       = "Вы уже записаны в паре с %s 🤔\n\nЕсли нужно записаться кем-то другим, удали регистрацию и начни заново."
	ResultAlreadyInSameCouple   = "Вы уже записаны в паре с этим партнером 🤓"
	ResultPartnerTaken          = "Кто-то другой уже записался в паре с %s 😅"
	ResultPartnerSameRole       = "Нельзя записаться с партнером в той же роли, что и ты 🤭"
	ResultSelfNotAllowed        = "Не получится записаться в пару с самим собой 🤓"
	ResultNotRegistered         = "Не могу удалить, так как не вижу в списке участников 🤔"
	ResultEventClosed           = "Сожалеем, но запись на это мероприятие закрыта 😔"
	ResultEventForbiddenDancer  = SignupForbidden
	ResultEventForbiddenPartner = "Твоему партнеру запрещено записываться на это мероприятие 😔\n\nОбратитесь к организатору, чтобы уточнить причину."
	ResultSinglesNotAllowed     = "На это мероприятие можно записаться только в паре 😔"
	ResultSinglesNotAllowedRole = "На это мероприятие можно записаться только в паре 😔"
)

const (
	QueryTextEmpty        = "✏️ Напиши текст анонса"
	QueryDescriptionEmpty = "Например: Класс по основам танца 1 марта"
	QueryDescription      = "Нажми для публикации анонса"
	QueryRemaining        = "Осталось %d %s"
	QueryOverflow         = "⚠️ Длина сообщения превышена!"
)

var NumSymbols = numerals.Ru("символ", "символа", "символов")

type roleMap map[models.Role]string

var RoleIcon = roleMap{
	models.RoleLeader:   "🕺",
	models.RoleFollower: "💃",
}

var AnnouncementSingles = roleMap{
	models.RoleLeader:   "🙋‍♂️ <b>Ищут пару</b>\n",
	models.RoleFollower: "🙋‍♀️ <b>Ищут пару</b>\n",
}

var BtnSingle = roleMap{
	models.RoleLeader:   "🙋‍♂️ Ищу партнершу",
	models.RoleFollower: "🙋‍♀️ Ищу партнера",
}

var IconSingle = roleMap{
	models.RoleLeader:   "🙋‍♂️",
	models.RoleFollower: "🙋‍♀️",
}

var Notifications = map[models.NotificationTmpl]string{
	models.TmplRegisteredWithSingle: "🔔 %s\n\n%s зарегистрировался с тобой в паре! 🎉",
	models.TmplCanceledWithSingle:   "🔔 %s\n\n%s отменил вашу регистрацию. Я вернул тебя в список ищущих пару 🤗",
	models.TmplCanceledByPartner:    "🔔 %s\n\n%s отменил вашу регистрацию.",
}
