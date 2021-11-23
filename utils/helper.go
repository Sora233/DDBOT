package utils

import (
	"fmt"
	"github.com/guonaihong/gout"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"io/fs"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func reflectToString(v reflect.Value) (string, error) {
	if !v.IsValid() || v.IsZero() {
		return "", nil
	}
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return "", fmt.Errorf("not support type %v", v.Type().Kind().String())
	}
}

func ToDatas(data interface{}) (map[string]string, error) {
	params := make(map[string]string)

	if m, ok := data.(map[string]string); ok {
		return m, nil
	}

	rg := reflect.ValueOf(data)
	for rg.Kind() == reflect.Ptr || rg.Kind() == reflect.Interface {
		rg = rg.Elem()
	}
	if rg.Kind() == reflect.Map {
		iter := rg.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()
			k1, err := reflectToString(key)
			if err != nil {
				return nil, err
			}
			v1, err := reflectToString(value)
			if err != nil {
				return nil, err
			}
			params[k1] = v1
		}
	} else if rg.Kind() == reflect.Struct {
		for i := 0; ; i++ {
			if i >= rg.Type().NumField() {
				break
			}
			field := rg.Type().Field(i)
			fillname, found := field.Tag.Lookup("json")
			if !found {
				fillname = toCamel(field.Name)
			} else {
				if pos := strings.Index(fillname, ","); pos != -1 {
					fillname = fillname[:pos]
				}
			}
			if fillname == "-" {
				continue
			}
			s, err := reflectToString(rg.Field(i))
			if err != nil {
				return nil, err
			}
			params[fillname] = s
		}
	}
	return params, nil
}

func ToParams(data interface{}) (gout.H, error) {
	if p, ok := data.(gout.H); ok {
		return p, nil
	}
	params := make(gout.H)
	m, err := ToDatas(data)
	if err != nil {
		return nil, err
	}
	for k, v := range m {
		params[k] = v
	}
	return params, nil
}

func UrlEncode(data map[string]string) string {
	params := url.Values{}
	for k, v := range data {
		params.Add(k, v)
	}
	return params.Encode()
}

func toCamel(name string) string {
	if len(name) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString(strings.ToLower(name[:1]))
	for _, c := range name[1:] {
		if c >= 'A' && c <= 'Z' {
			sb.WriteRune('_')
			sb.WriteRune(c - 'A' + 'a')
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func FuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

// PrefixMatch 从 opts 中选择一个前缀是 prefix 的字符串，如果有多个选项，则返回 false
func PrefixMatch(opts []string, prefix string) (string, bool) {
	if len(opts) == 0 {
		return "", false
	}
	var (
		found  = false
		result = ""
	)
	for _, opt := range opts {
		if strings.HasPrefix(opt, prefix) {
			if found == true {
				return "", false
			}
			found = true
			result = opt
		}
	}
	return result, found
}

func UnquoteString(s string) (string, error) {
	return strconv.Unquote(fmt.Sprintf(`"%s"`, strings.Trim(s, `"`)))
}

func TimestampFormat(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05")
}

func Retry(count int, interval time.Duration, f func() bool) bool {
	for retry := 0; retry < count; retry++ {
		if f() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

func ArgSplit(str string) (result []string) {
	r := regexp.MustCompile(`[^\s"]+|"([^"]*)"`)
	match := r.FindAllString(str, -1)
	for _, s := range match {
		result = append(result, strings.Trim(strings.TrimSpace(s), `" `))
	}
	return
}

func GroupLogFields(groupCode int64) logrus.Fields {
	var fields = make(logrus.Fields)
	fields["GroupCode"] = groupCode
	if groupInfo := GetBot().FindGroup(groupCode); groupInfo != nil {
		fields["GroupName"] = groupInfo.Name
	}
	return fields
}

func FriendLogFields(uin int64) logrus.Fields {
	var fields = make(logrus.Fields)
	fields["FriendUin"] = uin
	if info := GetBot().FindFriend(uin); info != nil {
		fields["FriendName"] = info.Nickname
	}
	return fields
}

func Switch2Bool(s string) bool {
	return s == "on"
}

func JoinInt64(ele []int64, sep string) string {
	var s []string
	for _, e := range ele {
		s = append(s, strconv.FormatInt(e, 10))
	}
	return strings.Join(s, sep)
}

var reHtmlTag = regexp.MustCompile(`<[^>]+>`)

func RemoveHtmlTag(s string) string {
	return reHtmlTag.ReplaceAllString(s, "")
}
