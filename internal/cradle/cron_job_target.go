package cradle

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/Code-Hex/golet"
)

type CronJobTarget struct {
	Config     *TargetConfig
	lastResult []byte
}

func (target *CronJobTarget) CreateService() *golet.Service {
	return &golet.Service{
		Code: func(ctx context.Context) error {
			return target.update(ctx)
		},
		Every: target.Config.CronJobConfig.Spec,
	}
}

func (target *CronJobTarget) Scrape() ([]byte, error) {
	return target.lastResult, nil
}

func (target *CronJobTarget) update(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, target.Config.CronJobConfig.Path, target.Config.CronJobConfig.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return err
	}
	target.lastResult = out.Bytes()
	return nil
}

func (target *CronJobTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
