package cradle

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/immortal/immortal"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
)

type Daemon struct {
	Path   string
	Config *immortal.Config `yaml:"cofig,omitempty"`
	Daemon *immortal.Daemon
}

type Script struct {
	Path string
}

type Cron struct {
	Script
	Schedule string
}

type File struct {
	Path string
}

type Cradle struct {
	Daemons []Daemon
	Scripts []Script
	Crons   []Cron
	Files   []File
}

func (cradle *Cradle) Run() error {
	var err error
	for _, daemon := range cradle.Daemons {
		daemon.Daemon, err = immortal.New(daemon.Config)
		if err != nil {
			return err
		}
	}
	var wg sync.WaitGroup
	for _, daemon := range cradle.Daemons {
		wg.Add(1)
		d := daemon
		go func() {
			defer wg.Done()
			err := immortal.Supervise(d.Daemon)
			if err != nil {
				log.Error("Failed to supervise daemon", zap.String("config", d.Path), zap.Error(err))
			}
		}()
	}
	wg.Wait()
	return nil
}

func (cradle *Cradle) Scrape() ([]byte, error) {
	var result bytes.Buffer
	var err error

	// Daemons

	// Files
	for _, file := range cradle.Files {
		err = cradle.scrapePath(&result, file.Path)
		if err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}

func (cradle *Cradle) scrapePath(w io.Writer, p string) error {
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
		return cradle.scrapePath(w, p)
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
		return cradle.scrapePath(w, p)
	}
	if info.Mode().IsDir() {
		return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			return cradle.scrapePath(w, path)
		})
	}
	log.Warn("Unknown file type", zap.Uint("mode", uint(info.Mode())))
	return nil
}
