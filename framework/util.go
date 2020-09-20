package framework

import (
	"fmt"
	"os"
	"context"
	"os/signal"
)

// Errors 实现多个error组合。
type Errors struct {
	errs []error
}

// NewErrors 创建Errors对象。
func NewErrors() *Errors {
	return &Errors{}
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (err *Errors) HandleError(errs ...error) {
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (err *Errors) Error() string {
	switch len(err.errs) {
	case 0:
		return ""
	case 1:
		return err.errs[0].Error()
	default:
		return fmt.Sprint(err.errs)
	}
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (err *Errors) GetError() error {
	switch len(err.errs) {
	case 0:
		return nil
	case 1:
		return err.errs[0]
	default:
		return err
	}
}

// Signaler 定义一个信号处理对象。
type Signaler struct {
	app         *App
	signalChan  chan os.Signal
	signalFuncs map[os.Signal][]func(*App)
}

// NewSignaler 函数创建一个信号处理者。
func NewSignaler(app *App) *Signaler {
	return &Signaler{
		app:         app,
		signalChan:  make(chan os.Signal),
		signalFuncs: make(map[os.Signal][]func(*App)),
	}
}

// HandleSignal 方法执行对应信号应该函数。
func (s *Signaler) HandleSignal(sig os.Signal) error {
	fns, ok := s.signalFuncs[sig]
	if ok {
		for _, fn := range fns {
			fn(s.app)
		}
	}
	return nil
}

// RegisterSignal 方法注册一个信号响应函数。
func (s *Signaler) RegisterSignal(sig os.Signal, fn func(*App)) {
	fns, ok := s.signalFuncs[sig]
	s.signalFuncs[sig] = append(fns, fn)
	if !ok {
		sigs := make([]os.Signal, 0, len(s.signalFuncs))
		for i := range s.signalFuncs {
			sigs = append(sigs, i)
		}

		signal.Stop(s.signalChan)
		signal.Notify(s.signalChan, sigs...)
	}
}

// Run 方法执行Signaler信号响应处理。
func (s *Signaler) Run(ctx context.Context) {
	defer signal.Stop(s.signalChan)
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-s.signalChan:
			s.HandleSignal(sig)
		}
	}
}
