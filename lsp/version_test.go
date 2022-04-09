package lsp

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompareVersion(t *testing.T) {
	assert.True(t, compareVersion("v0.0.1", "v0.0.2"))
	assert.False(t, compareVersion("v0.0.2", "v0.0.2"))
	assert.False(t, compareVersion("v0.0.3", "v0.0.2"))
	assert.True(t, compareVersion("v0.0.1", "v0.0.18"))
	assert.True(t, compareVersion("v0.0.1", "v1.1.8"))
	assert.False(t, compareVersion("v1.1.9", "v1.1.8"))
	assert.False(t, compareVersion("v1.0.0", "v0.0.8"))
	assert.True(t, compareVersion("v0.0.1", "v0.0.10"))
	assert.False(t, compareVersion("v10.0.10", "v1.0.10"))
	assert.True(t, compareVersion("v1", "v2"))
	assert.True(t, compareVersion("v1", "v2.0"))
	assert.True(t, compareVersion("v1", "v2.0.0"))
	assert.True(t, compareVersion("v1.0", "v2.0.0"))
	assert.False(t, compareVersion("v1.0.0", "v1.0"))
	assert.False(t, compareVersion("v1.0.0", "v1"))
	assert.True(t, compareVersion("v1.0.0", "v2"))
	assert.True(t, compareVersion("v1.0.0", "v2.0"))
	assert.True(t, compareVersion("v1.0.0", "v2.0.0"))
	assert.True(t, compareVersion("v0.0.19", "v0.0.19.1"))
	assert.True(t, compareVersion("v0.0.19.1", "v0.0.20"))
}

func TestCheckUpdate(t *testing.T) {
	var w = new(bytes.Buffer)
	logrus.SetOutput(w)
	Tags = "v0.0.1"
	CheckUpdate()
	fmt.Println(w.String())
	assert.Contains(t, w.String(), "更新检测完成")
}
