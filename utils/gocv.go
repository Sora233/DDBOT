// +build !nocv

package utils

import (
	"bytes"
	"errors"
	"github.com/ericpauley/go-quantize/quantize"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
)

func OpenCvAnimeFaceDetect(imgBytes []byte) ([]byte, error) {
	animeCascade := NewAnimeCascadeClassifier()
	faceCascade := NewFaceCascadeClassifier()
	defer animeCascade.Close()
	defer faceCascade.Close()

	format, err := ImageFormat(imgBytes)
	if err == nil && format == "gif" {
		g, err := DecodeGifWithCompleteFrame(imgBytes)
		if err != nil {
			return nil, err
		}
		for index := range g.Image {
			frame := g.Image[index]
			var frameByte = new(bytes.Buffer)
			err := png.Encode(frameByte, frame)
			if err != nil {
				continue
			}
			img, err := gocv.IMDecode(frameByte.Bytes(), gocv.IMReadUnchanged)
			if err != nil {
				continue
			}
			defer img.Close()
			if img.Empty() {
				continue
			}
			var rec []image.Rectangle
			rec = animeCascade.DetectMultiScale(img)
			if len(rec) == 0 {
				rec = faceCascade.DetectMultiScale(img)
			}
			if len(rec) == 0 {
				continue
			}
			for _, r := range rec {
				gocv.Rectangle(&img, r, color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 255,
				}, 2)
			}
			markedImgBytes, err := gocv.IMEncode(gocv.PNGFileExt, img)
			if err != nil {
				continue
			}
			markedImg, err := png.Decode(bytes.NewReader(markedImgBytes))
			if err != nil {
				continue
			}
			palettedImage := image.NewPaletted(frame.Bounds(), quantize.MedianCutQuantizer{}.Quantize(make([]color.Color, 0, 256), markedImg))
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
	defer img.Close()
	if img.Empty() {
		return nil, errors.New("不支持的格式")
	}

	var rec []image.Rectangle
	rec = animeCascade.DetectMultiScale(img)
	if len(rec) == 0 {
		rec = faceCascade.DetectMultiScale(img)
	}
	logger.WithField("method", "OpenCvAnimeFaceDetect").
		WithField("face_count", len(rec)).
		Debug("face detect")

	for _, r := range rec {
		gocv.Rectangle(&img, r, color.RGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 255,
		}, 2)
	}
	return gocv.IMEncode(gocv.PNGFileExt, img)
}

func NewAnimeCascadeClassifier() gocv.CascadeClassifier {
	cascade := gocv.NewCascadeClassifier()
	if !cascade.Load("lbpcascade_animeface.xml") {
		logger.Errorf("load lbpcascade_animeface.xml failed")
	}
	return cascade
}
func NewFaceCascadeClassifier() gocv.CascadeClassifier {
	cascade := gocv.NewCascadeClassifier()
	if !cascade.Load("haarcascade_frontalface_default.xml") {
		logger.Errorf("load haarcascade_frontalface_default.xml failed")
	}
	return cascade
}
