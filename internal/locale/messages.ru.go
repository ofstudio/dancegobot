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
	Ok = "Ок"

	CmdDescriptionStart    = "📖 Справка"
	CmdDescriptionSettings = "🔧️ Настройки"
	CmdDescriptionMy       = "🗓️ Мои мероприятия"

	BtnTry           = "👉 Попробовать"
	BtnClose         = "✖️Закрыть"
	BtnBack          = "🔙 Назад"
	BtnChatLink      = "Посмотреть в чате"
	BtnMyPrev        = "‹‹ Пред"
	BtnMyNext        = "След ››"
	BtnEventSettings = "🔧️ Настройки мероприятия"

	ErrSomethingWrong    = "Что-то пошло не так 👾"
	ErrStartPayload      = "Некорректные параметры 👾"
	ErrDancerNameEmpty   = "Имя партнера не может быть пустым 🤔"
	ErrDancerNameTooLong = "Имя партнера слишком длинное 🤔"
	ErrSingleNotFound    = "Такой танцор не найден 🤷‍♀️"

	IconPostClosed  = "🔒 "
	PostCouples     = "👫 <b>Пары</b>\n"
	PostCouplesWait = "\n⏳ <b>Список ожидания</b>\n"

	SignupPlaceholder   = "Введи имя партнера…"
	SignupNotRegistered = "Отправь мне имя партнера или выбери из списка..."
	SignupSingle        = "%s Ты в поиске пары. Если пара уже нашлась, отправь мне имя партнера или выбери из списка..."
	SignupInCouple      = "👫Вы записаны в паре с %s"
	SignupForbidden     = "Тебе запрещено записываться на это мероприятие 😔\n\nОбратись к организатору, чтобы уточнить причину."
	BtnSignupContact    = "👥 Из списка контактов"
	BtnSignupModify     = " Изменить"
	BtnSignupRefresh    = "🔄 Обновить"
	BtnDancerRemove     = "🗑️ Удалить регистрацию"

	ResultSuccessCouple       = "👫 Вы зарегистрировались в паре с %s"
	ResultCoupleWaitlist      = "\n\n⏳ Пока что вы в списке ожидания. Если кто-то отменит регистрацию и вы попадете в список участников, я сообщу об этом 🤗"
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

	MyEventHeader = "🗓️<i>%s</i>\n"
	MyNoEvents    = "🙃 Ты еще никуда не записан и не создал ни одного мероприятия.\n\nПопробуй сейчас!"

	UserSettingsCaption     = "🔧 <b>Настройки для организаторов</b>\n\n"
	UserSettingsDescription = "Эти настройки применяются к новым мероприятиям. Лимит пар можно изменить отдельно в настройках конкретного мероприятия."
	BtnUserSettingsHelp     = "Подробнее о настройках"
	UserSettingsHelp        = `🔧 <b>Настройки для организаторов</b>

🙋‍♀️ <b>Подбор пар</b>
По-умолчанию танцоры могут выбирать любого партнера из списка ожидания.

Если включить автоматический подбор пар, то бот будет самостоятельно составлять пары из танцоров, которые ищут партнера.

👫 <b>Лимит пар</b>
Можно заранее выбрать, сколько пар попадут в основной список участников новых мероприятий. Остальные пары смогут записаться в список ожидания и получат уведомление, если освободится место.

ℹ️ <i>Изменение настроек влияет только на новые мероприятия и не влияет на ранее созданные.</i>

👉 Если добавить бота в группу, то танцоры будут получать уведомления со ссылкой на пост в группе.
`
)

var EventSettingsAutoPair = map[bool]string{
	false: "🙋‍♀️ Можно выбирать из списка ожидания",
	true:  "🙋‍♀️ Пары подбираются автоматически",
}

var BtnEventSettingsAutoPair = map[bool]string{
	false: "🙋‍♀️ Подбирать пару автоматически",
	true:  "🙋‍♀️ Разрешить выбор из списка ожидания",
}

var EventSettingsClosed = map[bool]string{
	false: "🟢 Запись открыта",
	true:  "🔴 Запись закрыта",
}

var BtnEventSettingsClosed = map[bool]string{
	false: "🔴 Закрыть запись",
	true:  "🟢 Открыть запись",
}

const (
	EventSettingsCaption   = "🔧 <b>Настройки мероприятия</b>\n\n"
	EventSettingsLimitNone = "👫 Приходят все записавшиеся пары"
	EventSettingsLimit     = "👫 %s %d %s"
	BtnEventSettingsLimit  = "👫 Ограничить количество пар"
)

var (
	NumLimitCome    = numerals.Ru("Приходит первая", "Приходят первые", "Приходят первые")
	NumLimitCouples = numerals.Ru("пара", "пары", "пар")
)

const (
	BtnEventSettingsLimitNone = "👫 Без ограничений"
	BtnEventSettingsLimitMore = "Больше ››"
	BtnEventSettingsLimitLess = "‹‹ Меньше"
	LimitChangedNoLimit       = "👫 Лимит пар отключен и "
	LimitChangedIncreased     = "👫 Лимит пар увеличился и "
	LimitChangedDecreased     = "👫 Лимит пар уменьшился и "
	BtnLimitChangedNotify     = "🔔 Уведомить участников"
	BtnLimitChangedSkip       = "Не уведомлять"
	LimitChangedNotified      = "Уведомления отправлены 👌"
	LimitChangedCantNotify    = "Я уже не смогу отправить эти уведомления 🙄"
)

var (
	NumLimitIncreased = numerals.Ru(
		"%d пара вышла из списка ожидания:\n\n",
		"%d пары вышли из списка ожидания:\n\n",
		"%d пар вышло из списка ожидания:\n\n",
	)

	NumLimitDecreased = numerals.Ru(
		"%d пара добавлена в список ожидания:\n\n",
		"%d пары добавлены в список ожидания:\n\n",
		"%d пар добавлено в список ожидания:\n\n",
	)
)

const (
	QueryTextEmpty        = "✏️ Напиши текст анонса"
	QueryDescriptionEmpty = "Например: Класс по основам танца 1 марта"
	QueryDescription      = "Нажми для публикации анонса"
	QueryRemaining        = "Осталось %d %s"
	QueryOverflow         = "⚠️ Длина сообщения превышена!"
	QueryEventLimit       = "Лимит %d %s"
)

var NumQueryRemainingSymbols = numerals.Ru("символ", "символа", "символов")

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
