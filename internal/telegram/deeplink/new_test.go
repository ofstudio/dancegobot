package deeplink

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

func TestNew(t *testing.T) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	deeplink := New("signup", "eventID", "leader")
	assert.Regexp(t, `^https://t.me/my_bot\?start=[a-zA-Z0-9]{4}-signup-eventID-leader$`, deeplink)
}

func BenchmarkNew(b *testing.B) {
	config.SetBotProfile(&tele.User{Username: "my_bot"})
	for i := 0; i < b.N; i++ {
		_ = New("event_signup", randtoken.New(12), "leader")
	}
}
