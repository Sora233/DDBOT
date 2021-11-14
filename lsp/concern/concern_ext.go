package concern

// NotifyLiveExt 是一个针对直播推送过滤的扩展接口， Notify 可以选择性实现这个接口，如果实现了，则会自动使用默认的推送过滤逻辑
// 默认情况下，如果 IsLive 为 true，则根据以下规则推送：
// Living 为 true 且 LiveStatusChanged 为true（说明是开播了）进行推送
// Living 为 false 且 LiveStatusChanged 为true（说明是下播了）根据 OfflineNotify 配置进行推送推送
// Living 为 true 且 LiveStatusChanged 为false 且 TitleChanged 为true（说明是上播状态更改标题）根据 TitleChangeNotify 配置进行推送推送
// 如果 Notify 没有实现这个接口，则会推送所有内容
type NotifyLiveExt interface {
	IsLive() bool
	Living() bool
	TitleChanged() bool
	LiveStatusChanged() bool
}
