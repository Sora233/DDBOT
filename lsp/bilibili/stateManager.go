package bilibili

import (
	"encoding/json"
	"errors"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (c *StateManager) AddUserInfo(userInfo *UserInfo) error {
	if userInfo == nil {
		return errors.New("nil UserInfo")
	}
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := c.UserInfoKey(userInfo.Mid)
		_, _, err = tx.Set(key, userInfo.ToString(), nil)
		return err
	})
	return err
}

func (c *StateManager) GetUserInfo(mid int64) (*UserInfo, error) {
	var userInfo = &UserInfo{}
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.UserInfoKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), userInfo)
		if err != nil {
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
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentLiveKey(liveInfo.Mid), liveInfo.ToString(), nil)
		return err
	})
	return err
}

func (c *StateManager) GetLiveInfo(mid int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
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
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(c.CurrentNewsKey(newsInfo.Mid), newsInfo.ToString(), nil)
		return err
	})
	return err
}

func (c *StateManager) GetNewsInfo(mid int64) (*NewsInfo, error) {
	var newsInfo = &NewsInfo{}
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentNewsKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), newsInfo)
		if err != nil {
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
	return c.StateManager.Start()
}

func NewStateManager(emitChan chan interface{}) *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), emitChan)
	return sm
}
