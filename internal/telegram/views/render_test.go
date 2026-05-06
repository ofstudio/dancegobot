package views

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

func Test_fmtDancerEscapesHTML(t *testing.T) {
	t.Run("manual name", func(t *testing.T) {
		got := fmtDancer(models.Dancer{
			FullName: "Мария <Follower> & Co",
		})

		assert.Equal(t, "Мария &lt;Follower&gt; &amp; Co", got)
	})

	t.Run("profile name", func(t *testing.T) {
		profile := &models.Profile{
			ID:        42,
			FirstName: "Иван <Lead>",
			LastName:  "& Co",
		}
		got := fmtDancer(models.Dancer{
			Profile:  profile,
			FullName: profile.FullName(),
		})

		assert.Equal(t, `<a href="tg://user?id=42">Иван &lt;Lead&gt; &amp; Co</a>`, got)
	})
}

func Test_postTextBuilderEscapesDancerNames(t *testing.T) {
	leaderProfile := &models.Profile{
		ID:        42,
		FirstName: "Иван <Lead>",
		LastName:  "& Co",
	}
	event := &models.Event{
		Caption: "Test Event",
		Couples: []models.Couple{{
			Dancers: []models.Dancer{
				{
					Profile:  leaderProfile,
					FullName: leaderProfile.FullName(),
					Role:     models.RoleLeader,
				},
				{
					FullName: "Мария <Follower> & Co",
					Role:     models.RoleFollower,
				},
			},
		}},
	}

	text := postTextBuilder(event).String()

	assert.Contains(t, text, locale.PostCouples)
	assert.Contains(t, text, `<a href="tg://user?id=42">Иван &lt;Lead&gt; &amp; Co</a>`)
	assert.Contains(t, text, "Мария &lt;Follower&gt; &amp; Co")
	assert.NotContains(t, text, "Иван <Lead>")
	assert.NotContains(t, text, "Мария <Follower>")
}
