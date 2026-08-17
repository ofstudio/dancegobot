# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.1] - 2026-08-17

### Changed

- Update Telebot, GJSON, and related transitive dependencies.

### Fixed

- Stabilize event post deep links and closed-event callback data across repeated renders, and treat Telegram's "message is not modified" response as a successful render. This reduces redundant message edits that may cause event posts to disappear for individual users in mobile clients while remaining visible in Telegram Web, as described in [Telegram issue #14753](https://bugs.telegram.org/c/14753).

## [2.1.0] - 2026-05-06

### Added

- [#18](https://github.com/ofstudio/dancegobot/issues/18) `/my` command for event owners and dancers
- Event settings for event owners
- Couple limit and waitlist for event registrations
- Notifications when limit changes move couples into or out of the waitlist
- Default couple limit for new events in organizer settings
- Event owners can close and reopen registration
- Mark events as removed after repeated rendering failures
- [#20](https://github.com/ofstudio/dancegobot/issues/20) Ability to choose a partner from the waiting list by their number

### Changed

- Replace gock-based e2e tests with race-safe direct update processing and a fake Telegram API server

### Fixed

- Notify couples that leave the waitlist after an active couple is removed and the restored partner is auto-paired
- Match manual `@username` registrations in `/my` case-insensitively
- Parse inline limit shortcuts only as separate `/1` through `/99` tokens without matching dates
- Reset stale signup sessions on `/my`, `/settings`, and plain `/start`
- Reject whitespace-only event announcements and partner names
- Escape participant names in Telegram HTML messages so special characters do not break event posts or notifications
- Escape event announcements in Telegram HTML messages so special characters are shown as plain text
- Make random token generation safe for concurrent bot updates
- Clean up completed render repeat tasks so they do not accumulate in memory

## [2.0.4] - 2024-12-22

### Fixed

- Upsert user if profile is updated

## [2.0.3] - 2024-12-20

### Fixed

- Call `deleteWebhook` call on shutdown

## [2.0.2] - 2024-12-18

### Changed

- Rate limit and render repeats config params due to 429 errors from Telegram

## [2.0.1] - 2024-12-16

### Added

- User settings help message

### Fixed

- [#15](https://github.com/ofstudio/dancegobot/issues/15): when `post.inline_message_id` overwrites `post.chat`

## [2.0.0] - 2024-12-16

### Added

- User settings
- [#3](https://github.com/ofstudio/dancegobot/issues/3): auto pairing feature
- Link to original event post in notification message (bot should be a member of a group or channel)
- [#5](https://github.com/ofstudio/dancegobot/issues/5): re-rendering of recent events on bot startup

### Fixed

- [#2](https://github.com/ofstudio/dancegobot/issues/2): event creation bug
- [#6](https://github.com/ofstudio/dancegobot/issues/6): sequential event rendering

## [1.0.2-pre-v2-migration] - 2024-12-16

### Added
- v1 <=> v2 database migration scripts to be able to roll back from v2 to v1.

## [1.0.1] - 2024-11-27

### Fixed

- [#1](https://github.com/ofstudio/dancegobot/issues/1): EventUpdate logging

## [1.0.0] - 2024-11-27

- First production release

## [0.1.0] - 2024-11-26

### Added

- Warn messages on inline query length.

### Fixed

- Send start message only if user session is empty due to some Telegram clients (ie iOS, late 2024) can "double" /start messages on very first user interaction with the bot

## [0.0.1] - 2024-11-25

- Initial release
