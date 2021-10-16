//go:build nocv
// +build nocv

package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenCvAnimeFaceDetect(t *testing.T) {
	_, err := OpenCvAnimeFaceDetect(nil)
	assert.NotNil(t, err)
}
