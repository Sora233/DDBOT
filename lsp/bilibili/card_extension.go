package bilibili

import (
	"errors"
)

var ErrCardTypeMismatch = errors.New("card type mismatch")

func (m *Card) GetCardWithImage() (*CardWithImage, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithImage {
		var card = new(CardWithImage)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithOrig() (*CardWithOrig, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithOrigin {
		var card = new(CardWithOrig)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithVideo() (*CardWithVideo, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithVideo {
		var card = new(CardWithVideo)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardTextOnly() (*CardTextOnly, error) {
	if m.GetDesc().GetType() == DynamicDescType_TextOnly {
		var card = new(CardTextOnly)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithPost() (*CardWithPost, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithPost {
		var card = new(CardWithPost)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithMusic() (*CardWithMusic, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithMusic {
		var card = new(CardWithMusic)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithSketch() (*CardWithSketch, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithSketch {
		var card = new(CardWithSketch)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithLive() (*CardWithLive, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithLive {
		var card = new(CardWithLive)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithLiveV2() (*CardWithLiveV2, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithLiveV2 {
		var card = new(CardWithLiveV2)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}

func (m *Card) GetCardWithCourse() (*CardWithCourse, error) {
	if m.GetDesc().GetType() == DynamicDescType_WithCourse {
		var card = new(CardWithCourse)
		err := json.Unmarshal([]byte(m.GetCard()), card)
		return card, err
	}
	return nil, ErrCardTypeMismatch
}
