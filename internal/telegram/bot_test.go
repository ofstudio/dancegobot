package telegram

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestPoller(t *testing.T) {
	cfg := config.Bot{
		UseWebhook:       true,
		WebhookListen:    ":1234",
		WebhookPublicURL: "https://example.com/webhook",
		AllowedUpdates:   []string{"message", "callback_query"},
		Timeout:          5 * time.Second,
	}
	p := poller(cfg)
	require.IsType(t, &tele.Webhook{}, p)
	webhook := p.(*tele.Webhook)
	require.Equal(t, cfg.WebhookListen, webhook.Listen)
	require.Equal(t, cfg.WebhookPublicURL, webhook.Endpoint.PublicURL)
	require.Equal(t, cfg.AllowedUpdates, webhook.AllowedUpdates)
	require.Len(t, webhook.SecretToken, 64)
	require.Regexp(t, `^[a-zA-Z0-9]{64}$`, webhook.SecretToken)
	require.NotEqual(t, webhook.SecretToken, poller(cfg).(*tele.Webhook).SecretToken)

	cfg.UseWebhook = false
	p = poller(cfg)
	require.IsType(t, &tele.LongPoller{}, p)
	longPoller := p.(*tele.LongPoller)
	require.Equal(t, cfg.Timeout, longPoller.Timeout)
	require.Equal(t, cfg.AllowedUpdates, longPoller.AllowedUpdates)
}

func TestNewBotPollingMode(t *testing.T) {
	for _, mode := range []string{"webhook", "long polling"} {
		t.Run(mode, func(t *testing.T) {
			bot, tg := newBotTest(t, mode == "webhook")
			if mode == "webhook" {
				require.IsType(t, &tele.Webhook{}, bot.Poller)
				return
			}
			require.IsType(t, &tele.LongPoller{}, bot.Poller)
			tg.Wait("deleteWebhook")
		})
	}
}

func TestWebhookSecretAuthentication(t *testing.T) {
	bot, tg := newBotTest(t, true)
	webhook := bot.Poller.(*tele.Webhook)
	updates := make(chan tele.Update, 1)
	stop := make(chan struct{})
	close(stop)
	// With no listener and a closed stop channel, Poll initializes ServeHTTP synchronously.
	webhook.Poll(bot, updates, stop)

	registration := tg.Wait("setWebhook")
	require.Len(t, registration.String("secret_token"), 64)
	require.Equal(t, webhook.SecretToken, registration.String("secret_token"))
	require.Equal(t, webhook.Endpoint.PublicURL, registration.String("url"))
	allowedUpdates, err := json.Marshal(webhook.AllowedUpdates)
	require.NoError(t, err)
	require.JSONEq(t, string(allowedUpdates), registration.String("allowed_updates"))

	update := tele.Update{ID: 42, Message: &tele.Message{
		ID:   123,
		Chat: &tele.Chat{ID: 42, Type: tele.ChatPrivate},
		Text: "/start",
	}}
	body, err := json.Marshal(update)
	require.NoError(t, err)
	wrongSecret := []byte(webhook.SecretToken)
	wrongSecret[0] = 'a'
	if webhook.SecretToken[0] == 'a' {
		wrongSecret[0] = 'b'
	}
	tests := []struct {
		name      string
		secret    string
		delivered bool
	}{
		{name: "missing secret"},
		{name: "wrong secret", secret: string(wrongSecret)},
		{name: "truncated secret", secret: webhook.SecretToken[:63]},
		{name: "correct secret", secret: webhook.SecretToken, delivered: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, webhook.Endpoint.PublicURL, bytes.NewReader(body))
			if tt.secret != "" {
				req.Header.Set("X-Telegram-Bot-Api-Secret-Token", tt.secret)
			}
			webhook.ServeHTTP(httptest.NewRecorder(), req)
			select {
			case received := <-updates:
				require.True(t, tt.delivered, "unexpected update delivery")
				require.Equal(t, update, received)
			default:
				require.False(t, tt.delivered, "expected update delivery")
			}
		})
	}
}

func newBotTest(t *testing.T, useWebhook bool) (*tele.Bot, *teletest.Server) {
	t.Helper()
	tg := teletest.New(t)
	tg.Ignore("getMe", "setMyCommands")
	t.Cleanup(tg.AssertNoUnexpected)
	cfg := config.Default().Bot
	cfg.ApiURL = tg.URL()
	cfg.Token = "123456:ABCDEF"
	cfg.UseWebhook = useWebhook
	cfg.WebhookListen = ""
	cfg.WebhookPublicURL = "https://example.com/webhook"
	cfg.RPS = 1000
	cfg.Timeout = 5 * time.Second
	bot, err := NewBot(cfg, nil)
	require.NoError(t, err)
	return bot, tg
}
