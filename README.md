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
# HELP selenium_grid_hub_sessions_backlog number of sessions waiting for a slot
# TYPE selenium_grid_hub_sessions_backlog gauge
selenium_grid_hub_sessions_backlog 0
# HELP selenium_grid_hub_slotsFree number of free slots
# TYPE selenium_grid_hub_slotsFree gauge
selenium_grid_hub_slotsFree 4
# HELP selenium_grid_hub_slotsTotal total number of slots
# TYPE selenium_grid_hub_slotsTotal gauge
selenium_grid_hub_slotsTotal 8
# HELP selenium_grid_up was the last scrape of Selenium Grid successful.
# TYPE selenium_grid_up gauge
selenium_grid_up 1
```
