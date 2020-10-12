# cradle_exporter - Cradle for Prometheus Exporters

[![Build on Linux](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Linux/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Linux%22)
[![Build on macOS](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20macOS/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+macOS%22)
[![Build on Windows](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Windows/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Windows%22)  
[![Publish Docker image](https://github.com/link-u/cradle_exporter/workflows/Publish%20Docker%20image/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Publish+Docker+image%22)

`cradle_exporter` do:

 - Daemonizes multiple exporters
   - Executes with given settings.
   - Restarts when the executed process dies.
 - Executes given scripts peridolically
   - Supports cron-notations.
   - Cache the execution results.
 - Serve static files in specified directory.

`cradle_exporter` exposes just only one entrypoint, `/probe`, which gathers all metrics collected from all exporters, scripts and files.

# How to use?

## docker-compose.yml

```yaml
  mrtg-exporter:
    image: ghcr.io/link-u/cradle_exporter
    expose:
      - 9231
    restart: always
```

## Prometheus config

```yaml
  - job_name: 'cradles'
    metrics_path: '/probe'
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: '<cradle_exporter>:9230'
    static_configs:
      - targets:
        - 'http://host_to_nodes:port/'
        - 'https://host_to_nodes:port/'
```

# License

MIT
