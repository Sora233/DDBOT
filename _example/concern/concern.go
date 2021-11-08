package example_concern

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
)

var logger = utils.GetModuleLogger("example-concern")

const (
	Site string = "example"

	// Example 定义了本插件支持的一种订阅类型，如果支持多种订阅类型，可以继续定义更多
	Example concern_type.Type = "example"
)

type exampleStateManager struct {
	*concern.StateManager
}

// GetGroupConcernConfig 重写 concern.StateManager 的GetGroupConcernConfig方法，让我们自己定义的 GroupConcernConfig 生效
func (c *exampleStateManager) GetGroupConcernConfig(groupCode int64, id interface{}) concern.IConfig {
	return NewGroupConcernConfig(c.StateManager.GetGroupConcernConfig(groupCode, id))
}

// exampleConcern 通过 exampleStateManager 实现 concern.Concern 接口
type exampleConcern struct {
	*exampleStateManager
}

// Site 实现 concern.Concern 的Site接口
// 它的返回值应该是一个常量，会被用作watch命令的-s参数
// 这个例子中即是 /watch -s example
func (c *exampleConcern) Site() string {
	return Site
}

// exampleFresh 自定义 concern.FreshFunc，使用了EmitQueueFresher来刷新事件，
// EmitQueue是开箱即用的轮训组件，会自动每隔一段时间对每个id调用刷新函数。
func (c *exampleConcern) exampleFresh() concern.FreshFunc {
	return c.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		if p.ContainAny(Example) {
			return []concern.Event{&Event{id: id.(string)}}, nil
		}
		return nil, nil
	})
}

// exampleNotifyGenerator 使用了自定义的 concern.NotifyGeneratorFunc 函数，这个例子中简单地把 Event 直接转换成 Notify
func (c *exampleConcern) exampleNotifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, ievent concern.Event) []concern.Notify {
		exampleEvent := ievent.(*Event)
		exampleNotify := &Notify{groupCode: groupCode, Event: exampleEvent}
		return []concern.Notify{exampleNotify}
	}
}

// Start 会在框架准备好的时候被调用，刷新订阅的工作应该在 Start 之后运行。
func (c *exampleConcern) Start() error {
	c.UseFreshFunc(c.exampleFresh())
	c.UseNotifyGeneratorFunc(c.exampleNotifyGenerator())
	return c.StateManager.Start()
}

// Stop 会在程序结束的时候被调用，此时应该保存所有需要的状态并退出所有启动的Goroutine，如果需要的话 可以用 sync.WaitGroup 等组件。
func (c *exampleConcern) Stop() {
	c.StateManager.Stop()
}

// ParseId 会解析一个id，此处返回的id类型，即是其他地方id interface{}的类型
// 其他所有地方的id都由此函数生成
// 推荐在string 或者 int64类型中选择其一
// 如果订阅源有uid等数字唯一标识，请选择int64，如 bilibili
// 如果订阅源有数字并且有字符，请选择string， 如 douyu
func (c *exampleConcern) ParseId(s string) (interface{}, error) {
	return s, nil
}

// Add 实现添加一个订阅
func (c *exampleConcern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := c.GetStateManager().AddGroupConcern(groupCode, id.(string), ctype)
	if err != nil {
		return nil, err
	}
	return c.Get(id)
}

// Remove 实现删除一个订阅
func (c *exampleConcern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := c.GetStateManager().RemoveGroupConcern(groupCode, id.(string), ctype)
	if err != nil {
		return nil, err
	}
	return c.Get(id)
}

// List 实现查询所有订阅
func (c *exampleConcern) List(groupCode int64, ctype concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	_, ids, ctypes, err := c.GetStateManager().ListConcernState(func(_groupCode int64, id interface{}, p concern_type.Type) bool {
		return groupCode == _groupCode && p.ContainAny(ctype)
	})
	if err != nil {
		return nil, nil, err
	}
	ids, ctypes, err = c.GetStateManager().GroupTypeById(ids, ctypes)
	if err != nil {
		return nil, nil, err
	}
	var result []concern.IdentityInfo
	var resultType []concern_type.Type
	for index, id := range ids {
		info, err := c.Get(id)
		if err != nil {
			continue
		}
		result = append(result, info)
		resultType = append(resultType, ctypes[index])
	}
	return result, resultType, nil
}

// Get 实现查询单个订阅
func (c *exampleConcern) Get(id interface{}) (concern.IdentityInfo, error) {
	return concern.NewIdentity(id, id.(string)), nil
}

// GetStateManager 返回我们自定义修改过的 concern.IStateManager，让所有修改对框架生效
func (c *exampleConcern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

// NewConcern 返回一个新的 exampleConcern， 推荐像这样将 notify channel 通过参数穿进来，方便编写单元测试
func NewConcern(notify chan<- concern.Notify) *exampleConcern {
	sm := &exampleStateManager{concern.NewStateManagerWithStringID("example", notify)}
	sm.UseEmitQueue()
	return &exampleConcern{sm}
}

// init 向框架注册这个插件，引用这个插件即可使用
// 注册时需要指定这个插件支持的订阅类型，在本例子中即是 Example ，指定的类型在watch命令中即是-t的参数
// 这个例子中即是 /watch -s example -t example
func init() {
	concern.RegisterConcernManager(NewConcern(concern.GetNotifyChan()), []concern_type.Type{Example})
}
