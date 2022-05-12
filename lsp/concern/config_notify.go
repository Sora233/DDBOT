package concern

import "github.com/Sora233/DDBOT/lsp/concern_type"

// ConcernNotifyConfig 推送配置
type ConcernNotifyConfig struct {
	TitleChangeNotify concern_type.Type `json:"title_change_notify"`
	OfflineNotify     concern_type.Type `json:"offline_notify"`
}

func (g *ConcernNotifyConfig) CheckTitleChangeNotify(ctype concern_type.Type) bool {
	return g.TitleChangeNotify.ContainAll(ctype)
}

func (g *ConcernNotifyConfig) CheckOfflineNotify(ctype concern_type.Type) bool {
	return g.OfflineNotify.ContainAll(ctype)
}
