package main

import (
	"context"
	"errors"
	"github.com/PavelTsvetanov/sort-system/gen"
	"github.com/preslavmihaylov/ordertocubby"
	"log"
	"math"
	"sync"
)

func newFulfilmentService(sortingRobot gen.SortingRobotClient) gen.FulfillmentServer {
	f := &fulfilmentService{sortingRobot: sortingRobot}
	f.orderRequests = scheduleRequests(f.processRequest)
	return f
}

const (
	nrOfCubbies          = math.MaxInt32
	sortingServerAddress = "localhost:10000"
)

type fulfilmentService struct {
	sortingRobot  gen.SortingRobotClient
	oMap          sync.Map
	orderRequests chan *gen.LoadOrdersRequest
}

func scheduleRequests(processRequest func(request *gen.LoadOrdersRequest)) chan *gen.LoadOrdersRequest {
	requests := make(chan *gen.LoadOrdersRequest)
	go func() {
		log.Printf("Processing requests...")
		for {
			processRequest(<-requests)
		}
	}()
	return requests
}

func (s *fulfilmentService) GetOrderStatusById(ctx context.Context, request *gen.OrderIdRequest) (*gen.OrdersStatusResponse, error) {
	panic("implement me")
}

func (s *fulfilmentService) GetAllOrdersStatus(ctx context.Context, empty *gen.Empty) (*gen.OrdersStatusResponse, error) {
	panic("implement me")
}

func (s *fulfilmentService) MarkFulfilled(ctx context.Context, request *gen.OrderIdRequest) (*gen.Empty, error) {
	panic("implement me")
}

func (s *fulfilmentService) LoadOrders(ctx context.Context, request *gen.LoadOrdersRequest) (*gen.CompleteResponse, error) {
	go func() {
		s.orderRequests <- request
	}()

	return &gen.CompleteResponse{Message: []string{}}, nil
}

func (s *fulfilmentService) processRequest(request *gen.LoadOrdersRequest) {
	oToCubbies := mapOrdersToCubbies(request.Orders)
	itemToCubbies := s.mapItemToCubby(request.Orders, oToCubbies)
	for _, order := range request.Orders {
		for range order.Items {
			resp, err := s.sortingRobot.SelectItem(context.Background(), &gen.SelectItemRequest{})
			if err != nil {
				log.Fatalf("Robot failed to select an item: %s", err)
			}
			c, err := s.getCubbyForItem(resp.Item, itemToCubbies)
			if err != nil {
				log.Fatal(err)
			}
			_, err = s.sortingRobot.MoveItem(context.Background(), &gen.MoveItemRequest{Cubby: &gen.Cubby{Id: c}})
			if err != nil {
				log.Fatalf("Robot failed to move an item: %s", err)
			}
		}
	}
}

func getFreeCubby(orderId string, usedCubbies map[string]bool) string {
	times := 1
	for {
		cubbyID := ordertocubby.Map(orderId, uint32(times), uint32(nrOfCubbies))
		if !usedCubbies[cubbyID] {
			return cubbyID
		}
		times++
	}
}

func mapOrdersToCubbies(orders []*gen.Order) map[string]string {
	m := make(map[string]string)
	used := make(map[string]bool)
	for _, order := range orders {
		cubby := getFreeCubby(order.Id, used)
		used[order.Id] = true
		m[order.Id] = cubby
	}
	return m
}

func (s *fulfilmentService) mapItemToCubby(orders []*gen.Order, oToCubby map[string]string) map[string][]string {
	m := make(map[string][]string)
	for _, order := range orders {
		cubby := oToCubby[order.Id]
		for _, item := range order.Items {
			m[item.Code] = append(m[item.Code], cubby)
		}
	}
	return m
}

func (s *fulfilmentService) getCubbyForItem(item *gen.Item, itemToCubby map[string][]string) (string, error) {
	if cubbies, ok := itemToCubby[item.Code]; ok {
		if len(cubbies) != 0 {
			var cubby string
			cubby, itemToCubby[item.Code] = itemToCubby[item.Code][0], itemToCubby[item.Code][1:]
			return cubby, nil
		}
		return "", errors.New("no available cubbies left")
	}
	return "", errors.New("todo")
}
