package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/PavelTsvetanov/sort-system/gen"
)

func newSortingService() gen.SortingRobotServer {
	return &sortingService{}
}

type sortingService struct {
	items        []*gen.Item
	selectedItem *gen.Item
}

func (s *sortingService) LoadItems(ctx context.Context, req *gen.LoadItemsRequest) (*gen.LoadItemsResponse, error) {
	s.items = append(s.items, req.Items...)
	return &gen.LoadItemsResponse{}, nil
}

func (s *sortingService) MoveItem(context.Context, *gen.MoveItemRequest) (*gen.MoveItemResponse, error) {
	if s.selectedItem == nil {
		return nil, errors.New("no item is selected")
	}
	s.selectedItem = nil
	return &gen.MoveItemResponse{}, nil
}

func (s *sortingService) SelectItemImpl(ctx context.Context, req *gen.SelectItemRequest, seed int64) (*gen.SelectItemResponse, error) {
	if s.selectedItem != nil {
		return nil, errors.New("an item is already selected")
	}
	if len(s.items) == 0 {
		return nil, errors.New("sorting robot has no items loaded")
	}
	rand.Seed(seed)
	idx := rand.Intn(len(s.items))
	s.selectedItem = s.items[idx]
	s.items = append(s.items[:idx], s.items[idx+1:]...)
	return &gen.SelectItemResponse{Item: s.selectedItem}, nil
}

func (s *sortingService) SelectItem(ctx context.Context, req *gen.SelectItemRequest) (*gen.SelectItemResponse, error) {
	return s.SelectItemImpl(ctx, req, time.Now().Unix())
}
