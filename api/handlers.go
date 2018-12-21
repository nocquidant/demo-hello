package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/logger"
	"github.com/nocquidant/demo-hello/env"
	"github.com/prometheus/client_golang/prometheus"
)

var serviceName = "demo-hello"

var (
	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: "http_server",
		Name:      "resp_time",
		Help:      "Request response time",
	}, []string{
		"service",
		"code",
		"method",
		"path",
	})
)

func init() {
	prometheus.MustRegister(histogram)
}

func writeError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	io.WriteString(w, kvAsJson("error", msg))
}

// Handle the /heatlh GET HTTP endpoint
func HandlerHealth(w http.ResponseWriter, req *http.Request) {
	defer func() { recordMetrics(time.Now(), req, http.StatusOK) }()
	// This fuction is frequently used by K8S -> do not fill the logs

	io.WriteString(w, kvAsJson("health", "UP"))
}

// Handle the /hello GET HTTP endpoint
func HandlerHello(w http.ResponseWriter, req *http.Request) {
	defer func() { recordMetrics(time.Now(), req, http.StatusOK) }()
	logger.Infof("%s request to %s\n", req.Method, req.RequestURI)

	h, _ := os.Hostname()
	m := make(map[string]interface{})
	m["msg"] = fmt.Sprintf("Hello, my name is '%s', I'm served from '%s'", env.NAME, h)

	// Hidden feature: response with delay -> /hello?delay=valueInMillis
	delay := req.URL.Query().Get("delay")
	if len(delay) > 0 {
		delayNum, _ := strconv.Atoi(delay)
		time.Sleep(time.Duration(delayNum) * time.Millisecond)
	}

	// Hidden feature: response with error -> /hello?error=valueInPercent
	error := req.URL.Query().Get("error")
	if len(error) > 0 {
		errorNum, _ := strconv.Atoi(error)
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(100)
		if n <= errorNum {
			writeError(w, http.StatusInternalServerError, "Something, somewhere, went wrong!")
			return
		}
	}
	io.WriteString(w, mapAsJson(m))
}

// Handle the /refresh POST HTTP endpoint
func HandlerRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "This method is not allowed")
		return
	}

	defer func() { recordMetrics(time.Now(), req, http.StatusOK) }()
	logger.Infof("%s request to %s\n", req.Method, req.RequestURI)

	env.Load()

	io.WriteString(w, kvAsJson("msg", "Reloaded OK"))
}

// Handle the /remote GET HTTP endpoint
func HandlerRemote(w http.ResponseWriter, req *http.Request) {
	defer func() { recordMetrics(time.Now(), req, http.StatusOK) }()
	logger.Infof("%s request to %s\n", req.Method, req.RequestURI)

	// Build the request
	req, err := http.NewRequest("GET", "http://"+env.REMOTE_URL, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Error while building request")
		logger.Errorf("Error while building request: %s", err)
		return
	}

	// A Client is an HTTP client
	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	// Send the request via a client
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Error while requesting backend")
		logger.Errorf("Error while requesting backend: %s", err)
		return
	}

	// Callers should close resp.Body
	defer resp.Body.Close()

	// Get body as string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			writeError(w, http.StatusInternalServerError, "Error while getting body from remote")
			logger.Errorf("Error while getting body: %s", err)
			return
		}
		var x map[string]interface{}
		err := json.Unmarshal(bodyBytes, &x)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error while unmarshalling response from remote")
			logger.Errorf("Error while unmarshalling response from remote: %s", err)
		}
		h, _ := os.Hostname()
		msg := fmt.Sprintf("Hello, my name is '%s', I'm served from '%s'", env.NAME, h)
		io.WriteString(w, kmAsJson("msg", msg, "fromRemote", x))
	} else {
		io.WriteString(w, fmt.Sprintf("Error while calling the back: %d", resp.StatusCode))
	}
}

func recordMetrics(start time.Time, req *http.Request, code int) {
	duration := time.Since(start)
	histogram.With(
		prometheus.Labels{
			"service": serviceName,
			"code":    fmt.Sprintf("%d", code),
			"method":  req.Method,
			"path":    req.URL.Path,
		},
	).Observe(duration.Seconds())
}
