package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/PavelTsvetanov/sort-system/gen"
)

func newSortingService() gen.SortingRobotServer {
	return &sortingService{}
}

type sortingService struct {
	mutex        sync.Mutex
	items        []*gen.Item
	selectedItem *gen.Item
}

func (s *sortingService) LoadItems(ctx context.Context, req *gen.LoadItemsRequest) (*gen.LoadItemsResponse, error) {
	s.items = append(s.items, req.Items...)
	log.Printf("Added [%d] items to the bin, for total storage of [%d]", len(req.Items), len(s.items))
	return &gen.LoadItemsResponse{}, nil
}

func (s *sortingService) MoveItem(ctx context.Context, req *gen.MoveItemRequest) (*gen.MoveItemResponse, error) {
	if s.selectedItem == nil {
		log.Println("no item is selected")
		return nil, errors.New("no item is selected")
	}
	log.Printf("Placed %s in cubby %s", s.selectedItem.Code, req.Cubby.Id)
	s.selectedItem = nil
	return &gen.MoveItemResponse{}, nil
}

func (s *sortingService) SelectItemImpl(ctx context.Context, req *gen.SelectItemRequest, seed int64) (*gen.SelectItemResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.selectedItem != nil {
		log.Println("an item is already selected")
		return nil, errors.New("an item is already selected")
	}
	if len(s.items) == 0 {
		log.Println("no items loaded")
		return nil, errors.New("no items loaded")
	}
	rand.Seed(seed)
	idx := rand.Intn(len(s.items))
	s.selectedItem = s.items[idx]
	s.items = append(s.items[:idx], s.items[idx+1:]...)
	log.Printf("Selected item at position [%d] from the bin, [%d] items left", idx, len(s.items))

	return &gen.SelectItemResponse{Item: s.selectedItem}, nil
}

func (s *sortingService) SelectItem(ctx context.Context, req *gen.SelectItemRequest) (*gen.SelectItemResponse, error) {
	return s.SelectItemImpl(ctx, req, time.Now().Unix())
}
