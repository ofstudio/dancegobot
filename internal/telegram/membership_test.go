package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"
)

func TestMembership(t *testing.T) {
	tests := []struct {
		name          string
		member        tele.ChatMember
		administrator bool
		isMember      bool
		banned        bool
	}{
		{name: "creator", member: tele.ChatMember{Role: tele.Creator}, administrator: true, isMember: true},
		{name: "administrator", member: tele.ChatMember{Role: tele.Administrator}, administrator: true, isMember: true},
		{name: "member", member: tele.ChatMember{Role: tele.Member}, isMember: true},
		{name: "restricted member", member: tele.ChatMember{Role: tele.Restricted, Member: true}, isMember: true},
		{name: "restricted non-member", member: tele.ChatMember{Role: tele.Restricted}},
		{name: "left", member: tele.ChatMember{Role: tele.Left}},
		{name: "kicked", member: tele.ChatMember{Role: tele.Kicked}, banned: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			membership := membershipFromChatMember(&tt.member)
			require.Equal(t, tt.administrator, membership.Administrator)
			require.Equal(t, tt.isMember, membership.Member)
			require.Equal(t, tt.banned, membership.Banned)
		})
	}
}
