package telegock

import (
	"fmt"

	tele "gopkg.in/telebot.v4"
)

// ResponseResult is a response from the Telegram API with some kind of result.
type ResponseResult struct {
	Ok     bool `json:"ok"`
	Result any  `json:"result"`
}

func Result(v any) *ResponseResult {
	return &ResponseResult{
		Ok:     true,
		Result: v,
	}
}

// ResponseUpdates is a response from the Telegram API with updates.
type ResponseUpdates struct {
	Ok     bool          `json:"ok"`
	Result []tele.Update `json:"result"`
}

var (
	updateID  int
	messageID int
)

func Updates(upd ...tele.Update) *ResponseUpdates {
	for i := range upd {
		updateID++
		upd[i].ID = updateID
	}
	return &ResponseUpdates{
		Ok:     true,
		Result: upd,
	}
}

func (r *ResponseUpdates) Message(msg tele.Message) *ResponseUpdates {
	updateID++
	if msg.ID == 0 {
		messageID++
		msg.ID = messageID
	}
	r.Result = append(r.Result, tele.Update{ID: updateID, Message: &msg})
	return r
}

func (r *ResponseUpdates) InlineQuery(query tele.Query) *ResponseUpdates {
	updateID++
	if query.ID == "" {
		query.ID = fmt.Sprintf("inline_query_id_%d", updateID)
	}
	r.Result = append(r.Result, tele.Update{ID: updateID, Query: &query})
	return r
}

func (r *ResponseUpdates) InlineResult(result tele.InlineResult) *ResponseUpdates {
	updateID++
	if result.MessageID == "" {
		result.MessageID = fmt.Sprintf("inline_message_id_%d", updateID)
	}
	r.Result = append(r.Result, tele.Update{ID: updateID, InlineResult: &result})
	return r
}

func (r *ResponseUpdates) CallbackQuery(query tele.Callback) *ResponseUpdates {
	if query.ID == "" {
		updateID++
		query.ID = fmt.Sprintf("callback_query_id_%d", updateID)
	}
	r.Result = append(r.Result, tele.Update{ID: updateID, Callback: &query})
	return r
}
