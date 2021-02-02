# Selenium Grid exporter

A [Prometheus](https://prometheus.io/) exporter that collects [Selenium Grid](http://www.seleniumhq.org/projects/grid/) metrics.

### Usage

```sh
$ docker run -it wakeful/selenium-grid-exporter -h
Usage of /selenium_grid_exporter:
  -listen-address string
      Address on which to expose metrics. (default ":8080")
  -scrape-uri string
      URI on which to scrape Selenium Grid. (default "http://grid.local")
  -telemetry-path string
      Path under which to expose metrics. (default "/metrics")
```

## Metrics

```
# HELP selenium_grid_hub_sessionCount number of active sessions
# TYPE selenium_grid_hub_sessionCount gauge
selenium_grid_hub_sessionCount 0
# HELP selenium_grid_hub_maxSession number of max sessions
# TYPE sselenium_grid_hub_maxSession gauge
selenium_grid_hub_maxSession 0
# HELP selenium_grid_hub_totalSlots total number of slots
# TYPE selenium_grid_hub_totalSlots gauge
selenium_grid_hub_totalSlots 8
# HELP selenium_grid_hub_sessionQueueSize number of session in queue
# TYPE selenium_grid_hub_sessionQueueSize gauge
selenium_grid_hub_sessionQueueSize 0
# HELP selenium_grid_up was the last scrape of Selenium Grid successful.
# TYPE selenium_grid_up gauge
selenium_grid_up 1
```
