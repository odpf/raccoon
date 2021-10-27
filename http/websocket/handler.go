package websocket

import (
	"fmt"
	"net/http"
	"raccoon/config"
	"raccoon/http/websocket/connection"
	"raccoon/logger"
	"raccoon/metrics"
	"raccoon/pkg/collection"
	"raccoon/pkg/deserialization"
	"raccoon/pkg/serialization"
	"time"

	pb "raccoon/pkg/proto"

	"github.com/gorilla/websocket"
)

type Handler struct {
	upgrader    *connection.Upgrader
	PingChannel chan connection.Conn
}

func NewHandler(pingC chan connection.Conn) *Handler {
	ugConfig := connection.UpgraderConfig{
		ReadBufferSize:    config.ServerWs.ReadBufferSize,
		WriteBufferSize:   config.ServerWs.WriteBufferSize,
		CheckOrigin:       config.ServerWs.CheckOrigin,
		MaxUser:           config.ServerWs.ServerMaxConn,
		PongWaitInterval:  config.ServerWs.PongWaitInterval,
		WriteWaitInterval: config.ServerWs.WriteWaitInterval,
		ConnIDHeader:      config.ServerWs.ConnIDHeader,
		ConnGroupHeader:   config.ServerWs.ConnGroupHeader,
	}

	upgrader := connection.NewUpgrader(ugConfig)
	return &Handler{
		upgrader: upgrader,
	}
}

func (h *Handler) Table() *connection.Table {
	return h.upgrader.Table
}

//HandlerWSEvents handles the upgrade and the events sent by the peers
func (h *Handler) GetHandlerWSEvents(collector collection.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := h.upgrader.Upgrade(w, r)
		if err != nil {
			logger.Debugf("[websocket.Handler] %v", err)
			return
		}
		defer conn.Close()
		h.PingChannel <- conn
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway,
					websocket.CloseNormalClosure,
					websocket.CloseNoStatusReceived,
					websocket.CloseAbnormalClosure) {
					logger.Error(fmt.Sprintf("[websocket.Handler] %s closed abruptly: %v", conn.Identifier, err))
					metrics.Increment("batches_read_total", fmt.Sprintf("status=failed,reason=closeerror,conn_group=%s", conn.Identifier.Group))
					break
				}
				metrics.Increment("batches_read_total", fmt.Sprintf("status=failed,reason=unknown,conn_group=%s", conn.Identifier.Group))
				logger.Error(fmt.Sprintf("[websocket.Handler] reading message failed. Unknown failure for %s: %v", conn.Identifier, err)) //no connection issue here
				break
			}

			timeConsumed := time.Now()
			metrics.Count("events_rx_bytes_total", len(message), fmt.Sprintf("conn_group=%s", conn.Identifier.Group))
			payload := &pb.EventRequest{}

			d, s := h.getDeserializerSerializer(messageType)
			if err := d.Deserialize(message, payload); err != nil {
				logger.Error(fmt.Sprintf("[websocket.Handler] reading message failed for %s: %v", conn.Identifier, err))
				metrics.Increment("batches_read_total", fmt.Sprintf("status=failed,reason=serde,conn_group=%s", conn.Identifier.Group))
				writeBadRequestResponse(conn, s, messageType, err)
				continue
			}
			metrics.Increment("batches_read_total", fmt.Sprintf("status=success,conn_group=%s", conn.Identifier.Group))
			metrics.Count("events_rx_total", len(payload.Events), fmt.Sprintf("conn_group=%s", conn.Identifier.Group))
			collector.Collect(r.Context(), &collection.CollectRequest{
				ConnectionIdentifier: &conn.Identifier,
				TimeConsumed:         timeConsumed,
				EventRequest:         payload,
			})
			writeSuccessResponse(conn, s, messageType, payload.ReqGuid)
		}
	}

}

func writeSuccessResponse(conn connection.Conn, serializer serialization.Serializer, messageType int, requestGUID string) {
	response := &pb.EventResponse{
		Status:   pb.Status_SUCCESS,
		Code:     pb.Code_OK,
		SentTime: time.Now().Unix(),
		Reason:   "",
		Data: map[string]string{
			"req_guid": requestGUID,
		},
	}
	success, _ := serializer.Serialize(response)
	conn.WriteMessage(messageType, success)
}

func writeBadRequestResponse(conn connection.Conn, serializer serialization.Serializer, messageType int, err error) {
	response := &pb.EventResponse{
		Status:   pb.Status_ERROR,
		Code:     pb.Code_BAD_REQUEST,
		SentTime: time.Now().Unix(),
		Reason:   fmt.Sprintf("cannot deserialize request: %s", err),
		Data:     nil,
	}

	failure, _ := serializer.Serialize(response)
	conn.WriteMessage(messageType, failure)
}

func (h *Handler) getDeserializerSerializer(messageType int) (deserialization.Deserializer, serialization.Serializer) {
	switch messageType {
	case websocket.BinaryMessage:
		return deserialization.ProtoDeserilizer(), serialization.ProtoDeserilizer()
	case websocket.TextMessage:
		return deserialization.JSONDeserializer(), serialization.JSONSerializer()
	default:
		return deserialization.ProtoDeserilizer(), serialization.ProtoDeserilizer()
	}
}