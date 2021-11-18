package blockCache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlockCache(t *testing.T) {
	assert.Panics(t, func() {
		NewBlockCache(0, 0)
	})
	b := NewBlockCache(10, 10)
	const (
		key1 = "a"
		key2 = "b"
	)
	var (
		v1 interface{} = new(int)
	)
	result := b.WithCacheDo(key1, func() ActionResult {
		return NewResultWrapper(v1, nil)
	})
	assert.NotNil(t, result)
	assert.EqualValues(t, v1, result.Result())
	assert.Nil(t, result.Err())

	result = b.WithCacheDo(key1, func() ActionResult {
		assert.Fail(t, "should not run")
		return nil
	})
	assert.NotNil(t, result)
	assert.EqualValues(t, v1, result.Result())
	assert.Nil(t, result.Err())

	assert.Panics(t, func() {
		b.WithCacheDo(key2, func() ActionResult {
			panic("should panic here")
		})
	})

	b2 := NewBlockCache(0, 10)
	b2.WithCacheDo("a", func() ActionResult {
		return NewResultWrapper("a", nil)
	})

	b2.WithCacheDo("a", func() ActionResult {
		assert.Fail(t, "should not run")
		return nil
	})

	b3 := NewBlockCache(10, 10)
	go b3.WithCacheDo("a", func() ActionResult {
		time.Sleep(time.Second)
		return NewResultWrapper("a", nil)
	})
	time.Sleep(time.Millisecond * 50)
	result = b3.WithCacheDo("a", func() ActionResult {
		return NewResultWrapper("b", nil)
	})
	assert.NotNil(t, result)
	assert.EqualValues(t, "a", result.Result())
	assert.Nil(t, result.Err())
}
