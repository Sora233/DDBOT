package utils

import (
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/guonaihong/gout"
	"github.com/stretchr/testify/assert"
)

func TestFilePathWalkDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "ddbot-test")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)
	err = os.MkdirAll(filepath.Join(tempDir, "a", "b"), 0755)
	assert.Nil(t, err)

	touch := func(name string) {
		f, err := os.Create(name)
		assert.Nil(t, err)
		f.Close()
	}

	touch(filepath.Join(tempDir, "1.jpg"))
	touch(filepath.Join(tempDir, "2.jpg"))
	touch(filepath.Join(tempDir, "a", "3.jpg"))
	touch(filepath.Join(tempDir, "a", "b", "4.jpg"))

	result, err := FilePathWalkDir(tempDir)
	assert.Nil(t, err)
	assert.Len(t, result, 4)

	assert.Contains(t, result, filepath.Join(tempDir, "1.jpg"))
	assert.Contains(t, result, filepath.Join(tempDir, "2.jpg"))
	assert.Contains(t, result, filepath.Join(tempDir, "a", "3.jpg"))
	assert.Contains(t, result, filepath.Join(tempDir, "a", "b", "4.jpg"))
}

func TestArgSplit(t *testing.T) {
	result := ArgSplit(`-a "q w e" -b c`)
	assert.Len(t, result, 4)
	assert.EqualValues(t, []string{"-a", "q w e", "-b", "c"}, result)
}

func TestSwitch2Bool(t *testing.T) {
	assert.True(t, Switch2Bool("on"))
	assert.False(t, Switch2Bool("off"))
}

func TestJoinInt64(t *testing.T) {
	result := JoinInt64([]int64{1, 10, 99, 100}, ",")
	assert.EqualValues(t, "1,10,99,100", result)

	result = JoinInt64([]int64{15, 20, 25, 30}, "-")
	assert.EqualValues(t, "15-20-25-30", result)

	result = JoinInt64(nil, "-")
	assert.Empty(t, result)
	result = JoinInt64([]int64{1}, "-")
	assert.EqualValues(t, "1", result)
}

func TestRetry(t *testing.T) {
	var i = 0
	Retry(10, time.Millisecond*50, func() bool {
		i++
		return i == 2
	})

	assert.EqualValues(t, 2, i)
	i = 0

	Retry(1, time.Millisecond*50, func() bool {
		i++
		return i == 2
	})
	assert.EqualValues(t, 1, i)
}

func TestPrefixMatch(t *testing.T) {
	result, found := PrefixMatch([]string{"aa", "ab", "ac", "bb", "cc", "dd"}, "q")
	assert.False(t, found)

	result, found = PrefixMatch(nil, "q")
	assert.False(t, found)

	result, found = PrefixMatch([]string{"aa", "ab", "ac", "bb", "cc", "dd"}, "a")
	assert.False(t, found)

	result, found = PrefixMatch([]string{"aa", "ab", "ac", "bb", "cc", "dd"}, "c")
	assert.True(t, found)
	assert.Equal(t, "cc", result)
}

func TestToDatas(t *testing.T) {
	var m1 = map[string]string{
		"a": "b",
		"c": "d",
	}
	result, err := ToDatas(m1)
	assert.Nil(t, err)
	assert.EqualValues(t, m1, result)

	var s1 = struct {
		A string `json:"a,omitempty"`
		B int    `json:"b"`
		C uint   `json:"-"`
	}{
		A: "aaa",
		B: 10,
		C: 99,
	}
	result, err = ToDatas(&s1)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		"a": "aaa",
		"b": "10",
	}, result)

	var m2 = map[string]interface{}{
		"a": 1,
		"b": 99,
		"c": "123",
	}
	result, err = ToDatas(&m2)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		"a": "1",
		"b": "99",
		"c": "123",
	}, result)

	var s2 = struct {
		A string
		B int
	}{
		A: "a",
		B: 10,
	}
	result, err = ToDatas(&s2)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		"a": "a",
		"b": "10",
	}, result)
}

func TestToParams(t *testing.T) {
	var m1 = gout.H{
		"a": "b",
		"c": "d",
	}
	result, err := ToParams(m1)
	assert.Nil(t, err)
	assert.EqualValues(t, m1, result)

	var m2 = map[string]string{
		"a": "b",
		"c": "d",
	}
	result, err = ToParams(&m2)
	assert.Nil(t, err)
	assert.EqualValues(t, gout.H{"a": "b", "c": "d"}, result)

	var m3 = map[string]interface{}{
		"a": "b",
		"c": 100,
	}
	result, err = ToParams(&m3)
	assert.Nil(t, err)
	assert.EqualValues(t, gout.H{"a": "b", "c": "100"}, result)

	var s1 = struct {
		A string `json:"a,omitempty"`
		B int64
		C string `json:"-"`
		D bool
		E uint `json:"e"`
	}{
		A: "aa",
		E: 114514,
		B: 1919810,
		C: "cccc",
		D: true,
	}
	result, err = ToParams(&s1)
	assert.Nil(t, err)
	assert.EqualValues(t, gout.H{"a": "aa", "b": "1919810", "d": "true", "e": "114514"}, result)
}

func TestUrlEncode(t *testing.T) {
	var m1 = map[string]string{
		"a": "b",
		"c": "d",
	}
	result := UrlEncode(m1)
	assert.EqualValues(t, "a=b&c=d", result)
}

func TestReflectToString(t *testing.T) {
	var a interface{}
	reflectToString(reflect.ValueOf(a))
}

func TestToCamel(t *testing.T) {
	assert.EqualValues(t, "", toCamel(""))
	assert.EqualValues(t, "boy_loves_girl", toCamel("BoyLovesGirl"))
	assert.EqualValues(t, "dog_god", toCamel("DogGod"))
}

func TestRemoveHtmlTag(t *testing.T) {
	var testCase = []string{
		"</a>",
		"aaa??qweqwe",
		"qabcd<a></a>www",
		"qwe<html></html>",
		html.EscapeString("qwe<html></html>"),
	}
	var expected = []string{
		"",
		"aaa??qweqwe",
		"qabcdwww",
		"qwe",
		html.EscapeString("qwe<html></html>"),
	}
	assert.EqualValues(t, len(expected), len(testCase))
	for idx := range testCase {
		assert.EqualValues(t, expected[idx], RemoveHtmlTag(testCase[idx]))
	}
}

func TestGroupLogFields(t *testing.T) {
	assert.NotNil(t, GroupLogFields(test.G1))
}

func TestFriendLogFields(t *testing.T) {
	assert.NotNil(t, FriendLogFields(1))
}

func TestTimestampFormat(t *testing.T) {
	config.GlobalConfig.Viper.Set("timezone", "Asia/Shanghai")
	assert.EqualValues(t, "1970-01-01 08:00:00", TimestampFormat(0))
	config.GlobalConfig.Viper.Set("timezone", "America/New York")
	assert.EqualValues(t, "1969-12-31 19:00:00", TimestampFormat(0))

	localNow := time.Now()
	expectedStr := localNow.Format("2006-01-02 15:04:05")
	config.GlobalConfig.Viper.Set("timezone", "nowhere")
	assert.EqualValues(t, expectedStr, TimestampFormat(localNow.Unix()))
	config.GlobalConfig.Viper.Set("timezone", "")
	assert.EqualValues(t, expectedStr, TimestampFormat(localNow.Unix()))
}
