package concern

// GroupConcernNotifyConfig 推送配置
type GroupConcernNotifyConfig struct {
	TitleChangeNotify Type `json:"title_change_notify"`
	OfflineNotify     Type `json:"offline_notify"`
}

func (g *GroupConcernNotifyConfig) CheckTitleChangeNotify(ctype Type) bool {
	return g.TitleChangeNotify.ContainAll(ctype)
}

func (g *GroupConcernNotifyConfig) CheckOfflineNotify(ctype Type) bool {
	return g.OfflineNotify.ContainAll(ctype)
}
