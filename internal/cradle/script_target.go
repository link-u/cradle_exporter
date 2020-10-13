package cradle

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/Code-Hex/golet"
)

type ScriptTarget struct {
	Config *TargetConfig
}

func (target *ScriptTarget) CreateService() *golet.Service {
	return nil
}

func (target *ScriptTarget) Scrape() ([]byte, error) {
	cmd := exec.CommandContext(context.Background(), target.Config.ScriptConfig.Path, target.Config.ScriptConfig.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (target *ScriptTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
