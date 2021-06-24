package main

import (
	"context"
	"errors"
	"github.com/PavelTsvetanov/sort-system/gen"
	"github.com/preslavmihaylov/ordertocubby"
	"google.golang.org/grpc"
	"log"
)

func newFulfilmentService() gen.FulfillmentServer {
	return &fulfilmentService{}
}

const (
	nrOfCubbies          = 10
	sortingServerAddress = "localhost:10000"
)

type fulfilmentService struct {
}

func (s *fulfilmentService) LoadOrders(ctx context.Context, request *gen.LoadOrdersRequest) (*gen.CompleteResponse, error) {
	//move
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(sortingServerAddress, opts...)
	if err != nil {
		log.Fatalf("Failed ot dial sorting robot server: %s", err)
	}
	defer conn.Close()

	sortingService := gen.NewSortingRobotClient(conn)
	orderToCubby := mapOrdersToCubbies(request.Orders)
	itemToCubby := mapItemToCubby(request.Orders, orderToCubby)

	for _, order := range request.Orders {
		for range order.Items {
			resp, err := sortingService.SelectItem(ctx, &gen.SelectItemRequest{})
			if err != nil {
				log.Fatalf("Robot failed to select an item: %s", err)
			}
			c, err := getCubbyForItem(resp.Item, itemToCubby)
			if err != nil {
				log.Fatal(err)
			}
			_, err = sortingService.MoveItem(ctx, &gen.MoveItemRequest{Cubby: &gen.Cubby{Id: c}})
			if err != nil {
				log.Fatalf("Robot failed to move an item: %s", err)
			}
		}
	}
	return mapToCompleteResponse(orderToCubby), nil
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

func mapToCompleteResponse(orderToCubby map[string]string) *gen.CompleteResponse {
	var resp = gen.CompleteResponse{}
	for order, cubby := range orderToCubby {
		orderResp := gen.PreparedOrder{Order: &gen.Order{Id: order}, Cubby: &gen.Cubby{Id: cubby}}
		resp.Orders = append(resp.Orders, &orderResp)
	}
	return &resp
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

func mapItemToCubby(orders []*gen.Order, oToCubby map[string]string) map[string][]string {
	m := make(map[string][]string)
	for _, order := range orders {
		cubby := oToCubby[order.Id]
		for _, item := range order.Items {
			m[item.Code] = append(m[item.Code], cubby)
		}
	}
	return m
}

func getCubbyForItem(item *gen.Item, orders map[string][]string) (string, error) {
	if cubbies, ok := orders[item.Code]; ok {
		if len(cubbies) != 0 {
			var cubby string
			cubby, orders[item.Code] = orders[item.Code][0], orders[item.Code][1:]
			return cubby, nil
		}
		return "", errors.New("no available cubbies left")
	}
	return "", errors.New("todo")
}
