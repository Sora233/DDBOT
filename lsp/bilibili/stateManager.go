package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (c *StateManager) AddUserInfo(userInfo *UserInfo) error {
	if userInfo == nil {
		return errors.New("nil UserInfo")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.UserInfoKey(userInfo.Mid)
		_, _, err := tx.Set(key, userInfo.ToString(), nil)
		return err
	})
}

func (c *StateManager) GetUserInfo(mid int64) (*UserInfo, error) {
	var userInfo = &UserInfo{}

	err := c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.UserInfoKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), userInfo)
		if err != nil {
			fmt.Println(val)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (c *StateManager) AddLiveInfo(liveInfo *LiveInfo) error {
	if liveInfo == nil {
		return errors.New("nil LiveInfo")
	}
	err := c.RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentLiveKey(liveInfo.Mid), liveInfo.ToString(), localdb.ExpireOption(time.Hour*24))
		return err
	})
	return err
}

func (c *StateManager) GetLiveInfo(mid int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}

	err := c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
			fmt.Println(val)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (c *StateManager) AddNewsInfo(newsInfo *NewsInfo) error {
	if newsInfo == nil {
		return errors.New("nil NewsInfo")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentNewsKey(newsInfo.Mid), newsInfo.ToString(), localdb.ExpireOption(time.Hour*24))
		return err
	})
}

func (c *StateManager) DeleteNewsInfo(newsInfo *NewsInfo) error {
	if newsInfo == nil {
		return errors.New("nil NewsInfo")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(c.CurrentNewsKey(newsInfo.Mid))
		return err

	})
}

func (c *StateManager) GetNewsInfo(mid int64) (*NewsInfo, error) {
	var newsInfo = &NewsInfo{}

	err := c.RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentNewsKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), newsInfo)
		if err != nil {
			fmt.Println(val)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return newsInfo, nil
}

func (c *StateManager) Start() error {
	db := localdb.MustGetClient()
	db.CreateIndex(c.GroupConcernStateKey(), c.GroupConcernStateKey("*"), buntdb.IndexString)
	db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
	db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
	db.CreateIndex(c.UserInfoKey(), c.UserInfoKey("*"), buntdb.IndexString)
	db.CreateIndex(c.ConcernStateKey(), c.ConcernStateKey("*"), buntdb.IndexBinary)
	return c.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet())
	return sm
}
