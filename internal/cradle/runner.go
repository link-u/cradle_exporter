package cradle

import (
	"context"
	"os/exec"
	"sync"
	"syscall"

	"github.com/robfig/cron"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Runner struct {
	context context.Context
	cancel  context.CancelFunc
	targets map[string]Target
	cron    *cron.Cron
	daemons []*ServiceTarget
	halted  atomic.Bool
}

type ZapInfoWriter struct{}

func (_ ZapInfoWriter) Write(p []byte) (n int, err error) {
	log := zap.L()
	log.Info("From daemon", zap.String("msg", string(p)))
	return len(p), err
}

type ZapErrorWriter struct{}

func (_ ZapErrorWriter) Write(p []byte) (n int, err error) {
	log := zap.L()
	log.Error("From daemon", zap.String("msg", string(p)))
	return len(p), err
}

func (r *Runner) Run() error {
	var wg sync.WaitGroup
	r.cron.Start()
	for _, daemon := range r.daemons {
		log := zap.L()
		wg.Add(1)
		go func(daemon *ServiceTarget) {
			defer wg.Done()
			var err error
			for !r.halted.Load() {
				args := []string{daemon.Config.ServiceConfig.Path}
				args = append(args, daemon.Config.ServiceConfig.Args...)
				log.Info("Daemon starting...",
					zap.String("config-path", daemon.ConfigFilePath()),
					zap.Strings("args", args))
				cmd := exec.CommandContext(r.context, daemon.Config.ServiceConfig.Path, daemon.Config.ServiceConfig.Args...)
				err = cmd.Start()
				if err != nil {
					log.Error("Failed to start daemon", zap.String("config-path", daemon.ConfigFilePath()), zap.Error(err))
				}
				cmd.Stdout = ZapInfoWriter{}
				cmd.Stdout = ZapErrorWriter{}
				err = cmd.Wait()
				if err != nil {
					if exiterr, ok := err.(*exec.ExitError); ok {
						// See https://stackoverflow.com/a/10385867
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							if status.Signaled() {
								log.Error("Daemon caught signal", zap.String("config-path", daemon.ConfigFilePath()), zap.String("signal", status.StopSignal().String()))
							}
							if status.Exited() {
								log.Error("Daemon dead", zap.String("config-path", daemon.ConfigFilePath()), zap.Int("exit-status", status.ExitStatus()))
							}
						}
					}
				}
			}
		}(daemon)
	}
	wg.Wait()
	return nil
}

func (r *Runner) Shutdown() {
	r.halted.Store(true)
	r.cron.Stop()
	r.cancel()
}
