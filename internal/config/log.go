package config

import "log/slog"

func (b Bot) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("api_url", b.ApiURL),
		slog.Int("rps", b.RPS),
		slog.Duration("timeout", b.Timeout),
		slog.Any("allowed_updates", b.AllowedUpdates),
	}
	if b.UseWebhook {
		attrs = append(
			attrs,
			slog.String("poller_type", "webhook"),
			slog.String("webhook_listen", b.WebhookListen),
			slog.String("webhook_public_url", b.WebhookPublicURL),
		)
	} else {
		attrs = append(attrs, slog.String("poller_type", "long_poll"))
	}

	return slog.GroupValue(attrs...)
}

func (b DB) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("driver", "sqlite"),
		slog.String("file", b.Filepath),
		slog.Uint64("required_version", uint64(b.RequiredVersion)),
	)
}
