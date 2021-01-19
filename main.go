package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "guacamole"
const guacamoleTokenAPI = "/api/tokens"
const guacamoleActiveConnectionsAPI = "/api/session/data/%s/activeConnections"
const guacamoleUsersAPI = "/api/session/data/%s/users"
const guacamoleConnectionHistoryAPI = "/api/session/data/%s/history/connections"

type guacamoleTokenMap struct {
	AuthToken            string   `json:"authToken"`
	Username             string   `json:"username"`
	DataSource           string   `json:"dataSource"`
	AvailableDataSources []string `json:"availableDataSources"`
}

type exporter struct {
	guacamoleEndpoint, guacamolehUsername, guacamolePassword, guacamoleDataSource string
}

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client = &http.Client{Transport: tr}

	listenAddress = flag.String("web.listen-address", ":9623",
		"Address to listen on for telemetry")
	metricsPath = flag.String("web.telemetry-path", "/metrics",
		"Path under which to expose metrics")

	version = flag.Bool("version", false, "Show version information")

	// Metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last Guacamole query successful.",
		nil, nil,
	)

	connectionHistory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "connection_history_total"),
		"The total number of established connections",
		nil, nil,
	)

	users = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "number_of_users"),
		"The current number of registered users",
		nil, nil,
	)

	activeConnections = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "number_of_active_connections"),
		"The current number of active connections",
		nil, nil,
	)
)

func newExporter(guacamoleEndpoint string, guacamoleUsername string, guacamolePassword string, guacamoleDataSource string) *exporter {
	return &exporter{
		guacamoleEndpoint:   guacamoleEndpoint,
		guacamolehUsername:  guacamoleUsername,
		guacamolePassword:   guacamolePassword,
		guacamoleDataSource: guacamoleDataSource,
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- connectionHistory
	ch <- users
	ch <- activeConnections
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	token, err := getToken(e.guacamoleEndpoint, e.guacamolehUsername, e.guacamolePassword)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Println(err)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)

	counterConnectionHistory, err := getConnectionHistory(e.guacamoleEndpoint, token, e.guacamoleDataSource)
	if err != nil {
		log.Println("Could not get counterConnectionHistory", err)
	} else {
		ch <- prometheus.MustNewConstMetric(connectionHistory, prometheus.CounterValue, float64(counterConnectionHistory))
	}

	gaugeUsers, err := getUsers(e.guacamoleEndpoint, token, e.guacamoleDataSource)
	if err != nil {
		log.Println("Could not get gaugeUsers", err)
	} else {
		ch <- prometheus.MustNewConstMetric(users, prometheus.GaugeValue, float64(gaugeUsers))
	}

	gaugeActiveConnections, err := getActiveConnections(e.guacamoleEndpoint, token, e.guacamoleDataSource)
	if err != nil {
		log.Println("Could not get gaugeActiveConnections", err)
	} else {
		ch <- prometheus.MustNewConstMetric(activeConnections, prometheus.GaugeValue, float64(gaugeActiveConnections))
	}

	releaseToken(e.guacamoleEndpoint, token)
}

func getBody(token string, _url string) ([]byte, error) {
	//log.Printf("%v", _url)

	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//log.Printf("%v", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getConnectionHistory(endpoint string, token string, dataSource string) (int, error) {
	_url := endpoint + fmt.Sprintf(guacamoleConnectionHistoryAPI, dataSource)

	body, err := getBody(token, _url)
	if err != nil {
		return -1, err
	}

	var raw []interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return -1, err
	}

	return len(raw), nil
}

func getUsers(endpoint string, token string, dataSource string) (int, error) {
	_url := endpoint + fmt.Sprintf(guacamoleUsersAPI, dataSource)

	body, err := getBody(token, _url)
	if err != nil {
		return -1, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return -1, err
	}

	return len(raw), nil
}

func getActiveConnections(endpoint string, token string, dataSource string) (int, error) {
	_url := endpoint + fmt.Sprintf(guacamoleActiveConnectionsAPI, dataSource)

	body, err := getBody(token, _url)
	if err != nil {
		return -1, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return -1, err
	}

	return len(raw), nil
}

func getToken(endpoint string, username string, password string) (token string, err error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	req, err := http.NewRequest(http.MethodPost, endpoint+guacamoleTokenAPI, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	var guacamoleTokenMap guacamoleTokenMap
	err = json.Unmarshal(body, &guacamoleTokenMap)
	if err != nil {
		return "", err
	}

	return guacamoleTokenMap.AuthToken, nil
}

func releaseToken(endpoint string, token string) {
	req, err := http.NewRequest(http.MethodDelete, endpoint+guacamoleTokenAPI+"/"+token, nil)
	if err != nil {
		log.Println("releaseToken NewRequest failed", err)
		return
	}

	_, err = client.Do(req)
	if err != nil {
		log.Println("releaseToken Do failed", err)
		return
	}
	//log.Println("token released!")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assume env variables are set.")
	}

	flag.Parse()

	if *version {
		println("guacamole_exporter, version 0.1.1")
		return
	}

	guacamoleEndpoint := os.Getenv("GUACAMOLE_ENDPOINT")
	guacamoleUsername := os.Getenv("GUACAMOLE_USERNAME")
	guacamolePassword := os.Getenv("GUACAMOLE_PASSWORD")
	guacamoleDataSource := os.Getenv("GUACAMOLE_DATASOURCE")

	exporter := newExporter(guacamoleEndpoint, guacamoleUsername, guacamolePassword, guacamoleDataSource)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Guacamole Exporter</title></head>
             <body>
             <h1>Guacamole Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
