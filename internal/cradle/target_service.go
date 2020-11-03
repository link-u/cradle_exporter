package cradle

import (
	"context"
	"io"
)

type ServiceTarget struct {
	Config *TargetConfig
}

func (target *ServiceTarget) Scrape(ctx context.Context, w io.Writer) {
	for _, endpoint := range target.Config.ServiceConfig.Endpoints {
		scrapeEndpoint(ctx, w, target.Config.ConfigFilePath, endpoint)
	}
}

func (target *ServiceTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
