package main

import (
	"context"
	"github.com/PavelTsvetanov/sort-system/gen"
)

func newFulfilmentService() gen.FulfillmentServer {
	return &fulfilmentService{}
}

type fulfilmentService struct {
}

func (s *fulfilmentService) LoadOrders(ctx context.Context, request *gen.LoadOrdersRequest) (*gen.CompleteResponse, error) {
	panic("implement me")
}
