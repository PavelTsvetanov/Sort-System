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
	f := &fulfilmentService{
		sortingRobot: sortingRobot,
		orderStatus:  orderToStatus{mapper: map[string]*gen.FulfilmentStatus{}},
		itemToCubby:  itemToCubby{mapper: map[string][]string{}}}
	f.orderRequests = scheduleRequests(f.processRequest)
	return f
}

const (
	nrOfCubbies          = math.MaxInt32
	sortingServerAddress = "localhost:10000"
)

type itemToCubby struct {
	sync.Mutex
	mapper map[string][]string
}

type orderToStatus struct {
	sync.Mutex
	mapper map[string]*gen.FulfilmentStatus
}

type fulfilmentService struct {
	sortingRobot  gen.SortingRobotClient
	itemToCubby   itemToCubby
	orderStatus   orderToStatus
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
	s.orderStatus.Lock()
	defer s.orderStatus.Unlock()
	return &gen.OrdersStatusResponse{Status: []*gen.FulfilmentStatus{s.orderStatus.mapper[request.OrderId]}}, nil
}

func (s *fulfilmentService) GetAllOrdersStatus(ctx context.Context, empty *gen.Empty) (*gen.OrdersStatusResponse, error) {
	s.orderStatus.Lock()
	defer s.orderStatus.Unlock()
	orders := &gen.OrdersStatusResponse{Status: []*gen.FulfilmentStatus{}}
	for _, orderStatus := range s.orderStatus.mapper {
		orders.Status = append(orders.Status, orderStatus)
	}
	return orders, nil
}

func (s *fulfilmentService) MarkFulfilled(ctx context.Context, request *gen.OrderIdRequest) (*gen.Empty, error) {
	panic("implement me")
}

func (s *fulfilmentService) LoadOrders(ctx context.Context, request *gen.LoadOrdersRequest) (*gen.CompleteResponse, error) {
	go func() {
		s.orderStatus.Lock()
		for _, order := range request.Orders {
			s.orderStatus.mapper[order.Id] = &gen.FulfilmentStatus{
				Order: order,
				Cubby: &gen.Cubby{},
				State: gen.OrderState_PENDING,
			}
		}
		s.orderStatus.Unlock()
		s.orderRequests <- request
	}()

	return &gen.CompleteResponse{Message: []string{}}, nil
}

func (s *fulfilmentService) processRequest(request *gen.LoadOrdersRequest) {
	oToCubbies := mapOrdersToCubbies(request.Orders)
	s.mapItemToCubby(request.Orders)
	for _, order := range request.Orders {
		s.orderStatus.Lock()
		s.orderStatus.mapper[order.Id].Cubby = &gen.Cubby{Id: oToCubbies[order.Id]}
		s.orderStatus.Unlock()
		for range order.Items {
			resp, err := s.sortingRobot.SelectItem(context.Background(), &gen.SelectItemRequest{})
			if err != nil {
				log.Fatalf("Robot failed to select an item: %s", err)
			}
			c, err := s.popNextCubbyForItem(resp.Item)
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
		used[cubby] = true
		m[order.Id] = cubby
	}
	return m
}

func (s *fulfilmentService) mapItemToCubby(orders []*gen.Order) {
	for _, order := range orders {
		s.orderStatus.Lock()
		cubby := s.orderStatus.mapper[order.Id].Cubby.Id
		s.orderStatus.Unlock()
		for _, item := range order.Items {
			s.itemToCubby.Lock()
			s.itemToCubby.mapper[item.Code] = append(s.itemToCubby.mapper[item.Code], cubby)
			s.itemToCubby.Unlock()
		}
	}
}

func (s *fulfilmentService) popNextCubbyForItem(item *gen.Item) (string, error) {
	s.itemToCubby.Lock()
	defer s.itemToCubby.Unlock()
	if cubbies, ok := s.itemToCubby.mapper[item.Code]; ok {
		if len(cubbies) != 0 {
			var cubby string
			cubby, s.itemToCubby.mapper[item.Code] = s.itemToCubby.mapper[item.Code][0], s.itemToCubby.mapper[item.Code][1:]
			return cubby, nil
		}
		return "", errors.New("no available cubbies left")
	}
	return "", errors.New("todo")
}
