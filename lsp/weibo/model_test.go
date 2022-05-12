package weibo

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserInfo(t *testing.T) {
	userInfo := &UserInfo{
		Uid:             test.UID1,
		Name:            test.NAME1,
		ProfileImageUrl: test.FakeImage(10),
		ProfileUrl:      test.FakeImage(20),
	}
	assert.EqualValues(t, test.UID1, userInfo.GetUid())
	assert.EqualValues(t, test.NAME1, userInfo.GetName())
	assert.NotNil(t, userInfo.Logger())
	assert.EqualValues(t, Site, userInfo.Site())
}

func TestNewsInfo(t *testing.T) {
	userInfo := &UserInfo{
		Uid:             test.UID1,
		Name:            test.NAME1,
		ProfileImageUrl: test.FakeImage(10),
		ProfileUrl:      test.FakeImage(20),
	}
	newsInfo := &NewsInfo{
		UserInfo:     userInfo,
		LatestNewsTs: test.DynamicID1,
		Cards: []*Card{
			{
				CardType: CardType_Unknown,
				Mblog: &Card_Mblog{
					CreatedAt: "Mon Jan 02 15:04:05 -0700 2006",
				},
			},
			{
				CardType: CardType_Normal,
				Mblog: &Card_Mblog{
					RawText: "raw",
				},
				Scheme: "https://localho.st?a=b",
			},
			{
				CardType: CardType_Normal,
				Mblog: &Card_Mblog{
					Pics: []*Card_Mblog_Pics{
						{
							Url: test.FakeImage(10),
						},
					},
					RetweetedStatus: &Card_Mblog{
						User: &ApiContainerGetIndexProfileResponse_Data_UserInfo{
							ScreenName: test.NAME2,
						},
					},
				},
			},
		},
	}
	assert.EqualValues(t, News, newsInfo.Type())
	assert.NotNil(t, newsInfo.Logger())

	concernNews := NewConcernNewsNotify(mt.NewGroupTarget(test.G1), newsInfo)
	assert.NotNil(t, concernNews)
	assert.Len(t, concernNews, len(newsInfo.Cards))

	for _, concernNewsNotify := range concernNews {
		assert.EqualValues(t, News, concernNewsNotify.Type())
		assert.True(t, concernNewsNotify.GetTarget().Equal(mt.NewGroupTarget(test.G1)))
		assert.NotNil(t, concernNewsNotify.Logger())
		assert.NotNil(t, concernNewsNotify.ToMessage())
	}
}
