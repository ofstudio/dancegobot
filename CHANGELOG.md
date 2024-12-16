# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [v2.0.0] - 2024-12-16

- Added auto pairing feature (issue #3)
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
