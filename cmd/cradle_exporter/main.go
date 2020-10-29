package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/link-u/cradle_exporter/internal/cradle"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const version = "v1.0.0"

/*****************************************************************************
  Flags
 *****************************************************************************/

var standardLogOverridden = false
var standardLog = kingpin.
	Flag("cli.standard-log", "Print logs in standard format, not in json").
	Default("false").
	Action(func(_ *kingpin.ParseContext) error {
		standardLogOverridden = true
		return nil
	}).Bool()

var collectedPathOverridden = false
var collectedPath = kingpin.
	Flag("web.collected-path", "Path under which to expose metrics").
	Default("/collected").
	Action(func(_ *kingpin.ParseContext) error {
		collectedPathOverridden = true
		return nil
	}).String()

var metricsPathOverridden = false
var metricsPath = kingpin.
	Flag("web.metric-path", "Path under which to expose metrics").
	Default("/metrics").
	Action(func(_ *kingpin.ParseContext) error {
		metricsPathOverridden = true
		return nil
	}).String()

var listenAddressOverridden = false
var listenAddress = kingpin.
	Flag("web.listen-address", "Address to listen on for web interface and telemetry.").
	Default(":9231").
	Action(func(_ *kingpin.ParseContext) error {
		listenAddressOverridden = true
		return nil
	}).
	String()

var configPath = kingpin.
	Flag("config", "Config file path").
	Default("/etc/cradle_exporter/config.yml").String()

func loadConfig() *cradle.Config {
	config, err := cradle.ReadConfigFromFile(*configPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to read config file: %s, err=%v", *configPath, err)
		os.Exit(-1)
	}

	if standardLogOverridden {
		config.Cli.StandardLog = *standardLog
	}

	if collectedPathOverridden {
		config.Web.CollectedPath = *collectedPath
	}

	if metricsPathOverridden {
		config.Web.MetricPath = *metricsPath
	}
	if listenAddressOverridden {
		config.Web.ListenAddress = *listenAddress
	}
	return config
}

func main() {
	var err error
	var log *zap.Logger

	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	config := loadConfig()

	// Check weather terminal or not
	if config.Cli.StandardLog || isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
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

	cr := cradle.New(config)

	{
		// Setup signal handling
		signals := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			switch <-signals {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				log.Info("Shutting down...")
				err := cr.Shutdown(context.Background())
				if err != nil {
					log.Fatal("Failed to shutdown cradle", zap.Error(err))
				}
			case syscall.SIGHUP:
				err := cr.Reload(loadConfig())
				if err == nil {
					log.Info("Config reloaded")
				} else {
					log.Warn("Failed to reload config", zap.Error(err))
				}
			}
			done <- true
		}()
	}

	// reload with the same config.
	err = cr.Reload(config)
	if err != nil {
		log.Fatal("Failed to reload config", zap.Error(err))
	}
	err = cr.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to run cradle", zap.Error(err))
	}
	log.Info("All done, thanks!")
}
