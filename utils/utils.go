package utils

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Sora233/requests"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var ErrGoCvNotSetUp = errors.New("gocv not setup")

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if ImageSuffix(info.Name()) {
				files = append(files, path)
			}
		} else if path != root {
			subfiles, err := FilePathWalkDir(path)
			if err != nil {
				return err
			}
			for _, f := range subfiles {
				files = append(files, f)
			}
		}
		return nil
	})
	return files, err
}

func toMap(data interface{}) (map[string]string, error) {
	params := make(map[string]string)

	rg := reflect.ValueOf(data)
	if rg.Type().Kind() == reflect.Ptr {
		rg = rg.Elem()
	}
	if rg.Type().Kind() != reflect.Struct {
		return nil, errors.New("can only convert struct type")
	}
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
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			params[fillname] = strconv.FormatInt(rg.Field(i).Int(), 10)
		case reflect.String:
			params[fillname] = rg.Field(i).String()
		case reflect.Bool:
			params[fillname] = strconv.FormatBool(rg.Field(i).Bool())
		default:
			return nil, fmt.Errorf("not support type %v", field.Type.Kind().String())
		}

	}
	return params, nil
}

func ToParams(get interface{}) (requests.Params, error) {
	return toMap(get)
}

func ToDatas(data interface{}) (requests.Datas, error) {
	return toMap(data)
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

func PrefixMatch(opts []string, target string) (string, bool) {
	if len(opts) == 0 {
		return "", false
	}
	var (
		found  = false
		result = ""
	)
	for _, opt := range opts {
		if strings.HasPrefix(opt, target) {
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
	if bot.Instance == nil {
		return fields
	}
	if groupInfo := bot.Instance.FindGroup(groupCode); groupInfo != nil {
		fields["GroupName"] = groupInfo.Name
	}
	return fields
}
