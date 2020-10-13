package cradle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"text/template"

	"github.com/Code-Hex/golet"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Target interface {
	CreateService() *golet.Service
	Scrape() ([]byte, error)
	ConfigFilePath() string
}

type AppConfig struct {
	MetricPath     string
	CollectedPath  string
	HttpListenAddr string
}

type Cradle struct {
	AppConfig *AppConfig
	Targets   map[string]Target
}

func New(appConfig *AppConfig, config *Config) (*Cradle, error) {
	configs := make(map[string]*TargetConfig)
	for _, dir := range config.IncludeDirs {
		if err := collectTargetConfigsFromDir(dir, configs); err != nil {
			return nil, err
		}
	}
	targets := make(map[string]Target)
	for fpath, cfg := range configs {
		switch {
		case cfg.StaticConfig != nil:
			targets[fpath] = &StaticFileTarget{
				Config: cfg,
			}
		case cfg.CronJobConfig != nil:
			targets[fpath] = &CronJobTarget{
				Config: cfg,
			}
		case cfg.ScriptConfig != nil:
			targets[fpath] = &ScriptTarget{
				Config: cfg,
			}
		case cfg.ServiceConfig != nil:
			targets[fpath] = &ServiceTarget{
				Config: cfg,
			}
		case cfg.ExporterConfig != nil:
			targets[fpath] = &ExporterTarget{
				Config: cfg,
			}
		}
	}
	cradle := &Cradle{
		AppConfig: appConfig,
		Targets:   targets,
	}
	return cradle, nil
}

func (cradle *Cradle) Run() error {
	p := golet.New(context.Background())
	for _, target := range cradle.Targets {
		s := target.CreateService()
		if s != nil {
			if err := p.Add(*s); err != nil {
				return err
			}
		}
	}
	httpService := golet.Service{
		Code: func(ctx context.Context) error {
			return cradle.Expose(ctx)
		},
	}
	if err := p.Add(httpService); err != nil {
		return err
	}
	return p.Run()
}

const indexTemplate = `<html>
<head><title>Cradle Exporter</title></head>
<body>
<h1>Cradle Exporter</h1>
<p>Currently listen at {{ .HttpListenAddr }}</p>
<ul>
  <li><a href="{{ .MetricPath }}">Metrics</a></li>
  <li><a href="{{ .CollectedPath }}">CollectedMetrics</a></li>
</ul>
</body>
</html>
`

func (cradle *Cradle) Expose(ctx context.Context) error {
	log := zap.L()
	c := ctx.(*golet.Context)
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("letter").Parse(indexTemplate))
		var out bytes.Buffer
		if err := t.Execute(&out, &cradle.AppConfig); err != nil {
			http.Error(w, fmt.Sprintf("[Failed to execute template] %v", err), 500)
			return
		}
		w.WriteHeader(200)
		outSize := out.Len()
		written, err := io.Copy(w, &out)
		if err != nil {
			log.Warn("Failed to write entire index page", zap.String("endpoint", "/"), zap.Error(err))
		}
		if int(written) != outSize {
			log.Warn("Body length does not match to content-length header",
				zap.String("endpoint", "/"),
				zap.Int64("written", written),
				zap.Int("response-size", outSize))
		}
	})
	r.Handle(cradle.AppConfig.MetricPath, promhttp.Handler())
	r.HandleFunc(cradle.AppConfig.CollectedPath, func(w http.ResponseWriter, r *http.Request) {
		for name, target := range cradle.Targets {
			out, err := target.Scrape()
			if err == nil {
				_, _ = w.Write([]byte(fmt.Sprintf("### %s\n\n", name)))
				_, _ = w.Write(out)
				_, _ = w.Write([]byte("\n"))
			} else {
				log.Error("Failed to scrape target", zap.String("config-file-path", target.ConfigFilePath()), zap.Error(err))
			}
		}
	})
	go func() {
		err := http.ListenAndServe(cradle.AppConfig.HttpListenAddr, r)
		if err != nil {
			log.Error("Failed to run HTTP server", zap.Error(err))
		}
	}()
	for {
		select {
		// You can notify signal received.
		case <-c.Recv():
			signal, err := c.Signal()
			if err != nil {
				log.Error("Failed to receive signal", zap.Error(err))
				return err
			}
			switch signal {
			case syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT:
				log.Info("Signal catched", zap.String("signal", signal.String()))
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}
