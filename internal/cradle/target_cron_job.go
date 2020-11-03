package cradle

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"go.uber.org/zap"
)

type CronJobTarget struct {
	Config     *TargetConfig
	lastResult []byte
}

func (target *CronJobTarget) Scrape(ctx context.Context, w io.Writer) {
	log := zap.L()
	if target.lastResult == nil {
		err := target.update(ctx)
		if err != nil {
			log.Error("Err: Failed to update target (on the fly)", zap.Error(err))
			_, _ = io.WriteString(w, "### Cron Job Target\n")
			_, _ = io.WriteString(w, "### Err: Failed to execute target (on the fly)\n")
			_, _ = io.WriteString(w, "### Config: "+target.ConfigFilePath()+"\n")
			_, _ = io.WriteString(w, promCommentOut(err.Error()))
			return
		}
	}
	_, _ = io.WriteString(w, "### Cron Job Target\n")
	_, _ = io.WriteString(w, "### Config: "+target.ConfigFilePath()+"\n")
	_, err := w.Write(target.lastResult)
	if err != nil {
		log.Error("Failed to write out last result (on the fly)", zap.Error(err))
	}
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
