package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
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
			test.BibiliLive | test.BilibiliNews, test.BilibiliNews,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.YoutubeVideo,
		},
	}
	var expected = []Type{
		test.BibiliLive,
		test.BilibiliNews | test.BibiliLive,
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Remove(testCase[i][1]), expected[i])
	}
}

func TestType_Split(t *testing.T) {
	var a = test.BibiliLive | test.BilibiliNews | test.DouyuLive | test.YoutubeLive | test.YoutubeVideo | test.HuyaLive
	var expected = map[Type]bool{
		test.BibiliLive:   true,
		test.BilibiliNews: true,
		test.DouyuLive:    true,
		test.YoutubeLive:  true,
		test.YoutubeVideo: true,
		test.HuyaLive:     true,
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
			test.BibiliLive | test.BilibiliNews, test.BibiliLive | test.BilibiliNews,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.YoutubeVideo | test.BibiliLive,
		},
		{
			test.BibiliLive | test.YoutubeLive, test.YoutubeLive | test.BibiliLive,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.BilibiliNews,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.YoutubeLive | test.YoutubeVideo,
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
			test.BibiliLive | test.BilibiliNews, test.BibiliLive | test.BilibiliNews,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.YoutubeVideo | test.BibiliLive,
		},
		{
			test.BibiliLive | test.YoutubeLive, test.YoutubeLive | test.BibiliLive,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.BilibiliNews,
		},
		{
			test.BibiliLive | test.BilibiliNews, test.YoutubeLive | test.YoutubeVideo,
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
		test.BibiliLive, test.DouyuLive, test.BibiliLive | test.BilibiliNews, 0,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], FromString(testCase[i]))
	}
}

func TestType_Empty(t *testing.T) {
	var testCase = []Type{
		0,
		test.BibiliLive,
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
