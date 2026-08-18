package telegram

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/services"
)

// Membership returns a Telegram adapter for subscription membership checks.
func Membership(api tele.API) services.MembershipFunc {
	return func(chatID, userID int64) (models.Membership, error) {
		member, err := api.ChatMemberOf(&tele.Chat{ID: chatID}, &tele.User{ID: userID})
		if err != nil {
			return models.Membership{}, err
		}
		if member == nil {
			return models.Membership{}, fmt.Errorf("empty chat member")
		}
		return membershipFromChatMember(member), nil
	}
}

func membershipFromChatMember(member *tele.ChatMember) models.Membership {
	return models.Membership{
		Administrator: member.Role == tele.Creator || member.Role == tele.Administrator,
		Member: member.Role == tele.Creator || member.Role == tele.Administrator ||
			member.Role == tele.Member || (member.Role == tele.Restricted && member.Member),
		Banned: member.Role == tele.Kicked,
	}
}
