# dancegobot

Telegram bot for finding a partner for dance events.

Inspired by [Tayrinn/CoopDance](https://github.com/Tayrinn/CoopDance).

## Features

- ✅ Announcement publishing via bot inline query: `@dancegobot <announcement text>`.
- ✅ Dancer can sign up in a couple with a partner or as single looking for a partner.
- ✅ Partner can be selected from the contact list or by username or by name as a free text.
- ✅ Partner can be selected from the list of single dancers.
- ✅ Notifications to the single dancer when someone selects them as a partner.

## Installation

See `Dockerfile` and `docker-compose.yml` for an example of how to run the bot in a Docker container.

## Configuration

Configuration is done via environment variables.

| Variable                 | Default value              | Description                                                                                                                                                                                                        |
|--------------------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `DB_FILEPATH`            | –                          | **_Required._** Path to SQLite database file. Note that inside Docker container, `DB_FILEPATH` is already set to `/data/dancegobot.db` by default. See Dockerfile for more details.                                |
| `BOT_TOKEN`              | –                          | **_Required._** Telegram bot token.                                                                                                                                                                                |
| `BOT_API_URL`            | `https://api.telegram.org` | _Optional._ Telegram bot API URL.                                                                                                                                                                                  |
| `BOT_USE_WEBHOOK`        | `false`                    | _Optional._ Should bot use [webhook](https://core.telegram.org/bots/webhooks) or [long polling](https://core.telegram.org/bots/api#getupdates) for receiving updates. Default is `false` which means long polling. |
| `BOT_WEBHOOK_LISTEN`     | `:8080 `                   | _Optional._ Host and port to listen for incoming webhooks. Only used if BOT_USE_WEBHOOK is true.                                                                                                                   |
| `BOT_WEBHOOK_PUBLIC_URL` | –                          | _Optional._ Public URL for the webhook. Only used if BOT_USE_WEBHOOK is true. Note that bot doesn't implement TLS termination, so it should be done by a reverse proxy like Nginx or Traefik.                      |
| `THUMBNAIL_URL`          | –                          | _Optional._ URL to a thumbnail image that will be used for announcement inline query answer. It should be a square image.                                                                                          |

## License

Apache License 2.0

## Contributing

Feel free to open an issue or a pull request.

## Author

Oleg Fomin [@ofstudio](https://t.me/ofstudio)
