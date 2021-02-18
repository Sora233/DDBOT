package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/requests"
	"github.com/ericpauley/go-quantize/quantize"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
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

func ImageSuffix(name string) bool {
	for _, suf := range []string{"jpg", "png"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func ToParams(get interface{}) (requests.Params, error) {
	params := make(requests.Params)

	rg := reflect.ValueOf(get)
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

func GuessImageFormat(img []byte) (format string, err error) {
	r := bytes.NewReader(img)
	_, format, err = image.DecodeConfig(r)
	return format, err
}

func FuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}

func ImageGet(url string, prefer proxy_pool.Prefer) ([]byte, error) {
	if url == "" {
		return nil, errors.New("empty url")
	}
	req := requests.Requests()
	proxy, err := proxy_pool.Get(prefer)
	if err == nil {
		req.Proxy(proxy.ProxyString())
	}
	resp, err := req.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Content(), nil
}

func ImageNormSize(origImage []byte) ([]byte, error) {
	dImage, format, err := image.Decode(bytes.NewReader(origImage))
	if err != nil {
		return nil, fmt.Errorf("image decode failed %v", err)
	}
	resizedImage := resize.Thumbnail(1280, 860, dImage, resize.Lanczos3)
	resizedImageBuffer := bytes.NewBuffer(make([]byte, 0))
	switch format {
	case "jpeg":
		err = jpeg.Encode(resizedImageBuffer, resizedImage, &jpeg.Options{Quality: 100})
	case "gif":
		err = gif.Encode(resizedImageBuffer, resizedImage, &gif.Options{
			Quantizer: quantize.MedianCutQuantizer{},
		})
	case "png":
		err = png.Encode(resizedImageBuffer, resizedImage)
	default:
		err = errors.New("unknown format")
	}
	return resizedImageBuffer.Bytes(), err
}

func ImageGetAndNorm(url string, prefer proxy_pool.Prefer) ([]byte, error) {
	img, err := ImageGet(url, prefer)
	if err != nil {
		return img, err
	}
	img, err = ImageNormSize(img)
	return img, err
}

func ImageFormat(origImage []byte) (string, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(origImage))
	return format, err
}

func ImageReserve(imgBytes []byte) ([]byte, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", FuncName()).Tracef("cost %v", ed.Sub(st))
	}()

	format, err := ImageFormat(imgBytes)
	if err != nil {
		return nil, err
	} else if format != "gif" {
		return nil, errors.New("不是动图")
	}
	img, err := DecodeGifWithCompleteFrame(imgBytes)
	if err != nil {
		return nil, err
	}
	length := len(img.Image)
	for idx := range img.Image {
		oidx := length - 1 - idx
		if idx >= oidx {
			break
		}
		tmp := img.Image[idx]
		img.Image[idx] = img.Image[oidx]
		img.Image[oidx] = tmp
	}
	var result = bytes.NewBuffer(nil)
	err = gif.EncodeAll(result, img)
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func DecodeGifWithCompleteFrame(imgBytes []byte) (*gif.GIF, error) {
	img, err := gif.DecodeAll(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, err
	}
	if len(img.Image) == 0 {
		return nil, errors.New("unknown image without frame")
	}
	var baseImg = image.NewPaletted(img.Image[0].Bounds(), img.Image[0].Palette)
	for index, src := range img.Image {
		draw.Draw(baseImg, src.Bounds(), src, src.Rect.Min, draw.Over)
		draw.Draw(img.Image[index], baseImg.Bounds(), baseImg, baseImg.Rect.Min, draw.Over)
	}
	return img, nil
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

func MessageFilter(msg []message.IMessageElement, filter func(message.IMessageElement) bool) []message.IMessageElement {
	var result []message.IMessageElement
	for _, e := range msg {
		if filter(e) {
			result = append(result, e)
		}
	}
	return result
}

func MessageTextf(format string, args ...interface{}) *message.TextElement {
	return message.NewText(fmt.Sprintf(format, args...))
}

func UnquoteString(s string) (string, error) {
	return strconv.Unquote(fmt.Sprintf(`"%s"`, strings.Trim(s, `"`)))
}

func TimestampFormat(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05")
}
