package task

import (
	"github.com/eudore/eudore"
)

type (
	HttpTrigger struct {
		Event  chan *Event
		data   map[string]*Event
		router *router
	}
	router struct {
		get    routerNode
		post   routerNode
		put    routerNode
		delete routerNode
		head   routerNode
	}
	routerNode struct {
		path string
		name string
		kind uint8
		pnum uint8

		Cchildren []*routerNode
		Rchildren []*routerNode
		Pchildren []*routerNode
		Vchildren []*routerNode
		Wchildren *routerNode
		// 校验函数
		check eudore.RouterCheckFunc
		// data
		event Event
	}
)

func NewHttpTrigger(sc chan *Event) *HttpTrigger {
	return &HttpTrigger{
		data:  make(map[string]*Event),
		Event: sc,
	}
}

func (tg *HttpTrigger) Handle(ctx eudore.Context) {
	path := ctx.GetParam("path")
	if path == "" {
		path = ctx.Path()
	}
	ctx.Debug(ctx.Method(), path)
	event := tg.data[path]
	if event != nil {
		tg.Event <- event
	}

	// tg.Event <- tg.router.Match(ctx.Method(), path, make(map[string]string))
}

func (tg *HttpTrigger) AddEntry(i map[string]interface{}) error {
	e := NewEvent(i)
	tg.data[e.Params["url"].(string)] = e
	return nil
}

func (r *router) Match(method, path string, params map[string]string) *Event {
	return nil
}

func (r *router) Insert(method, path string, event Event) {

}
