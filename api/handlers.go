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
)

func kvAsJson(key string, val string) string {
	m := make(map[string]string)
	m[key] = val
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}

func kmAsJson(key string, v map[string]interface{}) string {
	m := make(map[string]interface{})
	m[key] = v
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}

func mapAsJson(m map[string]interface{}) string {
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}

func HandlerHealth(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, kvAsJson("health", "UP"))
}

func HandlerHello(w http.ResponseWriter, r *http.Request) {
	h, _ := os.Hostname()
	m := make(map[string]interface{})
	m["msg"] = fmt.Sprintf("Hello, my name is '%s' (id#%s) and I'm served from '%s'", env.NAME, env.INSTANCE_ID[:8], h)

	// Hidden feature: response with delay -> /hello?delay=valueInMillis
	delay := r.URL.Query().Get("delay")
	if len(delay) > 0 {
		delayNum, _ := strconv.Atoi(delay)
		time.Sleep(time.Duration(delayNum) * time.Millisecond)
	}

	// Hidden feature: response with error -> /hello?error=valueInPercent
	error := r.URL.Query().Get("error")
	if len(error) > 0 {
		errorNum, _ := strconv.Atoi(error)
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(100)
		if n <= errorNum {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, kvAsJson("error", "Something, somewhere, went wrong!"))
			return
		}
	}
	io.WriteString(w, mapAsJson(m))
}

func HandlerRemote(w http.ResponseWriter, r *http.Request) {
	// Build the request
	req, err := http.NewRequest("GET", "http://"+env.REMOTE_URL, nil)
	if err != nil {
		http.Error(w, "Error while building request", http.StatusBadRequest)
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
		http.Error(w, "Error while requesting backend", http.StatusServiceUnavailable)
		logger.Errorf("Error while requesting backend: %s", err)
		return
	}

	// Callers should close resp.Body
	defer resp.Body.Close()

	// Get body as string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			http.Error(w, "Error while getting body", http.StatusInternalServerError)
			logger.Errorf("Error while getting body: %s", err)
			return
		}
		var x map[string]interface{}
		err := json.Unmarshal(bodyBytes, &x)
		if err != nil {
			http.Error(w, "Error while unmarshalling response from remote", http.StatusInternalServerError)
			logger.Errorf("Error while unmarshalling response from remote: %s", err)
		}
		io.WriteString(w, kmAsJson("fromRemote", x))
	} else {
		io.WriteString(w, fmt.Sprintf("Error while calling the back: %d", resp.StatusCode))
	}
}
