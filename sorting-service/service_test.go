package main

import (
	"context"
	"github.com/PavelTsvetanov/sort-system/gen"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createServiceWithTwoLoadedItems(t *testing.T) gen.SortingRobotServer {
	items := []*gen.Item{{Code: "1", Label: "Apple"}, {Code: "2", Label: "Banana"}}
	return createServiceAndLoadItems(t, items)
}

func createServiceAndLoadItems(t *testing.T, items []*gen.Item) gen.SortingRobotServer {
	s := newSortingService()
	resp, err := s.LoadItems(context.Background(), &gen.LoadItemsRequest{Items: items})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	return s
}

func TestLoadItemNoError(t *testing.T) {
	s := newSortingService()
	items := []*gen.Item{{Code: "1", Label: "Apple"}, {Code: "2", Label: "Banana"}}
	resp, err := s.LoadItems(context.Background(), &gen.LoadItemsRequest{Items: items})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestSelectItemWhenNoLoadedItems(t *testing.T) {
	s := newSortingService()
	resp, err := s.SelectItem(context.Background(), &gen.SelectItemRequest{})
	assert.Nil(t, resp)
	expectedErrorMsg := "sorting robot has no items loaded"
	assert.EqualError(t, err, expectedErrorMsg)
}

func TestSelectItemWhenAnItemIsAlreadySelected(t *testing.T) {
	s := createServiceWithTwoLoadedItems(t)
	resp, err := s.SelectItem(context.Background(), &gen.SelectItemRequest{})
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	resp, err = s.SelectItem(context.Background(), &gen.SelectItemRequest{})
	assert.Nil(t, resp)
	expectedErrorMsg := "an item is already selected"
	assert.EqualError(t, err, expectedErrorMsg)
}

func TestSelectItemWhenTwoItemsLoaded(t *testing.T) {
	items := []*gen.Item{{Code: "1", Label: "Apple"}, {Code: "2", Label: "Banana"}}
	s := createServiceAndLoadItems(t, items)
	resp, err := s.SelectItem(context.Background(), &gen.SelectItemRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, items, resp.Item)
}

func TestMoveItemWhenAnItemIsSelected(t *testing.T) {
	s := createServiceWithTwoLoadedItems(t)
	resp, err := s.SelectItem(context.Background(), &gen.SelectItemRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	moveItemResp, err := s.MoveItem(context.Background(), &gen.MoveItemRequest{Cubby: &gen.Cubby{Id: "1"}})
	assert.Nil(t, err)
	assert.NotNil(t, moveItemResp)
}

func TestMoveItemWhenNoItemIsSelected(t *testing.T) {
	s := newSortingService()
	resp, err := s.MoveItem(context.Background(), &gen.MoveItemRequest{Cubby: &gen.Cubby{Id: "1"}})
	assert.Nil(t, resp)
	expectedErrorMsg := "no item is selected"
	assert.EqualError(t, err, expectedErrorMsg)
}
