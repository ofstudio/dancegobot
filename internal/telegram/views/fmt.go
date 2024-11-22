package views

import (
	"regexp"
	"strconv"

	"github.com/ofstudio/dancegobot/internal/models"
)

// FmtName formats the dancer with a link to the profile.
// If Telegram profile is not provided, the full name is returned.
//
// If profile has a username, the link is created to the username. Example:
//
//	<a href='https://t.me/username'>Full Name</a>
//
// If the profile has no username, the link is created to the user ID. Example:
//
//	<a href='tg://user?id=123456789'>Full Name</a>
func FmtName(d *models.Dancer) string {
	if d.Profile == nil {
		return d.FullName
	}
	if d.Profile.Username != "" {
		return "<a href='https://t.me/" + d.Profile.Username + "'>" + d.FullName + "</a>"
	}
	return "<a href='tg://user?id=" + strconv.FormatInt(d.Profile.ID, 10) + "'>" + d.FullName + "</a>"
}

// FmtSingles makes [models.SessionSingle] from the list of singles with given role.
// Returns the list of profiles with reply button captions.
// Caption format: "1. Full Name (@username)"
// or just "1. Full Name" if no Telegram username.
func FmtSingles(singles []models.Dancer, role models.Role) []models.SessionSingle {
	var s []models.SessionSingle
	for i, d := range singles {
		if d.Profile == nil {
			continue
		}
		if d.Role == role {
			caption := strconv.Itoa(i+1) + ". " + d.FullName
			if d.Profile.Username != "" {
				caption += " (@" + d.Profile.Username + ")"
			}
			s = append(s, models.SessionSingle{
				Caption: caption,
				Profile: *d.Profile,
			})

		}
	}
	return s
}

var reSingleCapt = regexp.MustCompile(`^\d+\. .+$`)

// IsSingleCaption checks if the text is a single button caption.
func IsSingleCaption(text string) bool {
	return reSingleCapt.MatchString(text)
}
