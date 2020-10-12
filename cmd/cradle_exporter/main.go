package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/mattn/go-isatty"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	probePath     = flag.String("web.probe-path", "/probe", "Path under which to expose metrics")
	metricsPath   = flag.String("web.metric-path", "/metric", "Path under which to expose metrics")
	listenAddress = flag.String("web.listen-address", ":9230", "Address to listen on for web interface and telemetry.")
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(
		[]byte(fmt.Sprintf(`
<html>
	<head><title>MRTG Exporter</title></head>
	<body>
		<h1>MRTG Exporter</h1>
		<p><a href="%s">Metrics</a></p>
	</body>
</html>
`, *metricsPath)))
}

var whiteSpace = regexp.MustCompile(`\s+`)

func probeHandler(w http.ResponseWriter, r *http.Request) {
}

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

	http.HandleFunc("/", indexHandler)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc(*probePath, probeHandler)

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal("Faled to run web server", zap.Error(err))
	}
}
