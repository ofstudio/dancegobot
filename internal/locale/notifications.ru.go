package locale

import (
	"github.com/ofstudio/dancegobot/internal/models"
)

//goland:noinspection HtmlUnknownTarget
const NotificationsBase =
// language=GoTemplate
`{{define "dancer"}}<a href="{{urlTo .Profile}}">{{.FullName}}</a>{{end}}`

var Notifications = map[models.NotificationTmpl]string{
	// language=GoTemplate
	models.TmplRegisteredWithSingle: `🔔 {{.Event.Caption}}

{{template "dancer" .Partner}} зарегистрировался с тобой в паре! 🎉`,

	// language=GoTemplate
	models.TmplCanceledWithSingle: `🔔 {{.Event.Caption}}

{{template "dancer" .Partner}} отменил вашу регистрацию. Я вернул тебя в список ищущих пару 🤗`,

	// language=GoTemplate
	models.TmplCanceledByPartner: `🔔 {{.Event.Caption}}

{{template "dancer" .Partner}} отменил вашу регистрацию.`,

	// language=GoTemplate
	models.TmplAutoPairPartnerFound: `🔔 {{.Event.Caption}}

Я подобрал тебе в пару {{template "dancer" .Partner}} 👌`,

	// language=GoTemplate
	models.TmplAutoPairPartnerChanged: `🔔 {{.Event.Caption}}

{{template "dancer" .Partner}} отменил вашу регистрацию. 
Я записал тебя вместе с {{template "dancer" .NewPartner}} 👌`,
}
