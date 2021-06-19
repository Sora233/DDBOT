package test

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

// 这个包只允许在单元测试中使用

const (
	G1   int64 = 123456
	G2   int64 = 654321
	UID  int64 = 777
	UID2 int64 = 888
)

func InitBuntdb(t *testing.T) {
	assert.Nil(t, localdb.InitBuntDB(localdb.MEMORYDB))
}
func CloseBuntdb(t *testing.T) {
	assert.Nil(t, localdb.Close())
}
