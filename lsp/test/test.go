package test

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

// 这个包只允许在单元测试中使用

const (
	ROOMID1    int64 = 1
	ROOMID2    int64 = 2
	UID1       int64 = 777
	UID2       int64 = 888
	DynamicID1 int64 = 1001
	DynamicID2 int64 = 1002
	MessageID1 int32 = 5001
	MessageID2 int32 = 5002
	G1         int64 = 123456
	G2         int64 = 654321
	TIMESTAMP1 int64 = 1624126814
	TIMESTAMP2 int64 = 1624126914

	NAME1 = "name1"
	NAME2 = "name2"

	CMD1 = "command1"
	CMD2 = "command2"

	BVID1 = "bvid1"

	ID1 = 2001
	ID2 = 2002

	VersionName = "testVersion"
)

func InitBuntdb(t *testing.T) {
	assert.Nil(t, localdb.InitBuntDB(localdb.MEMORYDB))
}
func CloseBuntdb(t *testing.T) {
	assert.Nil(t, localdb.Close())
}
