package version

type oldType int64

const (
	bibiliLive oldType = 1 << iota
	bilibiliNews
	douyuLive
	youtubeLive
	youtubeVideo
	huyaLive
)

type atSomeone struct {
	Ctype  oldType `json:"ctype"`
	AtList []int64 `json:"at_list"`
}

type groupConcernAtConfig struct {
	AtAll     oldType      `json:"at_all"`
	AtSomeone []*atSomeone `json:"at_someone"`
}

type groupConcernNotifyConfig struct {
	TitleChangeNotify oldType `json:"title_change_notify"`
	OfflineNotify     oldType `json:"offline_notify"`
}

type groupConcernFilterConfig struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

type groupConcernConfig struct {
	GroupConcernAt     groupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify groupConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter groupConcernFilterConfig `json:"group_concern_filter"`
}

func V1() error {
	return nil
}
