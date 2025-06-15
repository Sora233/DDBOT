package bilibili

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestRelationModify(t *testing.T) {
	_, err := RelationModify(test.UID1, ActSub)
	assert.NotNil(t, err)
}
