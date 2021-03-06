package permission

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoleType_String(t *testing.T) {
	var testCase = []RoleType{
		Admin,
		GroupAdmin,
		Unknown,
	}
	var expected = []string{
		"Admin",
		"GroupAdmin",
		"",
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := range expected {
		assert.Equal(t, expected[i], testCase[i].String())
	}
}

func TestFromString(t *testing.T) {
	var testCase = []string{
		"Admin",
		"GroupAdmin",
		"",
	}
	var expected = []RoleType{
		Admin,
		GroupAdmin,
		Unknown,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := range expected {
		assert.Equal(t, expected[i], NewRoleFromString(testCase[i]))
	}
}
