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
	data, err := json.Marshal(ErrorResponse{msg})
	if err != nil {
		w.WriteHeader(statusCode)
		io.WriteString(w, fmt.Sprintf("Error while building response: %s", err))
		logger.Errorf("Error while building response: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	io.WriteString(w, string(data))
}

func writeJson(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

// Handle the /heatlh GET HTTP endpoint
func HandlerHealth(w http.ResponseWriter, req *http.Request) {
	code := http.StatusOK
	start := time.Now()
	defer func() { recordMetrics(start, req, code) }()

	// This fuction is frequently used by K8S -> do not fill the logs

	data, err := json.Marshal(HealthResponse{"UP"})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error while marshalling HealthResponse")
	} else {
		writeJson(w, data)
	}
}

// Handle the /hello GET HTTP endpoint
func HandlerHello(w http.ResponseWriter, req *http.Request) {
	code := http.StatusOK
	start := time.Now()
	defer func() { recordMetrics(start, req, code) }()

	logger.Infof("%s request to %s\n", req.Method, req.RequestURI)

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

	h, _ := os.Hostname()
	hello := fmt.Sprintf("Hello, my name is '%s', I'm served from '%s'", env.NAME, h)
	data, err := json.Marshal(MsgResponse{hello})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error while marshalling MsgResponse")
	} else {
		writeJson(w, data)
	}
}

// Handle the /refresh POST HTTP endpoint
func HandlerRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "This method is not allowed")
		return
	}

	code := http.StatusOK
	start := time.Now()
	defer func() { recordMetrics(start, req, code) }()

	logger.Infof("%s request to %s\n", req.Method, req.RequestURI)

	env.Load()

	data, err := json.Marshal(MsgResponse{"Reloaded OK"})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error while marshalling MsgResponse")
	} else {
		writeJson(w, data)
	}
}

// Handle the /remote GET HTTP endpoint
func HandlerRemote(w http.ResponseWriter, req *http.Request) {
	code := http.StatusOK
	start := time.Now()
	defer func() { recordMetrics(start, req, code) }()

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

		var respRemote MsgResponse
		err := json.Unmarshal(bodyBytes, &respRemote)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error while unmarshalling response from remote")
			logger.Errorf("Error while unmarshalling response from remote: %s", err)
		}

		h, _ := os.Hostname()
		respCurrent := MsgRemoteResponse{
			Msg:        fmt.Sprintf("Hello, my name is '%s', I'm served from '%s'", env.NAME, h),
			FromRemote: respRemote,
		}
		data, err := json.Marshal(respCurrent)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Error while marshalling MsgRemoteResponse")
		} else {
			writeJson(w, data)
		}
	} else {
		io.WriteString(w, fmt.Sprintf("Error while calling the back: %d", resp.StatusCode))
	}
}

func recordMetrics(start time.Time, req *http.Request, code int) {
	elapsed := time.Since(start)

	histogram.With(
		prometheus.Labels{
			"service": serviceName,
			"code":    fmt.Sprintf("%d", code),
			"method":  req.Method,
			"path":    req.URL.Path,
		},
	).Observe(elapsed.Seconds())
}
