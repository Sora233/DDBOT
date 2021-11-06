package bilibili

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern.StateManager
	*extraKey
	concern *Concern
}

func (c *StateManager) GetGroupConcernConfig(groupCode int64, id interface{}) (concernConfig concern.IConfig) {
	return NewGroupConcernConfig(c.StateManager.GetGroupConcernConfig(groupCode, id), c.concern)
}

func (c *StateManager) AddUserInfo(userInfo *UserInfo) error {
	if userInfo == nil {
		return errors.New("nil UserInfo")
	}
	return c.SetJson(c.UserInfoKey(userInfo.Mid), userInfo)
}

func (c *StateManager) GetUserInfo(mid int64) (*UserInfo, error) {
	var userInfo = &UserInfo{}
	err := c.GetJson(c.UserInfoKey(mid), userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (c *StateManager) AddUserStat(userStat *UserStat, expire time.Duration) error {
	if userStat == nil {
		return errors.New("nil UserStat")
	}
	return c.RWCover(func() error {
		return c.SetJson(c.UserStatKey(userStat.Mid), userStat, localdb.SetExpireOpt(expire))
	})
}

func (c *StateManager) GetUserStat(mid int64) (*UserStat, error) {
	var userStat = &UserStat{}
	err := c.GetJson(c.UserStatKey(mid), userStat)
	if err != nil {
		return nil, err
	}
	return userStat, nil
}

func (c *StateManager) AddLiveInfo(liveInfo *LiveInfo) error {
	if liveInfo == nil {
		return errors.New("nil LiveInfo")
	}
	return c.RWCover(func() error {
		err := c.SetJson(c.UserInfoKey(liveInfo.Mid), liveInfo.UserInfo)
		if err != nil {
			return err
		}
		return c.SetJson(c.CurrentLiveKey(liveInfo.Mid), liveInfo, localdb.SetExpireOpt(time.Hour*24*7))
	})
}

func (c *StateManager) GetLiveInfo(mid int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}
	err := c.GetJson(c.CurrentLiveKey(mid), liveInfo)
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (c *StateManager) AddNewsInfo(newsInfo *NewsInfo) error {
	if newsInfo == nil {
		return errors.New("nil NewsInfo")
	}
	return c.RWCover(func() error {
		err := c.SetJson(c.UserInfoKey(newsInfo.Mid), newsInfo.UserInfo)
		if err != nil {
			return err
		}
		return c.SetJson(c.CurrentNewsKey(newsInfo.Mid), newsInfo)
	})
}

func (c *StateManager) DeleteNewsInfo(mid int64) error {
	_, err := c.Delete(c.CurrentNewsKey(mid))
	return err
}

func (c *StateManager) DeleteLiveInfo(mid int64) error {
	_, err := c.Delete(c.CurrentLiveKey(mid))
	return err
}

func (c *StateManager) DeleteNewsAndLiveInfo(mid int64) error {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(c.CurrentLiveKey(mid))
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		_, err = tx.Delete(c.CurrentNewsKey(mid))
		return err
	})
}

func (c *StateManager) ClearByMid(mid int64) error {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
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
	err := c.GetJson(c.CurrentNewsKey(mid), newsInfo)
	if err != nil {
		return nil, err
	}
	return newsInfo, nil
}

func (c *StateManager) CheckDynamicId(dynamic int64) (result bool) {
	_, err := c.Get(c.DynamicIdKey(dynamic))
	if err == buntdb.ErrNotFound {
		return true
	}
	return false
}

func (c *StateManager) MarkDynamicId(dynamic int64) (bool, error) {
	//	一个错误的写法，用闭包返回值简单地替代了RWTxCover返回值
	//	在磁盘空间用尽的情况下，闭包可以成功执行，但RWTxCover执行持久化时会报错，这个错误就被意外地忽略了
	//	c.RWCoverTx(func(tx *buntdb.Tx) error {
	//		key := c.DynamicIdKey(dynamic)
	//		_, replaced, err = tx.Set(key, "", localdb.ExpireOption(time.Hour*120))
	//		return err
	//	})
	var isOverwrite bool
	err := c.Set(c.DynamicIdKey(dynamic), "",
		localdb.SetExpireOpt(time.Hour*120), localdb.SetGetIsOverwrite(&isOverwrite))
	return isOverwrite, err
}

func (c *StateManager) IncNotLiveCount(uid int64) int64 {
	result, err := c.SeqNext(c.NotLiveKey(uid))
	if err != nil {
		result = 0
	}
	return result
}

func (c *StateManager) ClearNotLiveCount(uid int64) error {
	_, err := c.Delete(c.NotLiveKey(uid), localdb.DeleteIgnoreNotFound())
	return err
}

func (c *StateManager) SetUidFirstTimestampIfNotExist(uid int64, timestamp int64) error {
	return c.SetInt64(c.UidFirstTimestamp(uid), timestamp, localdb.SetNoOverWriteOpt())
}

func (c *StateManager) UnsetUidFirstTimestamp(uid int64) error {
	_, err := c.Delete(c.UidFirstTimestamp(uid))
	return err
}

func (c *StateManager) GetUidFirstTimestamp(uid int64) (timestamp int64, err error) {
	return c.GetInt64(c.UidFirstTimestamp(uid))
}

func (c *StateManager) SetGroupCompactMarkIfNotExist(groupCode int64, compactKey string) error {
	return c.Set(c.CompactMarkKey(groupCode, compactKey), "",
		localdb.SetExpireOpt(CompactExpireTime), localdb.SetNoOverWriteOpt(),
	)
}
func (c *StateManager) SetLastFreshTime(ts int64) error {
	return c.SetInt64(c.LastFreshKey(), ts)
}

func (c *StateManager) GetLastFreshTime() (int64, error) {
	return c.GetInt64(c.LastFreshKey())
}

func (c *StateManager) SetNotifyMsg(notifyKey string, msg *message.GroupMessage) error {
	tmp := &message.GroupMessage{
		Id:        msg.Id,
		GroupCode: msg.GroupCode,
		Sender:    msg.Sender,
		Time:      msg.Time,
		Elements: localutils.MessageFilter(msg.Elements, func(e message.IMessageElement) bool {
			return e.Type() == message.Text || e.Type() == message.Image
		}),
	}
	value, err := localutils.SerializationGroupMsg(tmp)
	if err != nil {
		return err
	}
	return c.Set(c.NotifyMsgKey(tmp.GroupCode, notifyKey), value,
		localdb.SetExpireOpt(CompactExpireTime), localdb.SetNoOverWriteOpt(),
	)
}

func (c *StateManager) GetNotifyMsg(groupCode int64, notifyKey string) (*message.GroupMessage, error) {
	value, err := c.Get(c.NotifyMsgKey(groupCode, notifyKey))
	if err != nil {
		return nil, err
	}
	return localutils.DeserializationGroupMsg(value)
}

func SetCookieInfo(username string, cookieInfo *LoginResponse_Data_CookieInfo) error {
	if cookieInfo == nil {
		return errors.New("<nil> cookieInfo")
	}
	var expire int64
	for _, cookie := range cookieInfo.GetCookies() {
		expire = cookie.GetExpires()
		break
	}
	var opt localdb.OptionFunc
	if expire != 0 {
		opt = localdb.SetExpireOpt(time.Duration(expire-time.Now().Unix()) * time.Second)
	}
	return localdb.SetJson(localdb.BilibiliUserCookieInfoKey(username), cookieInfo, opt)
}

func GetCookieInfo(username string) (cookieInfo *LoginResponse_Data_CookieInfo, err error) {
	err = localdb.GetJson(localdb.BilibiliUserCookieInfoKey(username), &cookieInfo)
	return
}

func ClearCookieInfo(username string) error {
	_, err := localdb.Delete(localdb.BilibiliUserCookieInfoKey(username), localdb.DeleteIgnoreNotFound())
	return err
}

func (c *StateManager) Start() error {
	for _, pattern := range []localdb.KeyPatternFunc{
		c.GroupConcernStateKey, c.CurrentLiveKey, c.FreshKey,
		c.UserInfoKey, c.UserStatKey, c.DynamicIdKey} {
		c.CreatePatternIndex(pattern, nil)
	}
	return c.StateManager.Start()
}

func NewStateManager(c *Concern) *StateManager {
	sm := &StateManager{
		concern: c,
	}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern.NewStateManagerWithCustomKey(Site, NewKeySet(), c.notify)
	return sm
}
