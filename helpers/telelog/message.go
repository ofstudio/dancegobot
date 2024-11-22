package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// MessageValue returns a slog.Value for the given tele.Message.
// Supported message types:
//   - text
//   - audio
//   - document
//   - photo
//   - sticker
//   - video
//   - voice
//   - video_note
//   - animation
//   - contact
//   - location
//   - user_shared
//   - chat_shared
//   - venue
//   - poll
//   - dice
func MessageValue(v tele.Message) slog.Value {
	var t string

	switch {
	case v.Text != "":
		t = "text"
	case v.Audio != nil:
		t = "audio"
	case v.Document != nil:
		t = "document"
	case v.Photo != nil:
		t = "photo"
	case v.Sticker != nil:
		t = "sticker"
	case v.Video != nil:
		t = "video"
	case v.Voice != nil:
		t = "voice"
	case v.VideoNote != nil:
		t = "video_note"
	case v.Animation != nil:
		t = "animation"
	case v.Contact != nil:
		t = "contact"
	case v.Location != nil:
		t = "location"
	case v.UserShared != nil:
		t = "user_shared"
	case v.ChatShared != nil:
		t = "chat_shared"
	case v.Venue != nil:
		t = "venue"
	case v.Poll != nil:
		t = "poll"
	case v.Dice != nil:
		t = "dice"
	default:
		t = "unsupported"
	}

	attrs := []slog.Attr{
		slog.Int("id", v.ID),
		slog.String("type", t),
	}

	if v.Chat != nil {
		attrs = append(attrs, slog.Any("chat", ChatValue(*v.Chat)))
		// Add sender only if it's not a private chat.
		if v.Chat.Type != tele.ChatPrivate && v.Sender != nil {
			attrs = append(attrs, slog.Any("from", UserValue(*v.Sender)))
		}
	}

	return slog.GroupValue(attrs...)
}
