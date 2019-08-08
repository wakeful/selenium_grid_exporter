package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const (
	nameSpace = "selenium_grid"
	subSystem = "hub"
)

var (
	listenAddress = flag.String("listen-address", ":8080", "Address on which to expose metrics.")
	metricsPath   = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
	scrapeURI     = flag.String("scrape-uri", "http://grid.local", "URI on which to scrape Selenium Grid.")
)

type Exporter struct {
	URI                                               string
	mutex                                             sync.RWMutex
	up, slotsTotal, slotsFree, newSessionRequestCount prometheus.Gauge
}

type hubResponse struct {
	Success      bool       `json:"success"`
	Debug        bool       `json:"debug"`
	CleanUpCycle int        `json:"cleanUpCycle"`
	Slots        slotCounts `json:"slotCounts"`
	NewSession   float64    `json:"newSessionRequestCount"`
}

type slotCounts struct {
	Free  float64 `json:"free"`
	Total float64 `json:"total"`
}

func NewExporter(uri string) *Exporter {
	log.Infoln("Collecting data from:", uri)

	return &Exporter{
		URI: uri,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Name:      "up",
			Help:      "was the last scrape of Selenium Grid successful.",
		}),
		slotsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "slotsTotal",
			Help:      "total number of slots",
		}),
		slotsFree: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "slotsFree",
			Help:      "number of free slots",
		}),
		newSessionRequestCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "sessions_backlog",
			Help:      "number of sessions waiting for a slot",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.slotsTotal.Describe(ch)
	e.slotsFree.Describe(ch)
	e.newSessionRequestCount.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.scrape()

	ch <- e.up
	ch <- e.slotsTotal
	ch <- e.slotsFree
	ch <- e.newSessionRequestCount

	return
}

func (e *Exporter) scrape() {

	e.slotsTotal.Set(0)
	e.slotsFree.Set(0)

	body, err := e.fetch()
	if err != nil {
		e.up.Set(0)

		log.Errorf("Can't scrape Selenium Grid: %v", err)
		return
	}

	e.up.Set(1)

	var hResponse hubResponse
	if err := json.Unmarshal(body, &hResponse); err != nil {

		log.Errorf("Can't decode Selenium Grid response: %v", err)
		return
	}

	e.slotsTotal.Set(hResponse.Slots.Total)
	e.slotsFree.Set(hResponse.Slots.Free)
	e.newSessionRequestCount.Set(hResponse.NewSession)

}

func (e Exporter) fetch() (output []byte, err error) {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	response, err := client.Get(e.URI + "/grid/api/hub")
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	output, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return
}

func main() {
	flag.Parse()

	log.Infoln("Starting selenium_grid_exporter")

	prometheus.MustRegister(NewExporter(*scrapeURI))
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
