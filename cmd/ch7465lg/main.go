package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dstotijn/ch7465lg"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dsSignalPower = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ch7465lg",
			Subsystem: "ds",
			Name:      "power_dbmv",
			Help:      "Power (dBmV) of downstream channel.",
		},
		[]string{"channel_id"},
	)
	dsSNR = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ch7465lg",
			Subsystem: "ds",
			Name:      "snr",
			Help:      "Signal-to-noise ratio (SNR) of a downstream channel.",
		},
		[]string{"channel_id"},
	)
	dsRxMER = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ch7465lg",
			Subsystem: "ds",
			Name:      "rxmer",
			Help:      "Receive modulation error ratio (RxMER) of a downstream channel.",
		},
		[]string{"channel_id"},
	)
)

func init() {
	prometheus.MustRegister(dsSignalPower, dsSNR, dsRxMER)
}

func main() {
	var gatewayAddr, promAddr string

	flag.StringVar(&gatewayAddr, "gw", "192.168.178.1", "Modem gateway IP address.")
	flag.StringVar(&promAddr, "prom", ":9810", "Prometheus exporter bind address.")
	flag.Parse()

	password := os.Getenv("MODEM_PASSWORD")
	if password == "" {
		log.Fatal("[ERROR] Environment variable `MODEM_PASSWORD` is required.")
	}

	client, err := ch7465lg.NewClient(gatewayAddr, password, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			if err := recordMetrics(client); err != nil {
				log.Printf("[ERROR] Could not record metrics: %v", err)
			}
			time.Sleep(30 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	if host, port, err := net.SplitHostPort(promAddr); err == nil {
		if host == "" {
			host = "localhost"
		}
		log.Printf("[INFO] Started Prometheus exporter (http://%v:%v/metrics).", host, port)
	}
	log.Fatal(http.ListenAndServe(promAddr, nil))
}

func recordMetrics(client *ch7465lg.Client) error {
	if err := client.Login(); err != nil {
		return err
	}

	downstreams, err := client.Downstreams()
	if err != nil {
		return err
	}

	for _, ds := range downstreams {
		labels := prometheus.Labels{"channel_id": strconv.Itoa(ds.ChannelID)}
		dsSignalPower.With(labels).Set(float64(ds.Power))
		dsSNR.With(labels).Set(float64(ds.SNR))
		dsRxMER.With(labels).Set(ds.RxMER)
	}

	log.Printf("[DEBUG] Fetched and recorded metrics for %v downstream channel(s).", len(downstreams))

	// TODO: Fetch upstream channel data and record metrics.

	if err := client.Logout(); err != nil {
		return err
	}

	return nil
}
