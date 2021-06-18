package concern

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestType_String(t *testing.T) {
	var testType = []Type{
		1,
		2,
		3,
		4,
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
			1, 2,
		},
		{
			2, 4,
		},
	}
	var expected = []Type{
		3,
		6,
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Add(testCase[i][1]), expected[i])
	}
}

func TestType_Remove(t *testing.T) {
	var testCase = [][]Type{
		{
			BibiliLive | BilibiliNews, BilibiliNews,
		},
		{
			BibiliLive | BilibiliNews, YoutubeVideo,
		},
	}
	var expected = []Type{
		BibiliLive,
		BilibiliNews | BibiliLive,
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Remove(testCase[i][1]), expected[i])
	}
}

func TestType_Split(t *testing.T) {
	var a = BibiliLive | BilibiliNews | DouyuLive | YoutubeLive | YoutubeVideo | HuyaLive
	var expected = map[Type]bool{
		BibiliLive:   true,
		BilibiliNews: true,
		DouyuLive:    true,
		YoutubeLive:  true,
		YoutubeVideo: true,
		HuyaLive:     true,
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
			BibiliLive | BilibiliNews, BibiliLive | BilibiliNews,
		},
		{
			BibiliLive | BilibiliNews, YoutubeVideo | BibiliLive,
		},
		{
			BibiliLive | YoutubeLive, YoutubeLive | BibiliLive,
		},
		{
			BibiliLive | BilibiliNews, BilibiliNews,
		},
		{
			BibiliLive | BilibiliNews, YoutubeLive | YoutubeVideo,
		},
		{
			0, 0,
		},
	}
	var expected = []bool{
		true, false, true, true, false, false,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i][0].ContainAll(testCase[i][1]))
	}
}

func TestType_ContainAny(t *testing.T) {
	var testCase = [][]Type{
		{
			BibiliLive | BilibiliNews, BibiliLive | BilibiliNews,
		},
		{
			BibiliLive | BilibiliNews, YoutubeVideo | BibiliLive,
		},
		{
			BibiliLive | YoutubeLive, YoutubeLive | BibiliLive,
		},
		{
			BibiliLive | BilibiliNews, BilibiliNews,
		},
		{
			BibiliLive | BilibiliNews, YoutubeLive | YoutubeVideo,
		},
		{
			0, 0,
		},
	}
	var expected = []bool{
		true, true, true, true, false, false,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i][0].ContainAny(testCase[i][1]))
	}
}

func TestFromString(t *testing.T) {
	var testCase = []string{
		"1", "4", "3", "error",
	}
	var expected = []Type{
		BibiliLive, DouyuLive, BibiliLive | BilibiliNews, 0,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], FromString(testCase[i]))
	}
}

func TestType_Description(t *testing.T) {
	var testCase = []Type{
		BibiliLive, BilibiliNews, DouyuLive, YoutubeVideo, BilibiliNews | BibiliLive,
	}
	var expected = []string{
		"live", "news", "live", "news", "live/news",
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], testCase[i].Description())
	}
}

func TestType_Empty(t *testing.T) {
	var testCase = []Type{
		0,
		BibiliLive,
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
