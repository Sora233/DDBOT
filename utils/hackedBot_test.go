package utils

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBot(t *testing.T) {
	bot := GetBot()
	assert.Nil(t, bot.FindFriend(1))
	assert.Nil(t, bot.FindGroup(test.G1))
	assert.False(t, bot.IsOnline())
	assert.Empty(t, bot.GetFriendList())
	assert.Empty(t, bot.GetGroupList())
	bot.SolveFriendRequest(nil, false)
	bot.SolveGroupJoinRequest(nil, false, false, "")

	test.InitMirai()
	defer test.CloseMirai()

	assert.NotNil(t, hackedBot.Bot)
	(*hackedBot.Bot).Online = true
	assert.True(t, hackedBot.IsOnline())
}
