# CH7465LG

[![GoDoc](https://godoc.org/github.com/dstotijn/ch7465lg?status.svg)](http://godoc.org/github.com/dstotijn/ch7465lg)
[![Go Report Card](https://goreportcard.com/badge/github.com/dstotijn/ch7465lg)](https://goreportcard.com/report/github.com/dstotijn/ch7465lg)

Go library and Prometheus exporter for the Compal CH7465LG cable modem (Ziggo
ConnectBox).

## Features

- Metrics for bonded downstream channels

## TODO

- Expose metrics for bonded upstream channels

## Installation

### Build from source

#### Prerequisites

- [Go](https://golang.org/)

```
$ GO111MODULE=on go get github.com/dstotijn/ch7465lg/cmd/ch7465lg
```

## Usage

```

$ ch7465lg -h
Usage of ./ch7465lg:
-gw string
Modem gateway IP address. (default "192.168.178.1")
-prom string
Prometheus exporter bind address. (default ":9810")
```

The password for the modem admin user is read from a (required) `MODEM_PASSWORD`
environment variable.

## Acknowledgements

- [compal_CH7465LG_py](https://github.com/ties/compal_CH7465LG_py/): Helped me
  figure out that the order of `POST` form values matters.

## License

[MIT](LICENSE)

---

© 2020 David Stotijn — [Twitter](https://twitter.com/dstotijn), [Email](mailto:dstotijn@gmail.com)
