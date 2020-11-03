package cradle

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Target interface {
	Scrape(ctx context.Context, w io.Writer)
	ConfigFilePath() string
}

type Cradle struct {
	configValue  atomic.Value
	targetsValue atomic.Value
	haltedValue  atomic.Bool
	//
	serverValue atomic.Value
	runnerValue atomic.Value
}

func New(config *Config) *Cradle {
	cradle := &Cradle{
		configValue:  atomic.Value{},
		targetsValue: atomic.Value{},
		haltedValue:  atomic.Bool{},
		//
		serverValue: atomic.Value{},
	}
	cradle.configValue.Store(config)
	return cradle
}

func (cradle *Cradle) Check(cfg *Config) error {
	_, err := newTargets(cfg)
	return err
}

func (cradle *Cradle) Reload(config *Config) error {
	log := zap.L()
	targets, err := newTargets(config)
	if err != nil {
		log.Error("Failed to read config file. Nothing reloaded.", zap.Error(err))
		return err
	}
	newServer, err := cradle.createServer(config)
	if err != nil {
		log.Error("Failed to create server. Nothing reloaded.", zap.Error(err))
		return err
	}
	newRunner, err := cradle.createRunner(targets)
	if err != nil {
		log.Error("Failed to create runner. Nothing reloaded.", zap.Error(err))
		return err
	}

	cradle.targetsValue.Store(targets)
	cradle.configValue.Store(config)
	// Swap server
	oldServer := cradle.Server()
	cradle.serverValue.Store(newServer)
	if oldServer != nil {
		oldServer.Shutdown()
	}
	// Swap runner
	oldRunner := cradle.Runner()
	cradle.runnerValue.Store(newRunner)
	if oldRunner != nil {
		oldRunner.Shutdown()
	}
	return nil
}

func (cradle *Cradle) Run() error {
	log := zap.L()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for !cradle.isHalted() {
			server := cradle.Server()
			if server == nil {
				continue
			}
			err := server.Run()
			if err != nil {
				log.Error("Failed to run server", zap.Error(err))
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for !cradle.isHalted() {
			runner := cradle.Runner()
			if runner == nil {
				continue
			}
			err := runner.Run()
			if err != nil {
				log.Error("Failed to run runner", zap.Error(err))
			}
		}
	}()
	wg.Wait()
	return nil
}

func (cradle *Cradle) Shutdown() error {
	cradle.haltedValue.Store(true)
	if server := cradle.Server(); server != nil {
		server.Shutdown()
	}
	return nil
}

// ---

func (cradle *Cradle) Config() *Config {
	if config, ok := cradle.configValue.Load().(*Config); ok {
		return config
	}
	return nil
}

func (cradle *Cradle) Targets() map[string]Target {
	if targets, ok := cradle.targetsValue.Load().(map[string]Target); ok {
		return targets
	}
	return nil
}

func (cradle *Cradle) Server() *Server {
	if server, ok := cradle.serverValue.Load().(*Server); ok {
		return server
	}
	return nil
}

func (cradle *Cradle) Runner() *Runner {
	if runner, ok := cradle.runnerValue.Load().(*Runner); ok {
		return runner
	}
	return nil
}

func (cradle *Cradle) isHalted() bool {
	return cradle.haltedValue.Load()
}

// ---
func (cradle *Cradle) createRunner(targets map[string]Target) (*Runner, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := Runner{
		context: ctx,
		cancel:  cancel,
		targets: targets,
		cron:    cron.New(),
		daemons: make([]*ServiceTarget, 0),
	}
	for _, target := range r.targets {
		switch target := target.(type) {
		case *CronJobTarget:
			config := target.Config.CronJobConfig
			err := r.cron.AddFunc(config.Every, func() {
				_ = target.update(r.context)
			})
			if err != nil {
				return nil, err
			}
		case *ServiceTarget:
			r.daemons = append(r.daemons, target)
		}
	}
	return &r, nil
}

func (cradle *Cradle) createServer(config *Config) (*Server, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil or wrong interface: config=%v", cradle.configValue.Load())
	}
	var tlsConfig *tls.Config = nil
	if len(config.Web.ServerTLSKeyPath) > 0 && len(config.Web.ServerTLSKeyPath) > 0 {
		serverCert, err := tls.LoadX509KeyPair(config.Web.ServerTLSCertPath, config.Web.ServerTLSKeyPath)
		if err != nil {
			return nil, fmt.Errorf("could not parse key/cert: %v", err)
		}
		clientCAs := x509.NewCertPool()
		if len(config.Web.ClientCAPath) > 0 {
			pem, err := ioutil.ReadFile(config.Web.ClientCAPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read client certs from file: path='%v', err='%v'", config.Web.ClientCAPath, err)
			}
			if !clientCAs.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("failed to add client certs from file: %s", config.Web.ClientCAPath)
			}
		}
		tlsConfig = &tls.Config{
			Rand:               rand.Reader,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{serverCert},
			ClientAuth:         tls.RequireAndVerifyClientCert,
			ClientCAs:          clientCAs,
		}
	}
	handler := cradle.createServerHandler(config)
	server := Server{
		listenAddress: config.Web.ListenAddress,
		tlsConfig:     tlsConfig,
		handler:       handler,
		listener:      atomic.Value{},
	}
	return &server, nil
}

func (cradle *Cradle) createServerHandler(config *Config) *mux.Router {
	log := zap.L()
	r := mux.NewRouter().StrictSlash(true)
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
		<p>Currently listen at {{ .Config.Web.ListenAddress }}</p>
		<h2>Metrics</h2>
			<ul>
				<li><a href="{{ .Config.Web.MetricPath }}">Metrics</a></li>
				<li><a href="{{ .Config.Web.ProbePath }}">Probed Metrics</a></li>
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
	r.Handle(config.Web.MetricPath, promhttp.Handler())
	r.HandleFunc(config.Web.ProbePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		targets := cradle.Targets()
		if targets != nil {
			for name, target := range targets {
				var buff bytes.Buffer
				target.Scrape(r.Context(), &buff)
				_, _ = io.WriteString(w, "################################################################################\n")
				_, _ = io.WriteString(w, fmt.Sprintf("### From: %s\n", name))
				_, _ = io.WriteString(w, "################################################################################\n\n")
				_, _ = io.Copy(w, &buff)
				_, _ = w.Write([]byte("\n"))
			}
		}
	})
	return r
}
