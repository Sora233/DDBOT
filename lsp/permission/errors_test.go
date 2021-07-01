package permission

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsPermissionError(t *testing.T) {
	var testCase = []error{
		ErrDisabled,
		ErrPermissionDenied,
		ErrPermissionNotExist,
		ErrPermissionExist,
		errors.New("false"),
	}
	var expected = []bool{
		true,
		true,
		true,
		true,
		false,
	}
	assert.Equal(t, len(expected), len(testCase))
	for idx := range expected {
		assert.Equal(t, expected[idx], IsPermissionError(testCase[idx]))
	}
}
