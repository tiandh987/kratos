package kratos

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// AppInfo is application context value.
type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

// App is an application components lifecycle manager.
type App struct {
	// 应用选项
	opts     options

	// 带有 cancel 函数的 contxt
	ctx      context.Context
	cancel   func()

	// 用于保护 instance 字段
	mu       sync.Mutex
	instance *registry.ServiceInstance
}

// New create an application lifecycle manager.
// 创建一个应用生命周期管理器
func New(opts ...Option) *App {
	o := options{
		ctx:              context.Background(),
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: 10 * time.Second,
		stopTimeout:      10 * time.Second,
	}

	// 设置 id
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}

	// 选项模式
	// 可以设置 id、name、version、metadata、endpoints、context、logger、servers、sigs、registrar、registrarTimeout、stopTimeout
	for _, opt := range opts {
		opt(&o)
	}

	// 设置自定义 logger
	if o.logger != nil {
		log.SetLogger(o.logger)
	}

	// 生成带有 cancel 的 contxt
	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   o,
	}
}

// ID returns app instance id.
// 返回应用实例 id
func (a *App) ID() string { return a.opts.id }

// Name returns service name.
// 返回应用实例 name
func (a *App) Name() string { return a.opts.name }

// Version returns app version.
// 返回应用实例版本号
func (a *App) Version() string { return a.opts.version }

// Metadata returns service metadata.
// 返回应用实例元数据
func (a *App) Metadata() map[string]string { return a.opts.metadata }

// Endpoint returns endpoints.
func (a *App) Endpoint() []string {
	if a.instance != nil {
		return a.instance.Endpoints
	}
	return nil
}

// Run executes all OnStart hooks registered with the application's Lifecycle.
// Run 执行所有在应用程序生命周期中注册的 OnStart 钩子。
func (a *App) Run() error {
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.instance = instance
	a.mu.Unlock()

	eg, ctx := errgroup.WithContext(NewContext(a.ctx, a))
	wg := sync.WaitGroup{}

	for _, srv := range a.opts.servers {
		srv := srv
		eg.Go(func() error {
			<-ctx.Done() // wait for stop signal
			stopCtx, cancel := context.WithTimeout(NewContext(a.opts.ctx, a), a.opts.stopTimeout)
			defer cancel()
			return srv.Stop(stopCtx)
		})
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			return srv.Start(NewContext(a.opts.ctx, a))
		})
	}
	wg.Wait()
	if a.opts.registrar != nil {
		rctx, rcancel := context.WithTimeout(ctx, a.opts.registrarTimeout)
		defer rcancel()
		if err := a.opts.registrar.Register(rctx, instance); err != nil {
			return err
		}
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				if err := a.Stop(); err != nil {
					return err
				}
			}
		}
	})
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// Stop gracefully stops the application.
// 优雅退出应用
func (a *App) Stop() error {
	a.mu.Lock()
	instance := a.instance
	a.mu.Unlock()
	if a.opts.registrar != nil && instance != nil {
		ctx, cancel := context.WithTimeout(NewContext(a.ctx, a), a.opts.registrarTimeout)
		defer cancel()
		if err := a.opts.registrar.Deregister(ctx, instance); err != nil {
			return err
		}
	}
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

// 构建实例
func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0, len(a.opts.endpoints))

	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}

	if len(endpoints) == 0 {
		for _, srv := range a.opts.servers {
			if r, ok := srv.(transport.Endpointer); ok {
				e, err := r.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, e.String())
			}
		}
	}

	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Version:   a.opts.version,
		Metadata:  a.opts.metadata,
		Endpoints: endpoints,
	}, nil
}

type appKey struct{}

// NewContext returns a new Context that carries value.
// 返回一个携带 value 的 context
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (s AppInfo, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppInfo)
	return
}
