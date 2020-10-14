package cradle

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"github.com/Code-Hex/golet"
)

type ScriptTarget struct {
	Config *TargetConfig
}

func (target *ScriptTarget) CreateService() *golet.Service {
	return nil
}

func (target *ScriptTarget) Scrape(ctx context.Context, w io.Writer) {
	cmd := exec.CommandContext(ctx, target.Config.ScriptConfig.Path, target.Config.ScriptConfig.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		_, _ = io.WriteString(w, "### Script File Target\n")
		_, _ = io.WriteString(w, "### Err: Failed to execute script\n")
		_, _ = io.WriteString(w, "### Config: "+target.ConfigFilePath()+"\n")
		_, _ = io.WriteString(w, promCommentOut(err.Error()))
		return
	}
	_, _ = io.WriteString(w, "### Script File Target\n")
	_, _ = io.WriteString(w, "### Config: "+target.ConfigFilePath()+"\n")
	_, _ = io.Copy(w, &out)
}

func (target *ScriptTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
