# cradle_exporter - Cradle for Prometheus Exporters

[![Build on Linux](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Linux/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Linux%22)
[![Build on macOS](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20macOS/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+macOS%22)
[![Build on Windows](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Windows/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Windows%22)  
[![Publish Docker image](https://github.com/link-u/cradle_exporter/workflows/Publish%20Docker%20image/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Publish+Docker+image%22)
[![Build debian packages](https://github.com/link-u/cradle_exporter/workflows/Build%20debian%20packages/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+debian+packages%22)

`cradle_exporter` do:

 - Daemonize multiple exporters
   - Executes exporters with given args.
   - Restarts when the executed process dies.
 - Executes given scripts periodically.
   - Supports cron-notations.
   - Cache the execution results.
 - Serve static files in specified directory.

`cradle_exporter` exposes just only one entrypoint, `/probe`(which is configurable), which gathers all metrics from all exporters, scripts and files to make it easier to collect metrics from multiple servers.

# How to use?

## docker-compose.yml

```yaml
  cradle-exporter:
    image: ghcr.io/link-u/cradle_exporter
    expose:
      - 9231
    restart: always
```

## Prometheus config

```yaml
  - job_name: 'cradle_exporter'
    metrics_path: '/collected'
    static_configs:
      - targets:
        - 'http://host_to_nodes:port/'
        - 'https://host_to_nodes:port/'
```

# License

MIT
