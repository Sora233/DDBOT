package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/asmcos/requests"
	"github.com/nfnt/resize"
	"gocv.io/x/gocv"
	"image"
	"image/color"
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
)

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

func ImageGet(url string) ([]byte, error) {
	if url == "" {
		return nil, errors.New("empty url")
	}
	req := requests.Requests()
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
		err = jpeg.Encode(resizedImageBuffer, resizedImage, nil)
	case "gif":
		err = gif.Encode(resizedImageBuffer, resizedImage, nil)
	case "png":
		err = png.Encode(resizedImageBuffer, resizedImage)
	default:
		err = errors.New("unknown format")
	}
	return resizedImageBuffer.Bytes(), err
}

func ImageGetAndNorm(url string) ([]byte, error) {
	img, err := ImageGet(url)
	if err != nil {
		return img, err
	}
	img, err = ImageNormSize(img)
	return img, err
}

func ImageFormat(origImage []byte) (string, error) {
	_, format, err := image.Decode(bytes.NewReader(origImage))
	return format, err
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

func OpenCvAnimeFaceDetect(imgBytes []byte) ([]byte, error) {
	cascade := gocv.NewCascadeClassifier()
	defer cascade.Close()
	if ok := cascade.Load("lbpcascade_animeface.xml"); !ok {
		panic(errors.New("not ok"))
	}

	format, err := ImageFormat(imgBytes)
	if err == nil && format == "gif" {
		// TODO: gif not work properly for now
		g, err := gif.DecodeAll(bytes.NewReader(imgBytes))
		if err != nil {
			return nil, err
		}
		for index := range g.Image {
			frame := g.Image[index]
			var frameByte = new(bytes.Buffer)
			err := jpeg.Encode(frameByte, frame, nil)
			if err != nil {
				continue
			}
			img, err := gocv.IMDecode(frameByte.Bytes(), gocv.IMReadColor)
			if err != nil || img.Empty() {
				continue
			}
			defer img.Close()
			rec := cascade.DetectMultiScale(img)
			for _, r := range rec {
				gocv.Rectangle(&img, r, color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 0,
				}, 2)
			}
			markedImgBytes, err := gocv.IMEncode(gocv.JPEGFileExt, img)
			if err != nil {
				continue
			}
			markedImg, err := jpeg.Decode(bytes.NewReader(markedImgBytes))
			if err != nil {
				continue
			}
			palettedImage := image.NewPaletted(frame.Bounds(), frame.Palette)
			draw.Draw(palettedImage, frame.Bounds(), markedImg, image.Point{}, draw.Src)
			g.Image[index] = palettedImage
		}
		var result = new(bytes.Buffer)
		err = gif.EncodeAll(result, g)
		if err != nil {
			return nil, err
		}
		return result.Bytes(), nil
	}

	img, err := gocv.IMDecode(imgBytes, gocv.IMReadColor)
	if err != nil {
		return nil, err
	}
	if img.Empty() {
		return nil, errors.New("不支持的格式")
	}
	defer img.Close()

	rec := cascade.DetectMultiScale(img)

	logger.WithField("method", "OpenCvAnimeFaceDetect").
		WithField("face_count", len(rec)).
		Debug("face detect")

	for _, r := range rec {
		gocv.Rectangle(&img, r, color.RGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 0,
		}, 2)
	}
	return gocv.IMEncode(gocv.JPEGFileExt, img)
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
