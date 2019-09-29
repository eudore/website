package controller

import (
	"fmt"
	"sync"

	"github.com/eudore/eudore"
)

type (
	ControllerSingleFlight struct {
		eudore.ContextData
		mu    *sync.Mutex
		calls map[string]*Call // 对于每一个需要获取的key有一个对应的call
	}

	// Call 代表需要被执行的函数
	Call struct {
		wg  sync.WaitGroup // 用于阻塞这个调用call的其他请求
		val interface{}    // 函数执行后的结果
		err error          // 函数执行后的error
	}
)

func NewControllerSingleFlight() eudore.Controller {
	return &ControllerSingleFlight{
		mu:    &sync.Mutex{},
		calls: make(map[string]*Call),
	}
}

// Init 实现控制器初始方法。
func (ctl *ControllerSingleFlight) Init(ctx eudore.Context) error {
	ctl.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerSingleFlight) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法。
func (ctl *ControllerSingleFlight) Inject(controller eudore.Controller, router eudore.RouterMethod) error {
	return eudore.ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerSingleFlight) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerSingleFlight) GetRouteParam(pkg, name, method string) string {
	return fmt.Sprintf("controllername=%s.%s controllermethod=%s", pkg, name, method)
}

func (ctl *ControllerSingleFlight) Do(fn func() (interface{}, error)) (interface{}, error) {
	return ctl.DoWithKey(ctl.Path(), fn)
}

func (ctl *ControllerSingleFlight) DoWithKey(key string, fn func() (interface{}, error)) (interface{}, error) {
	ctl.mu.Lock()
	// 如果获取当前key的函数正在被执行，则阻塞等待执行中的，等待其执行完毕后获取它的执行结果
	if c, ok := ctl.calls[key]; ok {
		ctl.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// 初始化一个call，往map中写后就解
	c := new(Call)
	c.wg.Add(1)
	ctl.calls[key] = c
	ctl.mu.Unlock()

	// 执行获取key的函数，并将结果赋值给这个Call
	c.val, c.err = fn()
	c.wg.Done()

	// 重新上锁删除key
	ctl.mu.Lock()
	delete(ctl.calls, key)
	ctl.mu.Unlock()

	return c.val, c.err
}
