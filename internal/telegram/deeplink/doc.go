// Package deeplink provides a way to generate deep links for Telegram bots.
// Example:
//
//	https://t.me/botname?start=payload
//
// When a user clicks on the link, the bot will receive the /start command with the payload.
// The payload can be used to pass additional information to the bot.
//
// Payload limitations:
//   - 1-64 characters
//   - only A-Z, a-z, 0-9, _ and - are allowed
package deeplink
