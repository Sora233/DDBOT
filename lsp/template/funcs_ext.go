package template

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"hash/adler32"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var funcsExt = make(FuncMap)

// RegisterExtFunc 在init阶段插入额外的template函数
func RegisterExtFunc(name string, fn interface{}) {
	checkValueFuncs(name, fn)
	funcsExt[name] = fn
}

func memberInfo(groupCode int64, uin int64) map[string]interface{} {
	var result = make(map[string]interface{})
	gi := localutils.GetBot().FindGroup(groupCode)
	if gi == nil {
		return result
	}
	fi := gi.FindMember(uin)
	if fi == nil {
		return result
	}
	result["name"] = fi.DisplayName()
	switch fi.Permission {
	case client.Owner:
		// 群主
		result["permission"] = 10
	case client.Administrator:
		// 管理员
		result["permission"] = 5
	default:
		// 其他
		result["permission"] = 1
	}
	switch fi.Gender {
	case 0:
		// 男
		result["gender"] = 2
	case 1:
		// 女
		result["gender"] = 1
	default:
		// 未知
		result["gender"] = 0
	}
	return result
}

func cut() *mmsg.CutElement {
	return new(mmsg.CutElement)
}

func prefix(commandName ...string) string {
	if len(commandName) == 0 {
		return cfg.GetCommandPrefix()
	} else {
		return cfg.GetCommandPrefix(commandName[0]) + commandName[0]
	}
}

func reply(msg interface{}) *message.ReplyElement {
	if msg == nil {
		return nil
	}
	switch e := msg.(type) {
	case *message.GroupMessage:
		return message.NewReply(e)
	case *message.PrivateMessage:
		return message.NewPrivateReply(e)
	default:
		panic(fmt.Sprintf("unknown reply message %v", msg))
	}
}

func at(uin int64) *mmsg.AtElement {
	return mmsg.NewAt(uin)
}

func pic(uri string, alternative ...string) (e *mmsg.ImageBytesElement) {
	logger := logger.WithField("uri", uri)
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		e = mmsg.NewImageByUrlWithoutCache(uri)
	} else {
		fi, err := os.Stat(uri)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Errorf("template: pic uri doesn't exist")
			} else {
				logger.Errorf("template: pic uri Stat error %v", err)
			}
			goto END
		}
		if fi.IsDir() {
			f, err := os.Open(uri)
			if err != nil {
				logger.Errorf("template: pic uri Open error %v", err)
				goto END
			}
			dirs, err := f.ReadDir(-1)
			if err != nil {
				logger.Errorf("template: pic uri ReadDir error %v", err)
				goto END
			}
			var result []os.DirEntry
			for _, file := range dirs {
				if file.IsDir() || !(strings.HasSuffix(file.Name(), ".jpg") ||
					strings.HasSuffix(file.Name(), ".png") ||
					strings.HasSuffix(file.Name(), ".gif")) {
					continue
				}
				result = append(result, file)
			}
			if len(result) > 0 {
				e = mmsg.NewImageByLocal(filepath.Join(uri, result[rand.Intn(len(result))].Name()))
			} else {
				logger.Errorf("template: pic uri can not find any images")
			}
		}
	END:
		if e == nil {
			e = mmsg.NewImageByLocal(uri)
		}
	}
	if len(alternative) > 0 && len(alternative[0]) > 0 {
		e.Alternative(alternative[0])
	}
	return e
}

func icon(uin int64, size ...uint) *mmsg.ImageBytesElement {
	var width uint = 120
	var height uint = 120
	if len(size) > 0 && size[0] > 0 {
		width = size[0]
		height = size[0]
		if len(size) > 1 && size[1] > 0 {
			height = size[1]
		}
	}
	return mmsg.NewImageByUrl(fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%v&s=640", uin)).Resize(width, height)
}

func roll(from, to int64) int64 {
	return rand.Int63n(to-from+1) + from
}

func choose(args ...reflect.Value) string {
	if len(args) == 0 {
		panic("empty choose")
	}
	var items []string
	var weights []int64
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var weight int64 = 1
		if arg.Kind() != reflect.String {
			panic("choose item must be string")
		}
		items = append(items, arg.String())
		if i+1 < len(args) {
			next := args[i+1]
			if next.Kind() != reflect.String {
				// 当作权重处理
				switch next.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					weight = next.Int()
				default:
					panic("item weight must be integer")
				}
				i++
			}
		}
		if weight <= 0 {
			panic("item weight must greater than 0")
		}
		weights = append(weights, weight)
	}

	if len(items) != len(weights) {
		logger.Errorf("Internal: items weights mismatched: %v %v", items, weights)
		panic("Internal: items weights mismatched")
	}

	var sum int64 = 0
	for _, w := range weights {
		sum += w
	}
	result := rand.Int63n(sum) + 1
	for i := 0; i < len(weights); i++ {
		result -= weights[i]
		if result <= 0 {
			return items[i]
		}
	}
	logger.Errorf("Internal: wrong rand: %v %v - %v", items, weights, result)
	panic("Internal: wrong rand")
}

func hour() int {
	return time.Now().Hour()
}

func minute() int {
	return time.Now().Minute()
}

func second() int {
	return time.Now().Second()
}

func month() int {
	return int(time.Now().Month())
}

func year() int {
	return time.Now().Year()
}

func day() int {
	return time.Now().Day()
}

func yearday() int {
	return time.Now().YearDay()
}

func weekday() int {
	// 星期天返回0，手动改成7
	t := time.Now().Weekday()
	if t == 0 {
		t = 7
	}
	return int(t)
}

func toFloat64(v interface{}) float64 {
	return cast.ToFloat64(v)
}

func toInt(v interface{}) int {
	return cast.ToInt(v)
}

func toInt64(v interface{}) int64 {
	return cast.ToInt64(v)
}

func max(a interface{}, i ...interface{}) int64 {
	return cast.ToInt64(maxf(a, i...))
}

func maxf(a interface{}, i ...interface{}) float64 {
	aa := toFloat64(a)
	for _, b := range i {
		bb := toFloat64(b)
		aa = math.Max(aa, bb)
	}
	return aa
}

func min(a interface{}, i ...interface{}) int64 {
	return cast.ToInt64(minf(a, i...))
}

func minf(a interface{}, i ...interface{}) float64 {
	aa := toFloat64(a)
	for _, b := range i {
		bb := toFloat64(b)
		aa = math.Min(aa, bb)
	}
	return aa
}

func base64encode(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func base64decode(v string) string {
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func execDecimalOp(a interface{}, b []interface{}, f func(d1, d2 decimal.Decimal) decimal.Decimal) float64 {
	prt := decimal.NewFromFloat(toFloat64(a))
	for _, x := range b {
		dx := decimal.NewFromFloat(toFloat64(x))
		prt = f(prt, dx)
	}
	rslt, _ := prt.Float64()
	return rslt
}

func sha256sum(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func sha1sum(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func adler32sum(input string) string {
	hash := adler32.Checksum([]byte(input))
	return fmt.Sprintf("%d", hash)
}

func strval(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func trunc(c int, s string) string {
	if c < 0 && len(s)+c > 0 {
		return s[len(s)+c:]
	}
	if c >= 0 && len(s) > c {
		return s[:c]
	}
	return s
}
