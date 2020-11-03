package cradle

import (
	"context"
	"io"
)

type ExporterTarget struct {
	Config *TargetConfig
}

func (target *ExporterTarget) Scrape(ctx context.Context, w io.Writer) {
	for _, endpoint := range target.Config.ExporterConfig.Endpoints {
		scrapeEndpoint(ctx, w, target.Config.ConfigFilePath, endpoint)
	}
}

func (target *ExporterTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
