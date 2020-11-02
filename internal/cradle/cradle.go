package cradle

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"text/template"

	"github.com/Code-Hex/golet"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Target interface {
	CreateService() *golet.Service
	Scrape(ctx context.Context, w io.Writer)
	ConfigFilePath() string
}

type Cradle struct {
	WebConfig     WebConfig
	Targets       map[string]Target
	Runner        *Runner
	Server        http.Server
	ServerHandler *mux.Router
}

func New(config *Config) *Cradle {
	cradle := &Cradle{
		WebConfig: config.Web,
		Targets:   make(map[string]Target),
		Runner:    nil,
		Server:    http.Server{},
	}
	cradle.initServer()
	return cradle
}

func (cradle *Cradle) Check(cfg *Config) error {
	_, err := newTargets(cfg)
	return err
}

func (cradle *Cradle) Reload(cfg *Config) error {
	log := zap.L()
	targets, err := newTargets(cfg)
	if err != nil {
		log.Error("Failed to read config file. Nothing reloaded.", zap.Error(err))
		return err
	}
	runner, err := newRunner(targets)
	if err != nil {
		log.Error("Failed to read config file. Nothing reloaded.", zap.Error(err))
		return err
	}
	if cradle.Runner != nil {
		cradle.Runner.Stop()
		cradle.Runner.Wait()
	}
	// start new runner with new targets
	cradle.Targets = targets
	cradle.Runner = runner
	go func() {
		err := cradle.Runner.Run()
		if err != nil {
			log.Error("Failed to execute runner", zap.Error(err))
		}
	}()
	return nil
}

func (cradle *Cradle) ListenAndServe() error {
	err := cradle.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (cradle *Cradle) Shutdown(ctx context.Context) error {
	var err error
	cradle.Runner.Stop()
	cradle.Runner.Wait()

	err = cradle.Server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (cradle *Cradle) stopServer(ctx context.Context) error {
	err := cradle.Server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

// ---

func (cradle *Cradle) initServer() {
	log := zap.L()
	r := mux.NewRouter().StrictSlash(true)
	cradle.Server.Handler = r
	cradle.Server.Addr = cradle.WebConfig.ListenAddress
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		funcMap := template.FuncMap{
			"typeOf": func(obj interface{}) string {
				name := reflect.TypeOf(obj).String()
				idx := strings.LastIndex(name, ".")
				if idx > 0 {
					name = name[idx+1:]
				}
				return name
			},
		}
		t, err := template.New("index").Funcs(funcMap).Parse(`
<html>
		<head><title>Cradle Exporter</title></head>
		<body>
		<h1>Cradle Exporter</h1>
		<p>Currently listen at {{ .WebConfig.ListenAddress }}</p>
		<h2>Metrics</h2>
			<ul>
				<li><a href="{{ .WebConfig.MetricPath }}">Metrics</a></li>
				<li><a href="{{ .WebConfig.CollectedPath }}">CollectedMetrics</a></li>
			</ul>
		<h2>Enabled Targets</h2>
			<ul>
			{{ range $key, $value := .Targets }}
				<li> [{{ typeOf $value }}] {{ html $key }}</li>
			{{ end }}
			</ul>
		</body>
		</html>
`)
		if err != nil {
			http.Error(w, fmt.Sprintf("[Failed to parse template] %v", err), 500)
			return
		}
		var out bytes.Buffer
		if err := t.Execute(&out, cradle); err != nil {
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
	r.HandleFunc(cradle.WebConfig.ProbePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
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
}
