package cradle

import (
	"context"
	"io"

	"github.com/Code-Hex/golet"
	"github.com/kballard/go-shellquote"
)

type ServiceTarget struct {
	Config *TargetConfig
}

func (target *ServiceTarget) CreateService() *golet.Service {
	args := []string{target.Config.ServiceConfig.Path}
	args = append(args, target.Config.ServiceConfig.Args...)
	execCmd := shellquote.Join(args...)
	return &golet.Service{
		Exec: execCmd,
		Tag:  target.Config.ConfigFilePath,
	}
}

func (target *ServiceTarget) Scrape(ctx context.Context, w io.Writer) {
	for _, endpoint := range target.Config.ServiceConfig.Endpoints {
		scrapeEndpoint(ctx, w, target.Config.ConfigFilePath, endpoint)
	}
}

func (target *ServiceTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
