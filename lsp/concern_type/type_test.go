package concern_type

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	bibiliLive   Type = "bilibiliLive"
	bilibiliNews Type = "bilibiliNews"
	douyuLive    Type = "douyuLive"
	youtubeLive  Type = "youtubeLive"
	youtubeVideo Type = "youtubeVideo"
	huyaLive     Type = "huyaLive"
)

func TestType_IsTrivial(t *testing.T) {
	assert.True(t, bilibiliNews.IsTrivial())
	assert.True(t, douyuLive.IsTrivial())
	assert.False(t, bilibiliNews.Add(douyuLive).IsTrivial())
}

func TestType_String(t *testing.T) {
	var testType = []Type{
		"1",
		"2",
		"3",
		"4",
	}
	var expected = []string{
		"1",
		"2",
		"3",
		"4",
	}

	assert.Equal(t, len(expected), len(testType))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testType[i].String())
	}
}

func TestType_Add(t *testing.T) {
	var testCase = [][]Type{
		{
			"1", "2",
		},
		{
			"2", "4",
		},
	}
	var expected = []Type{
		"1/2",
		"2/4",
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Add(testCase[i][1]), expected[i])
	}
}

func TestType_Remove(t *testing.T) {
	var testCase = [][]Type{
		{
			bibiliLive.Add(bilibiliNews), bilibiliNews,
		},
		{
			bibiliLive.Add(bilibiliNews), youtubeVideo,
		},
		{
			bibiliLive, bibiliLive,
		},
	}
	var expected = []Type{
		bibiliLive,
		bilibiliNews.Add(bibiliLive),
		Empty,
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Remove(testCase[i][1]), expected[i])
	}
}

func TestType_Split(t *testing.T) {
	var a = Empty.Add(bibiliLive, bilibiliNews, douyuLive, youtubeLive, youtubeVideo, huyaLive)
	var expected = map[Type]bool{
		bibiliLive:   true,
		bilibiliNews: true,
		douyuLive:    true,
		youtubeLive:  true,
		youtubeVideo: true,
		huyaLive:     true,
	}
	var testCase = make(map[Type]bool)
	for _, t := range a.Split() {
		testCase[t] = true
	}
	assert.Equal(t, expected, testCase)
}

func TestType_ContainAll(t *testing.T) {
	var testCase = [][]Type{
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(bibiliLive, bilibiliNews),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(youtubeVideo, bibiliLive),
		},
		{
			Empty.Add(bibiliLive, youtubeLive), Empty.Add(youtubeLive, bibiliLive),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(bilibiliNews),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(youtubeLive, youtubeVideo),
		},
		{
			Empty, Empty,
		},
		{
			Empty.Add(bibiliLive), Empty,
		},
	}
	var expected = []bool{
		true, false, true, true, false, false, true,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i][0].ContainAll(testCase[i][1]))
	}
}

func TestType_ContainAny(t *testing.T) {
	var testCase = [][]Type{
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(bibiliLive, bilibiliNews),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(youtubeVideo, bibiliLive),
		},
		{
			Empty.Add(bibiliLive, youtubeLive), Empty.Add(youtubeLive, bibiliLive),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(bilibiliNews),
		},
		{
			Empty.Add(bibiliLive, bilibiliNews), Empty.Add(youtubeLive, youtubeVideo),
		},
		{
			Empty, Empty,
		},
		{
			Empty.Add(bibiliLive), Empty,
		},
	}
	var expected = []bool{
		true, true, true, true, false, false, true,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i][0].ContainAny(testCase[i][1]))
	}
}

func TestFromString(t *testing.T) {
	var testCase = []string{
		"bilibiliLive", "douyuLive", "bilibiliLive/bilibiliNews", "",
	}
	var expected = []Type{
		bibiliLive, douyuLive, Empty.Add(bibiliLive, bilibiliNews), Empty,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], FromString(testCase[i]))
	}
}

func TestType_Empty(t *testing.T) {
	var testCase = []Type{
		Empty,
		bibiliLive,
	}
	var expected = []bool{
		true,
		false,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i].Empty())
	}
}
