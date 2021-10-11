package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getCard(t DynamicDescType) *Card {
	return &Card{
		Card: "{}",
		Desc: &Card_Desc{
			Type: t,
		},
	}
}

func TestCard_GetCard(t *testing.T) {
	_, err := getCard(DynamicDescType_WithImage).GetCardWithImage()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithOrigin).GetCardWithOrig()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithVideo).GetCardWithVideo()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_TextOnly).GetCardTextOnly()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithPost).GetCardWithPost()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithMusic).GetCardWithMusic()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithSketch).GetCardWithSketch()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithLive).GetCardWithLive()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithLiveV2).GetCardWithLiveV2()
	assert.Nil(t, err)
	_, err = getCard(DynamicDescType_WithCourse).GetCardWithCourse()
	assert.Nil(t, err)
}
