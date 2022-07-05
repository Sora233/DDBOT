package template

import (
	"github.com/spf13/cast"
	"math"
)

func max(a interface{}, i ...interface{}) int64 {
	return cast.ToInt64(maxf(a, i...))
}

func maxf(a interface{}, i ...interface{}) float64 {
	aa := toFloat64(a)
	for _, b := range i {
		bb := toFloat64(b)
		aa = math.Max(aa, bb)
	}
	return aa
}

func min(a interface{}, i ...interface{}) int64 {
	return cast.ToInt64(minf(a, i...))
}

func minf(a interface{}, i ...interface{}) float64 {
	aa := toFloat64(a)
	for _, b := range i {
		bb := toFloat64(b)
		aa = math.Min(aa, bb)
	}
	return aa
}
