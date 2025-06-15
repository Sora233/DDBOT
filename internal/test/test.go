package test

import (
	"fmt"
	"testing"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/stretchr/testify/assert"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/MiraiGo-Template/bot"
)

// 这个包只允许在单元测试中使用

const (
	ROOMID1    int64  = 1
	ROOMID2    int64  = 2
	UID1       uint32 = 777
	UID2       uint32 = 778
	UID3       uint32 = 779
	DynamicID1 int64  = 1001
	DynamicID2 int64  = 1002
	MessageID1 int32  = 5001
	MessageID2 int32  = 5002
	G1         uint32 = 123456
	G2         uint32 = 654321
	TIMESTAMP1 int64  = 1624126814
	TIMESTAMP2 int64  = 1624126914

	NAME1 = "name1"
	NAME2 = "name2"

	CMD1 = "command1"
	CMD2 = "command2"

	BVID1 = "bvid1"
	BVID2 = "bvid2"

	ID1 = 2001
	ID2 = 2002

	VersionName = "testVersion"

	Site1 = "site1"
	Site2 = "site2"
	Site3 = "site3"

	Type1 = "type1"
	Type2 = "type2"
	Type3 = "type3"
)

const (
	BibiliLive   concern_type.Type = "bilibiliLive"
	BilibiliNews concern_type.Type = "bilibiliNews"
	DouyuLive    concern_type.Type = "douyuLive"
	YoutubeLive  concern_type.Type = "youtubeLive"
	YoutubeVideo concern_type.Type = "youtubeVideo"
	HuyaLive     concern_type.Type = "huyaLive"
	T1           concern_type.Type = "t1"
	T2           concern_type.Type = "t2"
	T3           concern_type.Type = "t3"
)

var (
	Sender1 = &message.Sender{
		Uin:      uint32(UID1),
		Nickname: NAME1,
	}

	Sender2 = &message.Sender{
		Uin:      uint32(UID2),
		Nickname: NAME2,
	}
)

func InitBuntdb(t *testing.T) {
	assert.Nil(t, localdb.InitBuntDB(localdb.MEMORYDB))
}
func CloseBuntdb(t *testing.T) {
	assert.Nil(t, localdb.Close())
}

func FakeImage(size int) string {
	if size == 0 {
		size = 150
	}
	return fmt.Sprintf("https://via.placeholder.com/%v", size)
}

func InitMirai() {
	bot.QQClient = &bot.Bot{
		QQClient: client.NewClient(123456, "fake"),
	}
}

func CloseMirai() {
	bot.QQClient = nil
}
