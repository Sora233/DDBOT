package version

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoMigration(t *testing.T) {
	assert.Panics(t, func() {
		DoMigration()
	})
}
