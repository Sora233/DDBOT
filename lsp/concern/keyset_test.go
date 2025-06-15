package concern

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestNewPrefixKeySetWithInt64ID(t *testing.T) {
	pks := NewPrefixKeySetWithInt64ID("test1")
	assert.NotNil(t, pks)
	pks.FreshKey()
	pks.GroupAtAllMarkKey()
	pks.GroupConcernConfigKey()
	g, id, err := pks.ParseGroupConcernStateKey(pks.GroupConcernStateKey(test.G1, test.UID1))
	assert.Nil(t, err)
	assert.EqualValues(t, test.G1, g)
	assert.EqualValues(t, test.UID1, id)
}

func TestNewPrefixKeySetWithStringID(t *testing.T) {
	pks := NewPrefixKeySetWithStringID("test2")
	assert.NotNil(t, pks)
	pks.FreshKey()
	pks.GroupAtAllMarkKey()
	pks.GroupConcernConfigKey()
	g, id, err := pks.ParseGroupConcernStateKey(pks.GroupConcernStateKey(test.G1, test.NAME1))
	assert.Nil(t, err)
	assert.EqualValues(t, test.G1, g)
	assert.EqualValues(t, test.NAME1, id)
}
