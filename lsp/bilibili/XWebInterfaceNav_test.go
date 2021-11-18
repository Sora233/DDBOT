package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXWebInterfaceNav(t *testing.T) {
	_, err := XWebInterfaceNav()
	assert.NotNil(t, err)
}
