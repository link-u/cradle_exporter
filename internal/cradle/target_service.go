package cradle

import (
	"bytes"

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

func (target *ServiceTarget) Scrape() ([]byte, error) {
	var result bytes.Buffer
	for _, endpoint := range target.Config.ServiceConfig.Endpoints {
		err := scrapeEndpoint(&result, target.Config.ConfigFilePath, endpoint)
		if err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}

func (target *ServiceTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
