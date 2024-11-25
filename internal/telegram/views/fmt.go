package views

import (
	"regexp"
	"strconv"

	"github.com/ofstudio/dancegobot/internal/models"
)

// fmtName formats the fullname with a link to the Telegram profile.
// If profile is not provided, the full name is returned.
//
// If profile has a username, the link is created to the username. Example:
//
//	<a href='https://t.me/username'>Full Name</a>
//
// If the profile has no username, the link is created to the user ID. Example:
//
//	<a href='tg://user?id=123456789'>Full Name</a>
func fmtName(fullname string, profile *models.Profile) string {
	if profile == nil {
		return fullname
	}
	if profile.Username != "" {
		return "<a href='https://t.me/" + profile.Username + "'>" + fullname + "</a>"
	}
	return "<a href='tg://user?id=" + strconv.FormatInt(profile.ID, 10) + "'>" + fullname + "</a>"
}

func fmtDancerName(d *models.Dancer) string {
	return fmtName(d.FullName, d.Profile)
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
