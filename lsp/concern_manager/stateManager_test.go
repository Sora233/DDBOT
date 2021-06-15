package concern_manager

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernConfig_Compact(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: []*AtAll{
			{
				Id:    "1",
				Ctype: 1,
			},
			{
				Id:    "1",
				Ctype: 1,
			},
			{
				Id:    "2",
				Ctype: 1,
			},
			{
				Id:    "1",
				Ctype: 1,
			},
			{
				Id:    "1",
				Ctype: 2,
			},
		},
		AtSomeone: []*AtSomeone{
			{
				Id:     "1",
				Ctype:  1,
				AtList: nil,
			},
			{
				Id:     "1",
				Ctype:  1,
				AtList: nil,
			},
			{
				Id:     "2",
				Ctype:  1,
				AtList: nil,
			},
			{
				Id:     "2",
				Ctype:  2,
				AtList: nil,
			},
			{
				Id:     "1",
				Ctype:  1,
				AtList: nil,
			},
		},
	}
	var expected = &GroupConcernAtConfig{
		AtAll: []*AtAll{
			{
				Id:    "1",
				Ctype: 1,
			},
			{
				Id:    "2",
				Ctype: 1,
			},
			{
				Id:    "1",
				Ctype: 2,
			},
		},
		AtSomeone: []*AtSomeone{
			{
				Id:     "1",
				Ctype:  1,
				AtList: nil,
			},
			{
				Id:     "2",
				Ctype:  1,
				AtList: nil,
			},
			{
				Id:     "2",
				Ctype:  2,
				AtList: nil,
			},
		},
	}
	g.Compact()

	assert.EqualValues(t, expected, g)
}

func TestGroupConcernAtConfig_CheckAtAll(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: []*AtAll{
			{
				Id:    "1",
				Ctype: 1,
			},
			{
				Id:    "1",
				Ctype: 2,
			},
		},
	}
	assert.True(t, g.CheckAtAll(int64(1), concern.Type(2)))
	assert.False(t, g.CheckAtAll(int64(1), concern.Type(3)))
}
