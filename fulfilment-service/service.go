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
		activeOrders:        activeOrder{mapper: map[string]map[string]bool{}},
		sortingRobot:        sortingRobot,
		orderStatus:         orderToStatus{mapper: map[string]*gen.FulfilmentStatus{}},
		itemToPreparedOrder: itemToCubby{mapper: map[string][]*gen.PreparedOrder{}}}
	f.orderRequests = scheduleRequests(f.processRequest)
	f.orderItemCompleted = checkIfOrderIsCompleted(f.changeOrderStatusIfComplete)
	return f
}

const (
	nrOfCubbies          = math.MaxInt32
	sortingServerAddress = "localhost:10000"
)

type itemToCubby struct {
	sync.Mutex
	mapper map[string][]*gen.PreparedOrder
}

type orderToStatus struct {
	sync.Mutex
	mapper map[string]*gen.FulfilmentStatus
}

type activeOrder struct {
	sync.Mutex
	mapper map[string]map[string]bool
}

type fulfilmentService struct {
	sortingRobot        gen.SortingRobotClient
	itemToPreparedOrder itemToCubby
	orderStatus         orderToStatus
	orderRequests       chan *gen.LoadOrdersRequest
	orderItemCompleted  chan string
	activeOrders        activeOrder
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

func checkIfOrderIsCompleted(changeOrderStatus func(orderId string)) chan string {
	newItemPushed := make(chan string)
	go func() {
		for {
			changeOrderStatus(<-newItemPushed)
		}
	}()
	return newItemPushed
}

func (s *fulfilmentService) changeOrderStatusIfComplete(orderId string) {
	allItemsAreInCubby := true
	s.activeOrders.Lock()
	defer s.activeOrders.Unlock()
	for _, isInCubby := range s.activeOrders.mapper[orderId] {
		if !isInCubby {
			allItemsAreInCubby = false
			break
		}
	}
	if allItemsAreInCubby {
		delete(s.activeOrders.mapper, orderId)
		s.orderStatus.Lock()
		defer s.orderStatus.Unlock()
		s.orderStatus.mapper[orderId].State = gen.OrderState_READY
	}
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
			s.activeOrders.Lock()
			for _, item := range order.Items {
				if _, exists := s.activeOrders.mapper[order.Id]; !exists {
					s.activeOrders.mapper[order.Id] = make(map[string]bool)
				}
				s.activeOrders.mapper[order.Id][item.Code] = false
			}
			s.activeOrders.Unlock()
		}
		s.orderStatus.Unlock()
		s.orderRequests <- request
	}()

	return &gen.CompleteResponse{Message: []string{}}, nil
}

func (s *fulfilmentService) processRequest(request *gen.LoadOrdersRequest) {
	oToCubbies := mapOrdersToCubbies(request.Orders)
	s.mapItemToPreparedOrder(request.Orders)
	for _, order := range request.Orders {
		s.orderStatus.Lock()
		s.orderStatus.mapper[order.Id].Cubby = &gen.Cubby{Id: oToCubbies[order.Id]}
		s.orderStatus.Unlock()
		for range order.Items {
			resp, err := s.sortingRobot.SelectItem(context.Background(), &gen.SelectItemRequest{})
			if err != nil {
				log.Fatalf("Robot failed to select an item: %s", err)
			}
			preparedOrder, err := s.popNextPreparedOrderForItem(resp.Item)
			if err != nil {
				log.Fatal(err)
			}
			_, err = s.sortingRobot.MoveItem(
				context.Background(),
				&gen.MoveItemRequest{Cubby: &gen.Cubby{Id: preparedOrder.Cubby.Id}},
			)
			if err != nil {
				log.Fatalf("Robot failed to move an item: %s", err)
			}
			s.activeOrders.Lock()
			s.activeOrders.mapper[preparedOrder.Order.Id][resp.Item.Code] = true
			s.activeOrders.Unlock()
			s.orderItemCompleted <- preparedOrder.Order.Id
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

func (s *fulfilmentService) mapItemToPreparedOrder(orders []*gen.Order) {
	for _, order := range orders {
		s.orderStatus.Lock()
		cubby := s.orderStatus.mapper[order.Id].Cubby.Id
		s.orderStatus.Unlock()
		for _, item := range order.Items {
			s.itemToPreparedOrder.Lock()
			s.itemToPreparedOrder.mapper[item.Code] = append(
				s.itemToPreparedOrder.mapper[item.Code],
				&gen.PreparedOrder{
					Order: order,
					Cubby: &gen.Cubby{Id: cubby},
				},
			)
			s.itemToPreparedOrder.Unlock()
		}
	}
}

func (s *fulfilmentService) popNextPreparedOrderForItem(item *gen.Item) (*gen.PreparedOrder, error) {
	s.itemToPreparedOrder.Lock()
	defer s.itemToPreparedOrder.Unlock()
	if cubbies, ok := s.itemToPreparedOrder.mapper[item.Code]; ok {
		if len(cubbies) != 0 {
			var cubby *gen.PreparedOrder
			cubby, s.itemToPreparedOrder.mapper[item.Code] = s.itemToPreparedOrder.mapper[item.Code][0], s.itemToPreparedOrder.mapper[item.Code][1:]
			return cubby, nil
		}
		return nil, errors.New("no available cubbies left")
	}
	return nil, errors.New("no order was placed for item: " + item.Code)
}
