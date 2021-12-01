package rest

import (
	"context"
	"fmt"
	"net/http"
	"raccoon/collection"
	"raccoon/config"
	"raccoon/http/rest/websocket"
	"raccoon/http/rest/websocket/connection"
	"raccoon/metrics"
	"time"

	"github.com/gorilla/mux"
)

type Service struct {
	Buffer chan *collection.CollectRequest
	s      *http.Server
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func reportConnectionMetrics(conn connection.Table) {
	t := time.Tick(config.MetricStatsd.FlushPeriodMs)
	for {
		<-t
		for k, v := range conn.TotalConnectionPerGroup() {
			metrics.Gauge("connections_count_current", v, fmt.Sprintf("conn_group=%s", k))
		}
	}
}

func (s Service) Init() error {
	collector := collection.NewChannelCollector(s.Buffer)

	pingChannel := make(chan connection.Conn, config.ServerWs.ServerMaxConn)
	wh := websocket.NewHandler(pingChannel)
	go websocket.Pinger(pingChannel, config.ServerWs.PingerSize, config.ServerWs.PingInterval, config.ServerWs.WriteWaitInterval)

	go reportConnectionMetrics(*wh.Table())

	restHandler := NewHandler()
	router := mux.NewRouter()
	router.Path("/ping").HandlerFunc(pingHandler).Methods(http.MethodGet)
	subRouter := router.PathPrefix("/api/v1").Subrouter()
	subRouter.HandleFunc("/events", wh.GetHandlerWSEvents(collector)).Methods(http.MethodGet).Name("events")
	subRouter.HandleFunc("/events", restHandler.GetRESTAPIHandler(collector)).Methods(http.MethodPost).Name("events")

	server := &http.Server{
		Handler: router,
		Addr:    ":" + config.ServerWs.AppPort,
	}
	s.s = server
	return server.ListenAndServe()
}

func (s Service) Name() string {
	return "REST"
}

func (s Service) Shutdown(ctx context.Context) {
	s.s.Shutdown(ctx)
}