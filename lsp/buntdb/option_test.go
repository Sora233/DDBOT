package buntdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOption(t *testing.T) {
	var o *option
	o.getInnerExpire()
	o.getNoOverWrite()
	o.getExpire()
	o.getIgnoreNotFound()
	o.getIgnoreExpire()

	o.getTTL()
	o.setTTL(0)
	o = getOption(SetExpireOpt(time.Hour))
	assert.EqualValues(t, time.Hour, o.getExpire())
	var ttl time.Duration
	o.ttl = &ttl
	o.setTTL(-time.Second)

	var previous int64
	o.previous = &previous
	o.setPrevious("123a")
}
