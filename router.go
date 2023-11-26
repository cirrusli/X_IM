package X_IM

import (
	"X_IM/pkg/wire/pkt"
	"fmt"
	"sync"
)

//var ErrSessionLost = errors.New("err:session lost")

// Router defines
type Router struct {
	middlewares []HandlerFunc
	handlers    *FuncTree
	pool        sync.Pool
}

func NewRouter() *Router {
	r := &Router{
		handlers:    NewTree(),
		middlewares: make([]HandlerFunc, 0),
	}
	r.pool.New = func() any {
		return BuildContext()
	}
	return r
}

func (r *Router) Use(middlewares ...HandlerFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// Handle 注册指令路由
func (r *Router) Handle(command string, handlers ...HandlerFunc) {
	r.handlers.Add(command, r.middlewares...)
	r.handlers.Add(command, handlers...)
}

// Serve 发送消息到指令路由
func (r *Router) Serve(packet *pkt.LogicPkt, dispatcher Dispatcher, cache SessionStorage, session Session) error {
	if dispatcher == nil {
		return fmt.Errorf("dispatcher is nil")
	}
	if cache == nil {
		return fmt.Errorf("cache is nil")
	}
	ctx := r.pool.Get().(*ContextImpl)
	ctx.reset()
	ctx.request = packet
	ctx.Dispatcher = dispatcher
	ctx.SessionStorage = cache
	ctx.session = session

	r.serveContext(ctx)
	r.pool.Put(ctx)
	return nil
}

// serveContext 责任链模式
func (r *Router) serveContext(ctx *ContextImpl) {
	chain, ok := r.handlers.Get(ctx.Header().Command)
	if !ok {
		ctx.handlers = []HandlerFunc{handleNotFound}
		ctx.Next()
		return
	}
	ctx.handlers = chain
	ctx.Next()
}

// handleNotFound is the default handler when no route is found
func handleNotFound(ctx Context) {
	_ = ctx.Resp(pkt.Status_NotImplemented, &pkt.ErrorResp{Message: "NotImplemented"})
}

// FuncTree is a tree structure
type FuncTree struct {
	nodes map[string]HandlersChain
}

func NewTree() *FuncTree {
	return &FuncTree{nodes: make(map[string]HandlersChain, 10)}
}

// Add a handler to tree
func (t *FuncTree) Add(path string, handlers ...HandlerFunc) {
	if t.nodes[path] == nil {
		t.nodes[path] = HandlersChain{}
	}

	t.nodes[path] = append(t.nodes[path], handlers...)
}

// Get a handler from tree
func (t *FuncTree) Get(path string) (HandlersChain, bool) {
	f, ok := t.nodes[path]
	return f, ok
}
