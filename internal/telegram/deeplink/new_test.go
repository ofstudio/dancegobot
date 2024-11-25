package deeplink

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ofstudio/dancegobot/pkg/randtoken"
)

func TestNew(t *testing.T) {
	SetBotName("my_bot")
	deeplink := New("signup", "eventID", "leader")
	assert.Regexp(t, `^https://t.me/my_bot\?start=[a-zA-Z0-9]{4}-signup-eventID-leader$`, deeplink)
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New("my_bot", "event_signup", randtoken.New(12), "leader")
	}
}
