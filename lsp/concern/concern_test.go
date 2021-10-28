package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestType_String(t *testing.T) {
	var testType = []concern_type.Type{
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
	var testCase = [][]concern_type.Type{
		{
			"1", "2",
		},
		{
			"2", "4",
		},
	}
	var expected = []concern_type.Type{
		"1/2",
		"2/4",
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Add(testCase[i][1]), expected[i])
	}
}

func TestType_Remove(t *testing.T) {
	var testCase = [][]concern_type.Type{
		{
			test.BibiliLive.Add(test.BilibiliNews), test.BilibiliNews,
		},
		{
			test.BibiliLive.Add(test.BilibiliNews), test.YoutubeVideo,
		},
	}
	var expected = []concern_type.Type{
		test.BibiliLive,
		test.BilibiliNews.Add(test.BibiliLive),
	}

	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, testCase[i][0].Remove(testCase[i][1]), expected[i])
	}
}

func TestType_Split(t *testing.T) {
	var a = concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews, test.DouyuLive, test.YoutubeLive, test.YoutubeVideo, test.HuyaLive)
	var expected = map[concern_type.Type]bool{
		test.BibiliLive:   true,
		test.BilibiliNews: true,
		test.DouyuLive:    true,
		test.YoutubeLive:  true,
		test.YoutubeVideo: true,
		test.HuyaLive:     true,
	}
	var testCase = make(map[concern_type.Type]bool)
	for _, t := range a.Split() {
		testCase[t] = true
	}
	assert.Equal(t, expected, testCase)
}

func TestType_ContainAll(t *testing.T) {
	var testCase = [][]concern_type.Type{
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.YoutubeVideo, test.BibiliLive),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.YoutubeLive), concern_type.Empty.Add(test.YoutubeLive, test.BibiliLive),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.BilibiliNews),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.YoutubeLive, test.YoutubeVideo),
		},
		{
			concern_type.Empty, concern_type.Empty,
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
	var testCase = [][]concern_type.Type{
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.YoutubeVideo, test.BibiliLive),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.YoutubeLive), concern_type.Empty.Add(test.YoutubeLive, test.BibiliLive),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.BilibiliNews),
		},
		{
			concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty.Add(test.YoutubeLive, test.YoutubeVideo),
		},
		{
			concern_type.Empty, concern_type.Empty,
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
		"bilibiliLive", "douyuLive", "bilibiliLive/bilibiliNews", "",
	}
	var expected = []concern_type.Type{
		test.BibiliLive, test.DouyuLive, concern_type.Empty.Add(test.BibiliLive, test.BilibiliNews), concern_type.Empty,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], concern_type.FromString(testCase[i]))
	}
}

func TestType_Empty(t *testing.T) {
	var testCase = []concern_type.Type{
		concern_type.Empty,
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
