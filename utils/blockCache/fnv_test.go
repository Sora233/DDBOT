package blockCache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSum32(t *testing.T) {
	s := new(sum32)
	_, err := s.Write([]byte("qwertyuio"))
	assert.Nil(t, err)
	s.Sum32()
	s.Reset()
	s.Size()
}
