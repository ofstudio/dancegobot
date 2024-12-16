# dancegobot

Telegram bot for finding a partner for dance events.

Inspired by [Tayrinn/CoopDance](https://github.com/Tayrinn/CoopDance).

## Features

- ✅ Event publishing via bot inline query: `@dancegobot <announcement text>`.
- ✅ Dancer can sign up in a couple with a partner or as single looking for a partner.
- ✅ Automatic pairing of single dancers.
- ✅ Notifications to the dancer when someone selects them as a partner.

## Installation

### Docker
Obtain a bot token from [@BotFather](https://t.me/botfather) and run the following command:

```bash
docker run --name dancegobot \
  -e BOT_TOKEN=<your_bot_token> \
  -v /path/to/database:/data \
    ghcr.io/ofstudio/dancegobot:latest
```
This will start the bot in [long polling mode](https://core.telegram.org/bots/api#getupdates) 
with SQLite database stored in `/path/to/database/dancegobot.db`.

To run specific version of the bot, replace `latest` with the desired version tag, for example `v2.0.0`.
Version tags can be found at [Packages page](https://github.com/ofstudio/dancegobot/pkgs/container/dancegobot).
See `CHANGELOG.md` for the version history.

### Docker Compose
See `docker-compose.yaml` for an example of running the bot in [webhook mode](https://core.telegram.org/bots/webhooks)
with Traefik reverse proxy and Let's Encrypt TLS certificates.

### Build from source
1. Clone the repository.
2. Download dependencies: `go mod download`.
3. Build the bot: `go build -o dancegobot ./cmd/dancegobot`.

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
