package locale

import (
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/numerals"
)

const (
	Start = `Привет! Это бот для записи на танцы

📣 Публикую анонсы мероприятий
🙋‍♀️ Записываю в парах и поодиночке
🙌 Помогаю танцорам найти пару
🔔 Отправляю уведомления

Для создания записи напиши в своей группе или канале:

<b>@%s [Текст анонса]</b>

…и нажми «Опубликовать»
`
	CmdDescriptionStart    = "📖 Справка"
	CmdDescriptionSettings = "⚙️ Настройки"

	BtnTry   = "👉 Попробовать"
	BtnClose = "✖️Закрыть"
	BtnBack  = "🔙 Назад"
	Ok       = "Ок"

	ErrNotImplemented    = "Пока в разработке 🚧"
	ErrSomethingWrong    = "Что-то пошло не так 👾"
	ErrStartPayload      = "Некорректные параметры 👾"
	ErrDancerNameTooLong = "Имя партнера слишком длинное 🤔"
	ErrSingleNotFound    = "Такой танцор не найден 🤷‍♀️"

	PostCouples = "👫 <b>Пары</b>\n"

	SignupPlaceholder   = "Введи имя партнера…"
	SignupNotRegistered = "Отправь мне имя партнера или выбери из списка..."
	SignupSingle        = "%s Ты в поиске пары. Если пара уже нашлась, отправь мне имя партнера или выбери из списка..."
	SignupInCouple      = "👫Вы записаны в паре с %s"
	SignupForbidden     = "Тебе запрещено записываться на это мероприятие 😔\n\nОбратись к организатору, чтобы уточнить причину."
	BtnSignupContact    = "👥 Из списка контактов"
	BtnRemove           = "🗑️ Удалить регистрацию"

	ResultSuccessCouple       = "👫 Вы зарегистрировались в паре с %s"
	ResultSuccessSingle       = "%s Добавил тебя в список ищущих пару.\n\nЕсли кто-то зарегистрируется вместе с тобой, я об этом сообщу 🤗"
	ResultSuccessRemoved      = "Регистрация удалена 🗑"
	ResultAlreadyAsSingle     = "%s Ты в поиске пары. Если пара уже нашлась, отправь мне имя партнера или выбери из списка..."
	ResultAlreadyInCouple     = "Вы уже записаны в паре с %s 🤔\n\nЕсли нужно записаться кем-то другим, удали регистрацию и начни заново."
	ResultAlreadyInSameCouple = "Вы уже записаны в паре с этим партнером 🤓"
	ResultPartnerTaken        = "Кто-то другой уже записался в паре с %s 😅"
	ResultPartnerSameRole     = "Нельзя записаться с партнером в той же роли, что и ты 🤭"
	ResultSelfNotAllowed      = "Не получится записаться в пару с самим собой 🤓"
	ResultNotRegistered       = "Не могу удалить, так как не вижу в списке участников 🤔"
	ResultEventClosed         = "Сожалеем, но запись на это мероприятие закрыта 😔"
	ResultEventRemoved        = "Кажется, запись на это мероприятие удалена 🤔"
	ResultDancerForbidden     = SignupForbidden
	ResultPartnerForbidden    = "Твоему партнеру запрещено записываться на это мероприятие 😔\n\nОбратитесь к организатору, чтобы уточнить причину."
	ResultClosedForSingles    = "На это мероприятие можно записаться только в паре 😔"
	ResultClosedForSingleRole = "На это мероприятие можно записаться только в паре 😔"
)

const (
	SettingsCaption = "🔧 <b>Настройки для организаторов</b>\n\n"
	BtnSettingsHelp = "Подробнее о настройках"

	SettingsHelp = `🔧 <b>Настройки для организаторов</b>

🙋‍♀️ <b>Подбор пар</b>
По-умолчанию танцоры могут выбирать любого партнера из списка ожидания.

Если включить автоматический подбор пар, то бот будет самостоятельно составлять пары из танцоров, которые ищут партнера.

ℹ️ <i>Изменение настроек влияет только на новые мероприятия и не влияет на ранее созданные.</i>

👉 Если добавить бота в группу, то танцоры будут получать уведомления со ссылкой на пост в группе.
`
)

var SettingsAutoPairing = map[bool]string{
	false: "🙋‍♀️ Можно выбирать из списка ожидания",
	true:  "🙋‍♀️ Пары подбираются автоматически",
}

var BtnAutoPairing = map[bool]string{
	false: "🙋‍♀️ Подбирать пару автоматически",
	true:  "🙋‍♀️ Разрешить выбор из списка ожидания",
}

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

var PostSingles = roleMap{
	models.RoleLeader:   "🙋‍♂️ <b>Ищут пару</b>\n",
	models.RoleFollower: "🙋‍♀️ <b>Ищут пару</b>\n",
}

var BtnAsSingle = roleMap{
	models.RoleLeader:   "🙋‍♂️ Ищу партнершу",
	models.RoleFollower: "🙋‍♀️ Ищу партнера",
}

var IconSingle = roleMap{
	models.RoleLeader:   "🙋‍♂️",
	models.RoleFollower: "🙋‍♀️",
}

const BtnChatLink = "Посмотреть"
