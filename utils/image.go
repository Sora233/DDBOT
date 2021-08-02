package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/ericpauley/go-quantize/quantize"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"time"
)

func ImageGet(url string, prefer proxy_pool.Prefer, opt ...requests.Option) ([]byte, error) {
	if url == "" {
		return nil, errors.New("empty url")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	opts := []requests.Option{requests.ProxyOption(prefer), requests.TimeoutOption(time.Second * 5)}
	opts = append(opts, opt...)

	resp, err := requests.Get(ctx, url, nil, 3, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Content()
}

func ImageNormSize(origImage []byte) ([]byte, error) {
	dImage, format, err := image.Decode(bytes.NewReader(origImage))
	if err != nil {
		return nil, fmt.Errorf("image decode failed %v", err)
	}
	resizedImage := resize.Thumbnail(1200, 1200, dImage, resize.Lanczos3)
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
		err = fmt.Errorf("unknown format %v", format)
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
	img, err := DecodeGifWithCompleteFrame(bytes.NewReader(imgBytes))
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

func DecodeGifWithCompleteFrame(r io.Reader) (g *gif.GIF, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()
	g, err = gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}

	imgWidth, imgHeight := GetGifDimensions(g)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for index, srcImg := range g.Image {
		switch g.Disposal[index] {
		case gif.DisposalNone:
			draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.Point{}, draw.Over)
			draw.Draw(srcImg, overpaintImage.Bounds(), overpaintImage, image.Point{}, draw.Src)
		case gif.DisposalBackground:
			// TODO is it correct?
			//draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			// TODO how?
			draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.Point{}, draw.Over)
			draw.Draw(srcImg, overpaintImage.Bounds(), overpaintImage, image.Point{}, draw.Src)
		}

	}

	return g, nil
}

func GetGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

func ImageSuffix(name string) bool {
	for _, suf := range []string{"jpg", "png", "jpeg"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func MergeImages(images [][]byte) ([]byte, error) {
	if len(images) == 0 {
		return nil, errors.New("no image given")
	}

	// 1440可以被2 3 4 5 6整除
	const ColumnSize = 1440
	var columns int
	for columns = 2; columns < 6 && columns*columns < len(images); columns++ {
	}
	var subImageSize = ColumnSize / columns

	var rows = (len(images)-1)/columns + 1
	for _, img := range images {
		if len(img) == 0 {
			return nil, errors.New("empty image exists")
		}
	}
	var bg = image.NewNRGBA(image.Rect(0, 0, ColumnSize, subImageSize*rows))
	draw.Draw(bg, bg.Bounds(), image.White, image.Point{}, draw.Src)
	for i, imgBytes := range images {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(imgBytes))
		if err != nil {
			return nil, fmt.Errorf("DecodeConfig failed %v", err)
		}
		minSize := cfg.Width
		if minSize > cfg.Height {
			minSize = cfg.Height
		}
		img, _, err := image.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			return nil, fmt.Errorf("Decode failed %v", err)
		}
		if cfg.Width == cfg.Height {
			img = resize.Resize(uint(subImageSize), uint(subImageSize), img, resize.Lanczos3)
		} else {
			img = resize.Resize(uint(subImageSize), uint(subImageSize), SubImage(img, image.Rect(0, 0, minSize, minSize)), resize.Lanczos3)
		}
		si := i / columns
		sj := i % columns
		draw.Draw(bg, img.Bounds().Add(image.Point{X: sj * subImageSize, Y: si * subImageSize}), img, image.Point{}, draw.Src)
	}
	var result = new(bytes.Buffer)
	err := jpeg.Encode(result, bg, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	resBytes := result.Bytes()
	return resBytes, nil
}

func SubImage(img image.Image, r image.Rectangle) image.Image {
	var bgSize = image.Rect(0, 0, r.Dx(), r.Dy())
	var bg = image.NewNRGBA(bgSize)
	draw.Draw(bg, bg.Bounds(), img, r.Min, draw.Src)
	return bg
}
