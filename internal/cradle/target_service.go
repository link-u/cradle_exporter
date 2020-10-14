package cradle

import (
	"context"
	"io"

	"github.com/Code-Hex/golet"
)

type ServiceTarget struct {
	Config *TargetConfig
}

func (target *ServiceTarget) CreateService() *golet.Service {
	return &golet.Service{
		Exec: target.Config.ServiceConfig.Path,
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
