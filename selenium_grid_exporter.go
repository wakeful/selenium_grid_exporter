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
	URI                                                        string
	mutex                                                      sync.RWMutex
	up, totalSlots, maxSession, sessionCount, sessionQueueSize prometheus.Gauge
}

type hubResponse struct {
	Data struct {
		Grid struct {
			TotalSlots       float64 `json:"totalSlots"`
			MaxSession       float64 `json:"maxSession"`
			SessionCount     float64 `json:"sessionCount"`
			SessionQueueSize float64 `json:"sessionQueueSize"`
		} `json:"grid"`
	} `json:"data"`
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
			Help:      "total number of usedSlots",
		}),
		maxSession: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "maxSession",
			Help:      "maximum number of sessions",
		}),
		sessionCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "sessionCount",
			Help:      "number of active sessions",
		}),
		sessionQueueSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: nameSpace,
			Subsystem: subSystem,
			Name:      "sessionQueueSize",
			Help:      "number of queued sessions",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.totalSlots.Describe(ch)
	e.maxSession.Describe(ch)
	e.sessionCount.Describe(ch)
	e.sessionQueueSize.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.scrape()

	ch <- e.up
	ch <- e.totalSlots
	ch <- e.maxSession
	ch <- e.sessionCount
	ch <- e.sessionQueueSize

	return
}

func (e *Exporter) scrape() {

	e.totalSlots.Set(0)
	e.maxSession.Set(0)
	e.sessionCount.Set(0)
	e.sessionQueueSize.Set(0)

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
	e.maxSession.Set(hResponse.Data.Grid.MaxSession)
	e.sessionCount.Set(hResponse.Data.Grid.SessionCount)
	e.sessionQueueSize.Set(hResponse.Data.Grid.SessionQueueSize)
}

func (e Exporter) fetch() (output []byte, err error) {

	url := (e.URI + "/graphql")
	method := "POST"

	payload := strings.NewReader(`{
		"query": "{ grid {totalSlots, maxSession, sessionCount, sessionQueueSize} }"
	}`)

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

	//s := string(body)
	//fmt.Println(s)
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
