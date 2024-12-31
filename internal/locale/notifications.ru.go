package locale

import (
	"github.com/ofstudio/dancegobot/internal/models"
)

var Notifications = map[models.NotificationTmpl]string{
	// language=GoTemplate
	models.TmplRegisteredWithSingle: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} зарегистрировался с тобой в паре! 🎉`,

	// language=GoTemplate
	models.TmplCanceledWithSingle: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию. Я вернул тебя в список ищущих пару 🤗`,

	// language=GoTemplate
	models.TmplCanceledByPartner: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию.`,

	// language=GoTemplate
	models.TmplAutoPairPartnerFound: `🔔 {{.Event.Caption}}

Я подобрал тебе в пару {{fmtDancer .Partner}} 👌`,

	// language=GoTemplate
	models.TmplAutoPairPartnerChanged: `🔔 {{.Event.Caption}}

{{fmtDancer .Partner}} отменил вашу регистрацию. 
Я записал тебя вместе с {{fmtDancer .NewPartner}} 👌`,

	// language=GoTemplate
	models.TmplEventLimitIncreased: `🔔 {{.Event.Caption}}

{{fmtProfile .Event.Owner}} увеличил лимит пар и вы вместе с {{fmtDancer .Partner}} вышли из списка ожидания 🎉

Если планы изменились, и вы не сможете принять участие, пожалуйста, отмените вашу регистрацию.`,

	// language=GoTemplate
	models.TmplEventLimitDecreased: `🔔 {{.Event.Caption}}

{{fmtProfile .Event.Owner}} уменьшил лимит пар и вы вместе с {{fmtDancer .Partner}} теперь в списке ожидания.

Если кто-то отменит регистрацию и вы попадете в список участников, то я сообщу об этом 🤗`,
}
