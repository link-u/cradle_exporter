package main

import (
	"fmt"
	"os"

	"github.com/link-u/cradle_exporter/internal/cradle"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const version = "v1.0.0"

func main() {
	var err error
	var log *zap.Logger
	configPath := kingpin.Flag("config", "Config file path").
		Default("/etc/cradle_exporter/config.yml").String()

	standardLogOverridden := false
	standardLog := kingpin.Flag("cli.standard-log", "Print logs in standard format, not in json").
		Default("false").
		Action(func(_ *kingpin.ParseContext) error {
			standardLogOverridden = true
			return nil
		}).Bool()

	collectedPathOverridden := false
	collectedPath := kingpin.Flag("web.collected-path", "Path under which to expose metrics").
		Default("/collected").
		Action(func(_ *kingpin.ParseContext) error {
			collectedPathOverridden = true
			return nil
		}).String()

	metricsPathOverridden := false
	metricsPath := kingpin.Flag("web.metric-path", "Path under which to expose metrics").
		Default("/metrics").
		Action(func(_ *kingpin.ParseContext) error {
			metricsPathOverridden = true
			return nil
		}).String()

	listenAddressOverridden := false
	listenAddress := kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").
		Default(":9231").
		Action(func(_ *kingpin.ParseContext) error {
			listenAddressOverridden = true
			return nil
		}).
		String()
	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

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

	cr, err := cradle.New(config)
	if err != nil {
		log.Fatal("Failed to create cradle", zap.Error(err))
	}

	err = cr.Run()
	if err != nil {
		log.Fatal("Failed to run cradle", zap.Error(err))
	}
}
