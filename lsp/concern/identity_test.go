package concern

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestNewIdentity(t *testing.T) {
	a := NewIdentity(test.ID1, test.NAME1)
	assert.EqualValues(t, test.ID1, a.GetUid())
	assert.EqualValues(t, test.NAME1, a.GetName())
}
