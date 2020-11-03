package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/link-u/cradle_exporter/internal/cradle"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "v1.0.0"

/*****************************************************************************
  Flags
 *****************************************************************************/

var isConfigCheckMode = kingpin.
	Flag("test-config", "Check config file and exit.").
	Short('t').
	Default("false").
	Bool()

var standardLogOverridden = false
var standardLog = kingpin.
	Flag("cli.standard-log", "Print logs in standard format, not in json").
	Default("false").
	Action(func(_ *kingpin.ParseContext) error {
		standardLogOverridden = true
		return nil
	}).
	Bool()

var probePathOverridden = false
var probePath = kingpin.
	Flag("web.probe-path", "Path under which to expose metrics").
	Default("/probe").
	Action(func(_ *kingpin.ParseContext) error {
		probePathOverridden = true
		return nil
	}).
	String()

var metricsPathOverridden = false
var metricsPath = kingpin.
	Flag("web.metric-path", "Path under which to expose metrics").
	Default("/metrics").
	Action(func(_ *kingpin.ParseContext) error {
		metricsPathOverridden = true
		return nil
	}).
	String()

var listenAddressOverridden = false
var listenAddress = kingpin.
	Flag("web.listen-address", "Address to listen on for web interface and telemetry.").
	Default(":9231").
	Action(func(_ *kingpin.ParseContext) error {
		listenAddressOverridden = true
		return nil
	}).
	String()

var serverTLSCertPathOverridden = false
var serverTLSCertPath = kingpin.
	Flag("server-tls-cert-path", "Server TLS key path").
	Default("").
	Action(func(_ *kingpin.ParseContext) error {
		serverTLSCertPathOverridden = true
		return nil
	}).
	String()

var serverTLSKeyPathOverridden = false
var serverTLSKeyPath = kingpin.
	Flag("server-tls-key-path", "Server TLS key path").
	Default("").
	Action(func(_ *kingpin.ParseContext) error {
		serverTLSKeyPathOverridden = true
		return nil
	}).
	String()

var clientCAPathOverridden = false
var clientCAPath = kingpin.
	Flag("client-ca-path", "Client CA path").
	Default("").
	Action(func(_ *kingpin.ParseContext) error {
		clientCAPathOverridden = true
		return nil
	}).
	String()

var configPath = kingpin.
	Flag("config", "Config file path").
	Default("/etc/cradle_exporter/config.yml").
	String()

func loadConfig() (*cradle.Config, error) {
	config, err := cradle.ReadConfigFromFile(*configPath)
	if err != nil {
		err = fmt.Errorf("failed to read config file: %s, err=%v", *configPath, err)
		return nil, err
	}

	if standardLogOverridden {
		config.Cli.StandardLog = *standardLog
	}
	if probePathOverridden {
		config.Web.ProbePath = *probePath
	}
	if metricsPathOverridden {
		config.Web.MetricPath = *metricsPath
	}
	if listenAddressOverridden {
		config.Web.ListenAddress = *listenAddress
	}
	if serverTLSCertPathOverridden {
		config.Web.ServerTLSCertPath = *serverTLSCertPath
	}
	if serverTLSKeyPathOverridden {
		config.Web.ServerTLSKeyPath = *serverTLSKeyPath
	}
	if clientCAPathOverridden {
		config.Web.ClientCAPath = *clientCAPath
	}
	return config, nil
}

func main() {
	var err error
	var log *zap.Logger

	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cfg, err := loadConfig()
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(-1)
	}

	// Check weather terminal or not
	if cfg.Cli.StandardLog || isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create logger: %v", err)
		os.Exit(-1)
	}
	undo := zap.ReplaceGlobals(log)
	defer undo()
	log.Info("Log System Initialized.")

	cr := cradle.New(cfg)

	if *isConfigCheckMode {
		// Just check and report result
		err = cr.Check(cfg)
		if err != nil {
			log.Fatal("Failed to read config file", zap.Error(err))
		}
		log.Info("Checking config file: OK!")
		return
	}

	done := make(chan bool, 1)
	{
		// Setup signal handling
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for {
				switch <-signals {
				case syscall.SIGINT:
					fallthrough
				case syscall.SIGTERM:
					log.Info("Shutting down...")
					err := cr.Shutdown()
					if err != nil {
						log.Fatal("Failed to shutdown cradle", zap.Error(err))
					}
					done <- true
					return
				case syscall.SIGHUP:
					cfg, err := loadConfig()
					if err != nil {
						log.Warn("Failed to reload config", zap.Error(err))
						break
					}
					err = cr.Reload(cfg)
					if err == nil {
						log.Info("Config reloaded")
					} else {
						log.Warn("Failed to reload config", zap.Error(err))
					}
				}
			}
		}()
	}

	// reload with the same config.
	err = cr.Reload(cfg)
	if err != nil {
		log.Fatal("Failed to reload config", zap.Error(err))
	}
	err = cr.Run()
	if err != nil {
		log.Fatal("Failed to run cradle", zap.Error(err))
	}
	log.Info("All done, thanks!")
}
