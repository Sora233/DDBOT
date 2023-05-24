package requests

import (
	_ "embed"
	"math/rand"
)

//go:embed fake_useragent_0.2.0.json
var fakeUserAgentJson []byte

var fakeUserAgentMap = func() map[string][]string {
	var m = make(map[string][]string)
	json.Unmarshal(fakeUserAgentJson, &m)
	return m
}()

type FakeUAEntry string

const (
	Android          FakeUAEntry = "android"
	Chrome           FakeUAEntry = "chrome"
	Computer         FakeUAEntry = "computer"
	Firefox          FakeUAEntry = "firefox"
	InternetExplorer FakeUAEntry = "internet-explorer"
	Ios              FakeUAEntry = "ios"
	Ipad             FakeUAEntry = "ipad"
	Iphone           FakeUAEntry = "iphone"
	Linux            FakeUAEntry = "linux"
	MacOsX           FakeUAEntry = "mac-os-x"
	Mobile           FakeUAEntry = "mobile"
	Safari           FakeUAEntry = "safari"
)

// RandomUA entry is in ['android', 'chrome', 'computer', 'firefox', 'internet-explorer', 'ios', 'ipad', 'iphone', 'linux', 'mac-os-x', 'mobile', 'safari']
func RandomUA(entry FakeUAEntry) string {
	l := fakeUserAgentMap[string(entry)]
	count := len(l)
	if count == 0 {
		return defaultUA
	}
	return l[rand.Intn(count)]
}
