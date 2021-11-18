package concern

// NotifyLiveExt 是一个针对直播推送过滤的扩展接口， Notify 可以选择性实现这个接口，如果实现了，则会自动使用默认的推送过滤逻辑
// 默认情况下，如果 IsLive 为 true，则根据以下规则推送：
// Living 为 true 且 LiveStatusChanged 为true（说明是开播了）进行推送
// Living 为 false 且 LiveStatusChanged 为true（说明是下播了）根据 OfflineNotify 配置进行推送推送
// Living 为 true 且 LiveStatusChanged 为false 且 TitleChanged 为true（说明是上播状态更改标题）根据 TitleChangeNotify 配置进行推送推送
// 如果 Notify 没有实现这个接口，则会推送所有内容
type NotifyLiveExt interface {
	// IsLive 如果返回false则不会使用扩展逻辑
	IsLive() bool
	// Living 表示是否正在直播，注意有些网站在主播未直播的时候会循环播放主播的投稿或者直播回放
	// 这种情况不能算做正在直播
	Living() bool
	// TitleChanged 表示是否直播标题发生变换，当开启了OfflineNotify配置时会推送这个消息，默认为不推送
	TitleChanged() bool
	// LiveStatusChanged 表示是否直播状态发生了变化，这个配合 Living 可以用来判断是上播还是下播
	// 如果没有变化也可以发送给DDBOT，DDBOT会自动进行过滤
	LiveStatusChanged() bool
}
