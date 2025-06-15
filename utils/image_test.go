package utils

import (
	"bytes"
	"image"
	_ "image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

var imageUrl = "https://user-images.githubusercontent.com/11474360/111737379-78fbe200-88ba-11eb-9e7e-ecc9f2440dd8.jpg"

func TestImageNormSize(t *testing.T) {
	test.FakeImage(0)
	b, err := ImageGet(imageUrl)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 1080, cfg.Width)
	assert.EqualValues(t, 2340, cfg.Height)

	b, err = ImageNormSize(b)
	assert.Nil(t, err)
	cfg, _, err = image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 553, cfg.Width)
	assert.EqualValues(t, 1200, cfg.Height)
}

func TestImageGet(t *testing.T) {
	_, err := ImageGetWithoutCache(imageUrl)
	assert.Nil(t, err)
	b, err := ImageGet(imageUrl)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	b, err = ImageNormSize(b)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 553, cfg.Width)
	assert.EqualValues(t, 1200, cfg.Height)

	b, err = ImageResize(b, 500, 500)
	assert.Nil(t, err)
	cfg, _, err = image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 500, cfg.Width)
	assert.EqualValues(t, 500, cfg.Height)

	format, err := ImageFormat(b)
	assert.Nil(t, err)
	assert.EqualValues(t, "jpeg", format)

	_, err = ImageReserve(b)
	assert.NotNil(t, err)

	img, _, err := image.Decode(bytes.NewReader(b))
	assert.Nil(t, err)
	subImg := SubImage(img, image.Rect(0, 0, 10, 10))
	assert.NotNil(t, subImg)
}

func TestImageSuffix(t *testing.T) {
	assert.True(t, ImageSuffix("png"))
	assert.True(t, ImageSuffix("jpg"))
	assert.False(t, ImageSuffix("tar"))
}
