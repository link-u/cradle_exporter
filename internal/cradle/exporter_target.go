package cradle

import (
	"bytes"

	"github.com/Code-Hex/golet"
)

type ExporterTarget struct {
	Config *TargetConfig
}

func (target *ExporterTarget) CreateService() *golet.Service {
	return nil
}

func (target *ExporterTarget) Scrape() ([]byte, error) {
	var result bytes.Buffer
	for _, endpoint := range target.Config.ExporterConfig.Endpoints {
		err := scrapeEndpoint(&result, target.Config.ConfigFilePath, endpoint)
		if err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}

func (target *ExporterTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
