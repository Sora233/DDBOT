package concern_manager

import "github.com/Sora233/DDBOT/concern"

// GroupConcernNotifyConfig 推送配置
type GroupConcernNotifyConfig struct {
	TitleChangeNotify concern.Type `json:"title_change_notify"`
	OfflineNotify     concern.Type `json:"offline_notify"`
}

func (g *GroupConcernNotifyConfig) CheckTitleChangeNotify(ctype concern.Type) bool {
	return g.TitleChangeNotify.ContainAll(ctype)
}

func (g *GroupConcernNotifyConfig) CheckOfflineNotify(ctype concern.Type) bool {
	return g.OfflineNotify.ContainAll(ctype)
}
