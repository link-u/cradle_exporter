package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/link-u/cradle_exporter/internal/cradle"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
)

var (
	configPath    = flag.String("config", "/etc/cradle_exporter/config.yml", "Config file path")
	collectedPath = flag.String("web.collected-path", "/collected", "Path under which to expose metrics")
	metricsPath   = flag.String("web.metric-path", "/metrics", "Path under which to expose metrics")
	listenAddress = flag.String("web.listen-address", ":9231", "Address to listen on for web interface and telemetry.")
)

func main() {
	var err error
	var log *zap.Logger
	flag.Parse()

	// Check is terminal
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
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
	log.Info("Loading config file", zap.String("path", *configPath))
	config, err := cradle.ReadConfigFromFile(*configPath)
	if err != nil {
		log.Fatal("Failed to read config file", zap.Error(err))
	}

	appConfig := &cradle.AppConfig{
		MetricPath:     *metricsPath,
		CollectedPath:  *collectedPath,
		HttpListenAddr: *listenAddress,
	}
	cr, err := cradle.New(appConfig, config)
	if err != nil {
		log.Fatal("Failed to create cradle", zap.Error(err))
	}

	err = cr.Run()
	if err != nil {
		log.Fatal("Failed to run cradle", zap.Error(err))
	}
}
