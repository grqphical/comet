# comet - A reverse proxy made with Golang

[![Go Tests](https://github.com/grqphical/comet/actions/workflows/go.yml/badge.svg)](https://github.com/grqphical/comet/actions/workflows/go.yml)

Comet is a highly customizable reverse proxy that allows you to proxy requests and host static files.

## Installation

Run:

```bash
$ go install github.com/grqphical/comet@latest
```

## Usage

Create a `comet.toml` file in the directory of your application and setup your configuration. Below is an example configuration, you can learn more about the configuration in this repo's wiki

```toml
proxy_address = "127.0.0.1:5000"
log_requests = true
health_check_interval = 5

[[backend]]
type = "proxy"
route_filter = "/*"
strip_filter = true
address = "http://127.0.0.1:8000"
health_endpoint = "/"
check_health = false
hidden_routes = []

[[backend]]
type = "staticfs"
route_filter = "/static/*"
directory = "/foo/bar"


[ip_filter]
blacklist = ["1.2.3.4"]

[logging]
level = "info"
output = "stderr"

[cors]
allowed_origins = ["*"]
allowed_methods = ["GET", "POST", "PUT"]
allowed_headers = ["Content-Type", "Authorization"]
```

Then run:

```bash
$ comet
```

And the app should run

## License

Comet is licensed under the Mozilla Public License
