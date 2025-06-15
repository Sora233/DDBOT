package utils

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestGetBot(t *testing.T) {
	defer GetBot().TESTReset()

	bot := GetBot()
	bot.TESTSetUin(test.UID1)
	assert.EqualValues(t, test.UID1, bot.GetUin())
	assert.Nil(t, bot.FindFriend(1))
	assert.Nil(t, bot.FindGroup(test.G1))
	assert.False(t, bot.IsOnline())
	assert.Empty(t, bot.GetFriendList())
	assert.Empty(t, bot.GetGroupList())
	bot.SolveFriendRequest(nil, false)
	bot.SolveGroupJoinRequest(nil, false, false, "")

	bot.TESTAddGroup(123)
	bot.TESTAddGroup(test.G2)
	bot.TESTAddGroup(test.G1)
	bot.TESTAddMember(test.G1, test.UID1, entity.Admin)
	bot.TESTAddMember(test.G1, test.UID2, entity.Admin)
	bot.TESTAddMember(test.G1, test.UID1, entity.Admin)
	bot.TESTAddMember(test.G2, test.UID2, entity.Admin)
	assert.Len(t, bot.GetGroupList(), 3)
	bot.TESTReset()
	assert.Empty(t, bot.GetGroupList())

	test.InitMirai()
	defer test.CloseMirai()

	assert.NotNil(t, hackedBot.Bot)
	(*hackedBot.Bot).Online.Store(true)
	assert.True(t, hackedBot.IsOnline())
}
