package permission

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoleType_String(t *testing.T) {
	var testCase = []RoleType{
		Admin,
		TargetAdmin,
		User,
		Unknown,
	}
	var expected = []string{
		"Admin",
		"TargetAdmin",
		"User",
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
		"TargetAdmin",
		"User",
		"",
	}
	var expected = []RoleType{
		Admin,
		TargetAdmin,
		User,
		Unknown,
	}
	assert.Equal(t, len(expected), len(testCase))
	for i := range expected {
		assert.Equal(t, expected[i], NewRoleFromString(testCase[i]))
	}
}
