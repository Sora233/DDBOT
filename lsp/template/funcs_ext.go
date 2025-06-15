package template

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/shopspring/decimal"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/cfg"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/v2/utils"
)

var funcsExt = make(FuncMap)

// RegisterExtFunc 在init阶段插入额外的template函数
func RegisterExtFunc(name string, fn interface{}) {
	checkValueFuncs(name, fn)
	funcsExt[name] = fn
}

func memberInfo(groupCode uint32, uin uint32) map[string]interface{} {
	var result = make(map[string]interface{})
	gi := localutils.GetBot().FindGroup(groupCode)
	if gi == nil {
		return result
	}
	fi := localutils.GetBot().FindGroupMember(groupCode, uin)
	if fi == nil {
		return result
	}
	result["name"] = fi.DisplayName()
	switch fi.Permission {
	case entity.Owner:
		// 群主
		result["permission"] = 10
	case entity.Admin:
		// 管理员
		result["permission"] = 5
	default:
		// 其他
		result["permission"] = 1
	}
	switch fi.Sex {
	case 1:
		// 男
		result["gender"] = 2
	case 2:
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
		return message.NewGroupReply(e)
	case *message.PrivateMessage:
		return message.NewPrivateReply(e)
	default:
		panic(fmt.Sprintf("unknown reply message %v", msg))
	}
}

func at(uin uint32) *mmsg.AtElement {
	return mmsg.NewAt(uin)
}

// poke 戳一戳
func poke(uin uint32) *mmsg.PokeElement {
	return mmsg.NewPoke(uin)
}

func botUin() uint32 {
	return localutils.GetBot().GetUin()
}

func picUri(uri string) (e *mmsg.ImageBytesElement) {
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
	return e
}

func pic(input interface{}, alternative ...string) *mmsg.ImageBytesElement {
	var alt string
	if len(alternative) > 0 && len(alternative[0]) > 0 {
		alt = alternative[0]
	}
	switch e := input.(type) {
	case string:
		if b, err := base64.StdEncoding.DecodeString(e); err == nil {
			return mmsg.NewImage(b).Alternative(alt)
		}
		return picUri(e).Alternative(alt)
	case []byte:
		return mmsg.NewImage(e).Alternative(alt)
	default:
		panic(fmt.Sprintf("invalid input %v", input))
	}
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

func execDecimalOp(a interface{}, b []interface{}, f func(d1, d2 decimal.Decimal) decimal.Decimal) float64 {
	prt := decimal.NewFromFloat(toFloat64(a))
	for _, x := range b {
		dx := decimal.NewFromFloat(toFloat64(x))
		prt = f(prt, dx)
	}
	rslt, _ := prt.Float64()
	return rslt
}

func cooldown(ttlUnit string, keys ...interface{}) bool {
	ttl, err := time.ParseDuration(ttlUnit)
	if err != nil {
		panic(fmt.Sprintf("ParseDuration: can not parse <%v>: %v", ttlUnit, err))
	}
	key := localdb.NamedKey("TemplateCooldown", keys)

	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	err = localdb.Set(key, "",
		localdb.SetExpireOpt(ttl),
		localdb.SetNoOverWriteOpt(),
	)
	if err == localdb.ErrRollback {
		return false
	} else if err != nil {
		logger.Errorf("localdb.Set: cooldown set <%v> error %v", key, err)
		panic(fmt.Sprintf("INTERNAL: db error"))
	}
	return true
}

func openFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Errorf("template: openFile <%v> error %v", path, err)
		return nil
	}
	return data
}

type ddError struct {
	ddErrType string
	e         message.IMessageElement
	err       error
}

func (d *ddError) Error() string {
	if d.err != nil {
		return d.Error()
	}
	return ""
}

var errFin = &ddError{ddErrType: "fin", err: fmt.Errorf("fin")}

func abort(e ...interface{}) interface{} {
	if len(e) > 0 {
		i := e[0]
		aerr := &ddError{ddErrType: "abort", err: fmt.Errorf("abort")}
		switch s := i.(type) {
		case string:
			aerr.e = message.NewText(s)
		case message.IMessageElement:
			aerr.e = s
		default:
			panic("template: abort with invalid e")
		}
		panic(aerr)
	}
	panic(&ddError{ddErrType: "abort"})
}

func fin() interface{} {
	panic(errFin)
}
