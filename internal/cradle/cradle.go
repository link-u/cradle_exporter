package cradle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"text/template"

	"github.com/Code-Hex/golet"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Target interface {
	CreateService() *golet.Service
	Scrape(ctx context.Context, w io.Writer)
	ConfigFilePath() string
}

type Cradle struct {
	WebConfig WebConfig
	Targets   map[string]Target
}

func New(config *Config) (*Cradle, error) {
	configs := make(map[string]*TargetConfig)
	for _, dir := range config.IncludeDirs {
		if err := collectTargetConfigsFromDir(dir, configs); err != nil {
			return nil, err
		}
	}
	targets := make(map[string]Target)
	for fpath, cfg := range configs {
		target := NewTarget(cfg)
		if target == nil {
			yamlBytes, err := yaml.Marshal(cfg)
			if err != nil {
				return nil, fmt.Errorf("invalid config(unknown target type): \n%v", err)
			}
			return nil, fmt.Errorf("invalid config(unknown target type): \n%s", string(yamlBytes))
		}
		targets[fpath] = target
	}
	cradle := &Cradle{
		WebConfig: config.Web,
		Targets:   targets,
	}
	return cradle, nil
}

func NewTarget(cfg *TargetConfig) Target {
	switch {
	case cfg.StaticConfig != nil:
		return &StaticFileTarget{
			Config: cfg,
		}
	case cfg.CronJobConfig != nil:
		return &CronJobTarget{
			Config: cfg,
		}
	case cfg.ScriptConfig != nil:
		return &ScriptTarget{
			Config: cfg,
		}
	case cfg.ServiceConfig != nil:
		return &ServiceTarget{
			Config: cfg,
		}
	case cfg.ExporterConfig != nil:
		return &ExporterTarget{
			Config: cfg,
		}
	default:
		return nil
	}
}

type GoLetToZapLogger struct{}

// Implements io.Writer
func (_ GoLetToZapLogger) Write(p []byte) (n int, err error) {
	zap.L().Info("go-let", zap.String("msg", string(p)))
	return len(p), nil
}

func (cradle *Cradle) Run() error {
	p := golet.New(context.Background())
	p.SetLogger(GoLetToZapLogger{})
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
		if err := t.Execute(&out, cradle.WebConfig); err != nil {
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
	r.Handle(cradle.WebConfig.MetricPath, promhttp.Handler())
	r.HandleFunc(cradle.WebConfig.CollectedPath, func(w http.ResponseWriter, r *http.Request) {
		for name, target := range cradle.Targets {
			var buff bytes.Buffer
			target.Scrape(r.Context(), &buff)
			_, _ = io.WriteString(w, "################################################################################\n")
			_, _ = io.WriteString(w, fmt.Sprintf("### From: %s\n", name))
			_, _ = io.WriteString(w, "################################################################################\n\n")
			_, _ = io.Copy(w, &buff)
			_, _ = w.Write([]byte("\n"))
		}
	})
	go func() {
		ctx := ctx
		baseContext := func(listener net.Listener) context.Context {
			ctx := ctx
			return ctx
		}
		server := &http.Server{
			Addr:        cradle.WebConfig.ListenAddress,
			Handler:     r,
			BaseContext: baseContext,
		}
		err := server.ListenAndServe()
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
				log.Info("Signal caught", zap.String("signal", signal.String()))
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}
