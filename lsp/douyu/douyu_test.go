package douyu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDouyuPath(t *testing.T) {
	assert.NotEmpty(t, DouyuPath("/asd"))
}
