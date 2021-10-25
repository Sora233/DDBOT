package version

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/tidwall/buntdb"
)

func GetCurrentVersion(name string) int64 {
	var version int64
	err := localdb.RWCover(func() error {
		v, err := localdb.GetInt64(localdb.VersionKey(name))
		version = v
		return err
	})
	if err == buntdb.ErrNotFound {
		err = nil
	}
	if err != nil {
		version = -1
	}
	return version
}

func SetVersion(name string, version int64) (oldVersion int64, err error) {
	return localdb.SetInt64(localdb.VersionKey(name), version)
}
