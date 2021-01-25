package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	URI                                     string
	mutex                                   sync.RWMutex
	up, totalSlots, usedSlots, sessionCount prometheus.Gauge
}

type hubResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	Grid Grid `json:"grid"`
}

type Grid struct {
	TotalSlots   float64 `json:"totalSlots"`
	UsedSlots    float64 `json:"usedSlots"`
	SessionCount float64 `json:"sessionCount"`
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
		totalSlots: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "totalSlots",
			Help:      "total number of slots",
		}),
		usedSlots: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "usedSlots",
			Help:      "number of used slots",
		}),
		sessionCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "sessionCount",
			Help:      "number of active sessions",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.totalSlots.Describe(ch)
	e.usedSlots.Describe(ch)
	e.sessionCount.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.scrape()

	ch <- e.up
	ch <- e.totalSlots
	ch <- e.usedSlots
	ch <- e.sessionCount

	return
}

func (e *Exporter) scrape() {

	e.totalSlots.Set(0)
	e.usedSlots.Set(0)
	e.sessionCount.Set(0)

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
	e.totalSlots.Set(hResponse.Data.Grid.TotalSlots)
	e.usedSlots.Set(hResponse.Data.Grid.UsedSlots)
	e.sessionCount.Set(hResponse.Data.Grid.SessionCount)
}

func (e Exporter) fetch() (output []byte, err error) {

	url := (e.URI + "/graphql")
	method := "POST"

	payload := strings.NewReader(`{"query":"{ grid {totalSlots, usedSlots , sessionCount } }"}`)

	client := http.Client{
		Timeout: 3 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	return body, err
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
