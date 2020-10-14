package cradle

import (
	"bytes"
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

func (target *StaticFileTarget) Scrape() ([]byte, error) {
	var result bytes.Buffer
	var err error

	for _, file := range target.Config.StaticConfig.Paths {
		err = target.scrapePath(&result, file)
		if err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}

func (target *StaticFileTarget) scrapePath(w io.Writer, p string) error {
	log := zap.L()
	p = filepath.Clean(p)
	var err error
	info, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			return err
		}
		return target.scrapePath(w, p)
	}
	if info.Mode().IsRegular() {
		var written int64
		file, err := os.Open(p)
		if err != nil {
			return err
		}
		written, err = io.Copy(w, file)
		if err != nil {
			return err
		}
		if written != info.Size() {
			log.Warn("Failed to copy all contents of the file", zap.String("path", p), zap.Int64("size", info.Size()), zap.Int64("written", written))
		}
		defer func() {
			err = file.Close()
			if err != nil {
				log.Warn("Failed to close file", zap.String("path", p), zap.Error(err))
			}
		}()
		return nil
	}
	if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			return err
		}
		return target.scrapePath(w, p)
	}
	if info.Mode().IsDir() {
		return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			return target.scrapePath(w, path)
		})
	}
	log.Warn("Unknown file type", zap.Uint("mode", uint(info.Mode())))
	return nil
}

func (target *StaticFileTarget) ConfigFilePath() string {
	return target.Config.ConfigFilePath
}
