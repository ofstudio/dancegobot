# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [v2.0.2] - 2024-12-16

- Fixed bug when `post.inline_message_id` overwrites `post.chat` (issue #15)
- Added user settings help

# [v2.0.1] - 2024-12-18

- Changed rate limit and render repeats config params due to 429 errors from Telegram

# [v2.0.0] - 2024-12-16

- Added user settings
- Added auto pairing feature (issue #3)
- Added link to original event post in notification message (bot should be a member of a group or channel)
- Fixed bug when event is not created (issue #2)
- Added re-rendering of recent events on bot startup (issue #5)
- Added sequential event rendering (issue #6)

## [v1.0.2-pre-v2-migration] - 2024-12-16

- Added v1 ↔︎ v2 database migration scripts to be able to roll back from v2 to v1.

## [v1.0.1] - 2024-11-27

- Fixed EventUpdate logging (#1)

## [v1.0.0] - 2024-11-27

- First production release

## [v0.1.0] - 2024-11-26

- Added warn messages on inline query length.
- Send start message only if user session is empty due to some Telegram clients (ie iOS, late 2024) can "double" /start messages on very first user interaction with the bot

## [v0.0.1] - 2024-11-25

- Initial release
