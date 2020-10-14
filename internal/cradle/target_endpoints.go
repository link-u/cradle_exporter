package cradle

import (
	"context"
	"io"

	"github.com/Code-Hex/golet"
)

type EndpointTarget struct {
	Config *TargetConfig
}

func (target *EndpointTarget) CreateService() *golet.Service {
	return nil
}

func (target *EndpointTarget) Scrape(ctx context.Context, w io.Writer) {
	for _, endpoint := range target.Config.ExporterConfig.Endpoints {
		scrapeEndpoint(ctx, w, target.Config.ConfigFilePath, endpoint)
	}
}

func (target *EndpointTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}