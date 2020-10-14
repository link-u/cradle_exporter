package cradle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Code-Hex/golet"
	"go.uber.org/zap"
)

type StaticFileTarget struct {
	Config *TargetConfig
}

func (target *StaticFileTarget) CreateService() *golet.Service {
	return nil
}

func (target *StaticFileTarget) Scrape(_ context.Context, w io.Writer) {
	for _, file := range target.Config.StaticConfig.Paths {
		target.scrapePath(w, file)
	}
}

func (target *StaticFileTarget) scrapePath(w io.Writer, p string) {
	log := zap.L()
	p = filepath.Clean(p)
	configFilePath := target.ConfigFilePath()
	var err error
	info, err := os.Lstat(p)
	if err != nil {
		_, _ = io.WriteString(w, "### Static File Target\n")
		_, _ = io.WriteString(w, "### Err: Failed to lstat file\n")
		_, _ = io.WriteString(w, "### Path: "+p+"\n")
		_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
		_, _ = io.WriteString(w, promCommentOut(err.Error()))
		return
	}
	if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			_, _ = io.WriteString(w, "### Static File Target\n")
			_, _ = io.WriteString(w, "### Err: Failed to eval symlink\n")
			_, _ = io.WriteString(w, "### Path: "+p+"\n")
			_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
			_, _ = io.WriteString(w, promCommentOut(err.Error()))
		}
		target.scrapePath(w, p)
		return
	}
	if info.Mode().IsRegular() {
		var written int64
		file, err := os.Open(p)
		if err != nil {
			_, _ = io.WriteString(w, "### Static File Target\n")
			_, _ = io.WriteString(w, "### Err: Failed to open file\n")
			_, _ = io.WriteString(w, "### Path: "+p+"\n")
			_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
			_, _ = io.WriteString(w, promCommentOut(err.Error()))
			return
		}
		defer func() {
			err = file.Close()
			if err != nil {
				log.Warn("Failed to close file", zap.String("path", p), zap.Error(err))
			}
		}()
		var buff bytes.Buffer
		written, err = io.Copy(&buff, file)
		if err != nil {
			_, _ = io.WriteString(w, "### Static File Target\n")
			_, _ = io.WriteString(w, "### Err: Failed to read file\n")
			_, _ = io.WriteString(w, "### Path: "+p+"\n")
			_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
			_, _ = io.WriteString(w, promCommentOut(err.Error()))
		}
		if written != info.Size() {
			log.Warn("Failed to copy all contents of the file", zap.String("path", p), zap.Int64("size", info.Size()), zap.Int64("written", written))
		}
		_, _ = io.WriteString(w, "### Static File Target\n")
		_, _ = io.WriteString(w, "### Path: "+p+"\n")
		_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
		_, _ = w.Write(buff.Bytes())
		return
	}
	if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			_, _ = io.WriteString(w, "### Static File Target\n")
			_, _ = io.WriteString(w, "### Err: Failed to eval symlink\n")
			_, _ = io.WriteString(w, "### Path: "+p+"\n")
			_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
			_, _ = io.WriteString(w, promCommentOut(err.Error()))
			return
		}
		target.scrapePath(w, p)
		return
	}
	if info.Mode().IsDir() {
		_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			target.scrapePath(w, path)
			return nil
		})
		return
	}
	log.Warn("Unknown file type", zap.String("mode", info.Mode().String()))
	_, _ = io.WriteString(w, "### Static File Target\n")
	_, _ = io.WriteString(w, "### Err: Unknown file type\n")
	_, _ = io.WriteString(w, "### Path: "+p+"\n")
	_, _ = io.WriteString(w, "### Config: "+configFilePath+"\n")
	_, _ = io.WriteString(w, fmt.Sprintf("### FileType: %s", info.Mode().String()))
}

func (target *StaticFileTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
