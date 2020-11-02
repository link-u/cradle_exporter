# cradle_exporter - Cradle for Prometheus Exporters

[![Go Report Card](https://goreportcard.com/badge/github.com/link-u/cradle_exporter)](https://goreportcard.com/report/github.com/link-u/cradle_exporter)

[![Build on Linux](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Linux/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Linux%22)
[![Build on macOS](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20macOS/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+macOS%22)
[![Build on Windows](https://github.com/link-u/cradle_exporter/workflows/Build%20on%20Windows/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+on+Windows%22)  
[![Publish Docker image](https://github.com/link-u/cradle_exporter/workflows/Publish%20Docker%20image/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Publish+Docker+image%22)
[![Build debian packages](https://github.com/link-u/cradle_exporter/workflows/Build%20debian%20packages/badge.svg)](https://github.com/link-u/cradle_exporter/actions?query=workflow%3A%22Build+debian+packages%22)

`cradle_exporter` gathers output from other exporters into one endpoint, `/probe`.

Support these modes:

 - `service` - Daemonize (supervise) other exporter binary and scrape endpoints.
 - `exporter` - Just scrape other exporters and gather results (do not supervise them).
 - `script` - Execute a given shell script and expose a result.
 - `cron` - like `script` mode, but execute a script periodically.
 - `static` - Read static files from given paths.

# How to use?

## docker-compose.yml

```yaml
  cradle-exporter:
    image: ghcr.io/link-u/cradle_exporter
    expose:
       - 9231
    command:
       - '--config=/etc/cradle_exporter/config.yml'
    volume:
       # See example/config/config.yml as an example.
       - 'example/config:/etc/cradle_exporter'
       - 'example/script:/some/path/to/store/scripts'
    restart: always
```

## Prometheus config

```yaml
  - job_name: 'cradle_exporter'
    metrics_path: '/probe'
    static_configs:
      - targets:
        - 'http://host_to_nodes:port/'
        - 'https://host_to_nodes:port/'
```

# How to configure `cradle_exporter`

## Main config file (Given by `--config=<name>.yml`)

```yaml
---
include_dirs:
  - './example/config/conf.d'
cli:
  standard_log: true  # can be overridden by --cli.standard-log argument
web:
  listen_address: ':9231' # can be overridden by --web.listen-address argument
  metric_path:    '/metrics' # can be overridden by --web.metric-path argument
  probe_path:     '/probe' # can be overridden by --web.probe-path argument
```

It reads all files in `/etc/cradle_exporter/conf.d` as a target config.

Please see below:

### Service Target example config

`cradle_exporter` supervise `other_exporter` and proxy the endpoint of the exporter, `http://localhost:9222/metrics`, in `/collected`.

```yaml
---
service:
  path: '/path/to/other_exporter'
  args:
    - '--config=....'
    - '--web.listen-address=:9222'
  endpoints:
    - 'http://localhost:9222/metrics'
```

### Exporter Target example config

`cradle_exporter` just reads `http://localhost:9222/metrics` and expose it.

```yaml
---
exporter:
  endpoints:
    - 'http://localhost:9222/metrics'
```

### Script Target example config

`cradle_exporter` executes `/path/to/script.sh` on-the-fly and expose the execution result.

```yaml
---
script:
  path: '/path/to/script.sh'
  args:
    - 'arg1'
    - 'arg2'
    - '...'
```

### Cron Target example file

`cradle_exporter` executes `/path/to/script.sh` periodically and expose the result.

```yaml
---
cron:
  path: '/path/to/script.sh'
  args:
    - 'arg1'
    - 'arg2'
    - '...'
  # Execute every 10 second.
  every: "*/10 * * * * *" # See https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format
```

### Static Target example file

`cradle_exporter` reads all files in given paths recursively. Path can be both file or dir.

```yaml
---
static:
  paths:
    - '/path/to/static_files_dir' # files in a dir, recursively
    - '/path/to/static_file' # a file
```

# License

MIT
