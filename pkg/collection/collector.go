package collection

import (
	"context"
	"time"

	"raccoon/pkg/identification"
	pb "raccoon/pkg/proto"
)

type CollectRequest struct {
	ConnectionIdentifier *identification.Identifier
	TimeConsumed         time.Time
	*pb.EventRequest
}

type EventsBatch struct {
	ConnIdentifier *identification.Identifier
	EventReq       *pb.EventRequest
	TimeConsumed   time.Time
	TimePushed     time.Time
}

type Collector interface {
	Collect(ctx context.Context, req *CollectRequest) error
}