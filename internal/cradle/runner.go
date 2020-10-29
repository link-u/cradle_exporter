package cradle

import (
	"context"
	"sync"
	"syscall"
	"time"

	"github.com/Code-Hex/golet"
)

type Runner struct {
	impl      golet.Runner
	context   context.Context
	canceller context.CancelFunc
	wg        sync.WaitGroup
}

func newRunner(targets map[string]Target) (*Runner, error) {
	ctx, canceller := context.WithCancel(context.Background())
	p := golet.New(ctx)
	p.EnableColor()
	p.SetInterval(time.Second * 1)
	p.SetCtxCancelSignal(syscall.SIGTERM)

	p.SetLogger(GoLetToZapLogger{})

	for _, target := range targets {
		s := target.CreateService()
		if s != nil {
			if err := p.Add(*s); err != nil {
				canceller()
				return nil, err
			}
		}
	}
	r := Runner{
		impl:      p,
		context:   ctx,
		canceller: canceller,
		wg:        sync.WaitGroup{},
	}
	return &r, nil
}

func (r *Runner) Run() error {
	r.wg.Add(1)
	defer r.wg.Done()
	return r.impl.Run()
}

func (r *Runner) Stop() {
	r.canceller()
}

func (r *Runner) Wait() {
	r.wg.Wait()
}
