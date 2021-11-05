package utils

import (
	"bytes"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/stretchr/testify/assert"
	"image"
	"testing"
)

var imageUrl = test.FakeImage(1500)

func TestImageNormSize(t *testing.T) {
	test.FakeImage(0)
	b, err := ImageGet(imageUrl, proxy_pool.PreferAny)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 1500, cfg.Width)
	assert.EqualValues(t, 1500, cfg.Height)

	b, err = ImageNormSize(b)
	assert.Nil(t, err)
	cfg, _, err = image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 1200, cfg.Width)
	assert.EqualValues(t, 1200, cfg.Height)
}

func TestImageGetAndNorm(t *testing.T) {
	b, err := ImageGetAndNorm(imageUrl, proxy_pool.PreferAny)
	assert.Nil(t, err)
	assert.NotEmpty(t, b)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	assert.Nil(t, err)
	assert.EqualValues(t, 1200, cfg.Width)
	assert.EqualValues(t, 1200, cfg.Height)

	format, err := ImageFormat(b)
	assert.Nil(t, err)
	assert.EqualValues(t, "png", format)

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
