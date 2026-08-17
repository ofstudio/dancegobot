package views

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

type editAPI struct {
	tele.API
	err error
}

func (a editAPI) Edit(tele.Editable, interface{}, ...interface{}) (*tele.Message, error) {
	return nil, a.err
}

func Test_renderEditResult(t *testing.T) {
	failure := errors.New("edit failed")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "inline edit success", err: tele.ErrTrueResult},
		{name: "message not modified", err: tele.ErrMessageNotModified},
		{name: "same message content", err: tele.ErrSameMessageContent},
		{name: "other error", err: failure, want: failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := render(editAPI{err: tt.err}, &models.Event{ID: "eventID"}, "inlineMessageID")

			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
		})
	}
}

func Test_btnEventClosedStable(t *testing.T) {
	first := btnEventClosed()
	second := btnEventClosed()

	assert.Equal(t, first, second)
	assert.Equal(t, BtnEventClosed.Unique, first.InlineKeyboard[0][0].Unique)
	assert.Equal(t, "v1", first.InlineKeyboard[0][0].Data)
}

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

func Test_postTextBuilderEscapesCaption(t *testing.T) {
	event := &models.Event{
		Caption: "Танцы <tag> & friends",
	}

	text := postTextBuilder(event).String()

	assert.Contains(t, text, "Танцы &lt;tag&gt; &amp; friends")
	assert.NotContains(t, text, "Танцы <tag> & friends")
}
