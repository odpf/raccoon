package websocket

import (
	"context"
	"net/http"
	"raccoon/config"
	"raccoon/logger"
	"runtime"
	"time"

	"raccoon/metrics"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	// https://golang.org/pkg/net/http/pprof/
	_ "net/http/pprof"
)

type Server struct {
	HTTPServer    *http.Server
	bufferChannel chan EventsBatch
	user          *User
	pingChannel   chan connection
}

func (s *Server) StartHTTPServer(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		logger.Info("WebSocket Server --> startHttpServer")
		err := s.HTTPServer.ListenAndServe()
		if err != http.ErrServerClosed {
			logger.Errorf("WebSocket Server --> HTTP Server could not be started = %s", err.Error())
			cancel()
		}
	}()
	go s.ReportServerMetrics()
	go Pinger(s.pingChannel, config.Websocket.PingerSize, config.Websocket.PingInterval, config.Websocket.WriteWaitInterval)
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			logger.Errorf("WebSocket Server --> pprof could not be enabled: %s", err.Error())
			cancel()
		} else {
			logger.Info("WebSocket Server --> pprof :: Enabled")
		}
	}()
}

func (s *Server) ReportServerMetrics() {
	t := time.Tick(config.Statsd.FlushPeriodMs)
	m := &runtime.MemStats{}
	for {
		<-t
		metrics.Gauge("connections_count_current", s.user.TotalUsers(), "")
		metrics.Gauge("server_go_routines_count_current", runtime.NumGoroutine(), "")

		runtime.ReadMemStats(m)
		metrics.Gauge("server_mem_heap_alloc_bytes_current", m.HeapAlloc, "")
		metrics.Gauge("server_mem_heap_inuse_bytes_current", m.HeapInuse, "")
		metrics.Gauge("server_mem_heap_objects_total_current", m.HeapObjects, "")
		metrics.Gauge("server_mem_stack_inuse_bytes_current", m.StackInuse, "")
		metrics.Gauge("server_mem_gc_triggered_current", m.LastGC/1000, "")
		metrics.Gauge("server_mem_gc_pauseNs_current", m.PauseNs[(m.NumGC+255)%256]/1000, "")
		metrics.Gauge("server_mem_gc_count_current", m.NumGC, "")
		metrics.Gauge("server_mem_gc_pauseTotalNs_current", m.PauseTotalNs, "")
	}
}

//CreateServer - instantiates the http server
func CreateServer() (*Server, chan EventsBatch) {
	//create the websocket handler that upgrades the http request
	bufferChannel := make(chan EventsBatch, config.Worker.ChannelSize)
	pingChannel := make(chan connection, config.Websocket.ServerMaxConn)
	user := NewUserStore(config.Websocket.ServerMaxConn)
	wsHandler := &Handler{
		websocketUpgrader: getWebSocketUpgrader(config.Websocket.ReadBufferSize, config.Websocket.WriteBufferSize, config.Websocket.CheckOrigin),
		bufferChannel:     bufferChannel,
		user:              user,
		PongWaitInterval:  config.Websocket.PongWaitInterval,
		WriteWaitInterval: config.Websocket.WriteWaitInterval,
		PingChannel:       pingChannel,
		UserIDHeader:      config.Websocket.UserIDHeader,
	}
	server := &Server{
		HTTPServer: &http.Server{
			Handler: Router(wsHandler),
			Addr:    ":" + config.Websocket.AppPort,
		},
		bufferChannel: bufferChannel,
		user:          user,
		pingChannel:   pingChannel,
	}
	//Wrap the handler with a Server instance and return it
	return server, bufferChannel
}

// Router sets up the routes
func Router(h *Handler) http.Handler {
	router := mux.NewRouter()
	router.Path("/ping").HandlerFunc(PingHandler).Methods(http.MethodGet)
	subRouter := router.PathPrefix("/api/v1").Subrouter()
	subRouter.HandleFunc("/events", h.HandlerWSEvents).Methods(http.MethodGet).Name("events")
	return router
}

func getWebSocketUpgrader(readBufferSize int, writeBufferSize int, checkOrigin bool) websocket.Upgrader {
	ug := websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return checkOrigin
		},
	}
	return ug
}
