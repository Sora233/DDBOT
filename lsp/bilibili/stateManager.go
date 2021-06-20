package bilibili

import (
	"encoding/json"
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"strconv"
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
			logger.WithField("val", val).Errorf("user info json unmarshal failed")
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
		_, _, err := tx.Set(c.UserInfoKey(liveInfo.Mid), liveInfo.UserInfo.ToString(), nil)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(c.CurrentLiveKey(liveInfo.Mid), liveInfo.ToString(), localdb.ExpireOption(time.Hour*24*7))
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
			logger.WithField("val", val).Errorf("json Unmarshal live info error %v", err)
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
		_, _, err := tx.Set(c.UserInfoKey(newsInfo.Mid), newsInfo.UserInfo.ToString(), nil)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(c.CurrentNewsKey(newsInfo.Mid), newsInfo.ToString(), nil)
		return err
	})
}

func (c *StateManager) DeleteNewsInfo(mid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(c.CurrentNewsKey(mid))
		return err
	})
}

func (c *StateManager) DeleteLiveInfo(mid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(c.CurrentLiveKey(mid))
		return err
	})
}

func (c *StateManager) DeleteNewsAndLiveInfo(mid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(c.CurrentLiveKey(mid))
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		_, err = tx.Delete(c.CurrentNewsKey(mid))
		return err
	})
}

func (c *StateManager) ClearByMid(mid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		var errs []error
		_, err := tx.Delete(c.CurrentLiveKey(mid))
		errs = append(errs, err)
		_, err = tx.Delete(c.CurrentNewsKey(mid))
		errs = append(errs, err)
		_, err = tx.Delete(c.UidFirstTimestamp(mid))
		errs = append(errs, err)
		_, err = tx.Delete(c.UserInfoKey(mid))
		errs = append(errs, err)
		_, err = tx.Delete(c.NotLiveKey(mid))
		errs = append(errs, err)
		for _, e := range errs {
			if e != nil && e != buntdb.ErrNotFound {
				return e
			}
		}
		return nil
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
			logger.WithField("mid", mid).WithField("dbval", val).Errorf("Unmarshal error %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return newsInfo, nil
}

func (c *StateManager) CheckDynamicId(dynamic int64) bool {
	var result bool
	c.RTxCover(func(tx *buntdb.Tx) error {
		key := c.DynamicIdKey(dynamic)
		_, err := tx.Get(key)
		if err == nil {
			result = false
		} else if err == buntdb.ErrNotFound {
			result = true
		} else {
			result = false
		}
		return nil
	})
	return result
}

func (c *StateManager) MarkDynamicId(dynamic int64) (replaced bool, err error) {
	c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.DynamicIdKey(dynamic)
		_, replaced, err = tx.Set(key, "", localdb.ExpireOption(time.Hour*120))
		return err
	})
	return
}

func (c *StateManager) IncNotLiveCount(uid int64) (result int) {
	c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.NotLiveKey(uid)
		oldV, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			oldV = "0"
		} else if err != nil {
			return err
		}
		old, err := strconv.Atoi(oldV)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, strconv.Itoa(old+1), localdb.ExpireOption(time.Minute*30))
		if err == nil {
			result = old + 1
		}
		return err
	})
	return
}

func (c *StateManager) ClearNotLiveCount(uid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.NotLiveKey(uid)
		_, err := tx.Delete(key)
		if err == buntdb.ErrNotFound {
			err = nil
		}
		return err
	})
}

func (c *StateManager) SetUidFirstTimestampIfNotExist(uid int64, timestamp int64) error {
	err := c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.UidFirstTimestamp(uid)
		_, replaced, err := tx.Set(key, strconv.FormatInt(timestamp, 10), nil)
		if err != nil {
			return err
		}
		if replaced {
			return localdb.ErrRollback
		}
		return nil
	})
	if err != nil && err == localdb.ErrRollback {
		err = nil
	}
	return err
}

func (c *StateManager) UnsetUidFirstTimestamp(uid int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.UidFirstTimestamp(uid)
		_, err := tx.Delete(key)
		return err
	})
}

func (c *StateManager) GetUidFirstTimestamp(uid int64) (timestamp int64, err error) {
	c.RTxCover(func(tx *buntdb.Tx) error {
		key := c.UidFirstTimestamp(uid)
		var tsStr string
		tsStr, err = tx.Get(key)
		if err != nil {
			return err
		}
		timestamp, err = strconv.ParseInt(tsStr, 10, 64)
		return err
	})
	return
}

func (c *StateManager) Start() error {
	db := localdb.MustGetClient()
	db.CreateIndex(c.GroupConcernStateKey(), c.GroupConcernStateKey("*"), buntdb.IndexString)
	db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
	db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
	db.CreateIndex(c.UserInfoKey(), c.UserInfoKey("*"), buntdb.IndexString)
	db.CreateIndex(c.DynamicIdKey(), c.DynamicIdKey("*"), buntdb.IndexString)
	return c.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), false)
	return sm
}
