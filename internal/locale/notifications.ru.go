package locale

import (
	"github.com/ofstudio/dancegobot/internal/models"
)

const NotificationsBase =
// language=GoTemplate
`{{define "waitlist"}}{{if .WaitList}}

Ваша пара находится в списке ожидания. Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗{{end}}{{end}}`

var Notifications = map[models.NotificationTmpl]string{
	// language=GoTemplate
	models.TmplNewEvent: `🔔 Новая запись в {{if .Event.Post.Chat.Title}}{{.Event.Post.Chat.Title}}{{else if .Event.Post.Chat.Username}}@{{.Event.Post.Chat.Username}}{{else}}чате{{end}}.`,

	// language=GoTemplate
	models.TmplRegisteredWithSingle: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} зарегистрировался с тобой в паре! 🎉{{template "waitlist" .}}`,

	// language=GoTemplate
	models.TmplCanceledWithSingle: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию. Я вернул тебя в список ищущих пару 🤗`,

	// language=GoTemplate
	models.TmplCanceledByPartner: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию.`,

	// language=GoTemplate
	models.TmplAutoPairPartnerFound: `🔔 {{.Event.Caption}}

Я подобрал тебе в пару {{fmtDancer .Partner}} 👌{{template "waitlist" .}}`,

	// language=GoTemplate
	models.TmplAutoPairPartnerChanged: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию. 
Я записал тебя вместе с {{fmtDancer .NewPartner}} 👌{{template "waitlist" .}}`,

	// language=GoTemplate
	models.TmplCoupleWaitListLeft: `🔔 {{.Event.Caption}}

Вы вместе с {{fmtDancer .Partner}} вышли из списка ожидания 🎉

Если планы изменились, и вы не сможете принять участие, пожалуйста, отмените вашу регистрацию.`,

	// language=GoTemplate
	models.TmplEventLimitIncreased: `🔔 {{.Event.Caption}}

{{fmtProfile .Event.Owner}} увеличил лимит пар и вы вместе с {{fmtDancer .Partner}} вышли из списка ожидания 🎉

Если планы изменились, и вы не сможете принять участие, пожалуйста, отмените вашу регистрацию.`,

	// language=GoTemplate
	models.TmplEventLimitDecreased: `🔔 {{.Event.Caption}}

{{fmtProfile .Event.Owner}} уменьшил лимит пар и вы вместе с {{fmtDancer .Partner}} теперь в списке ожидания.

Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗`,
}
