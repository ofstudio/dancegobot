// Package telelog provides a simple way to log [gopkg.in/telebot.v4] types using [log/slog].
package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

func ExampleAttr() {
	// This example demonstrates how to log a tele.User using telelog.Attr.
	user := tele.User{
		ID:        123456789,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}
	slog.Info("User joined", Attr(user))
	// Example Output:
	// 2024/11/14 18:53:42 INFO User joined user.id=123456789 user.first_name=John user.username=johndoe
}
