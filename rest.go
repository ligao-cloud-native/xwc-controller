package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/ligao-cloud-native/xwc-controller/pkg/callback"
	"github.com/ligao-cloud-native/xwc-controller/pkg/metrics"
	"github.com/ligao-cloud-native/xwc-controller/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

type Options struct {
	Server *http.Server
	Metric *metrics.Metrics
}

func StartHttpServer(opt *Options) {
	router := mux.NewRouter()
	router.HandleFunc("/metrics", promhttp.Handler().(http.HandlerFunc)).Methods("GET")
	router.HandleFunc("/task-completion-callback", opt.handleCallback).Methods("POST")

	opt.Server = &http.Server{
		Addr:           ":7000",
		Handler:        router,
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	klog.Fatal(opt.Server.ListenAndServe())
}

func (opt *Options) handleCallback(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Error(err)
		return
	}

	jobCallback := callback.JobCallback{}
	if err := json.Unmarshal(reqBody, &jobCallback); err != nil {
		klog.Error(err)
		return
	}

	//TODO: UPDATE pwc status
	updatePWCStatus()

	opt.Metric.HandleMetrics(&jobCallback)
}

func updatePWCStatus() {
	xwcClient, _ := utils.XwcClient("")

}
