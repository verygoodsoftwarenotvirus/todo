package grpcserver

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
)

// GetItem implements our requisite gRPC server method
func (s *GRPCServer) GetItem(ctx context.Context, req *todoproto.GetItemRequest) (*todoproto.GetItemResponse, error) {
	return s.itemsService.GetItem(ctx, req)
}

// GetItemCount implements our requisite gRPC server method
func (s *GRPCServer) GetItemCount(ctx context.Context, req *todoproto.ItemListRequest) (*todoproto.ItemCountResponse, error) {
	return s.itemsService.GetItemCount(ctx, req)
}

// GetItems implements our requisite gRPC server method
func (s *GRPCServer) GetItems(ctx context.Context, req *todoproto.ItemListRequest) (*todoproto.ItemListResponse, error) {
	return s.itemsService.GetItems(ctx, req)
}

// CreateItem implements our requisite gRPC server method
func (s *GRPCServer) CreateItem(ctx context.Context, req *todoproto.CreateItemRequest) (*todoproto.CreateItemResponse, error) {
	return s.itemsService.CreateItem(ctx, req)
}

// UpdateItem implements our requisite gRPC server method
func (s *GRPCServer) UpdateItem(ctx context.Context, req *todoproto.UpdateItemRequest) (*todoproto.UpdateItemResponse, error) {
	return s.itemsService.UpdateItem(ctx, req)
}

// DeleteItem implements our requisite gRPC server method
func (s *GRPCServer) DeleteItem(ctx context.Context, req *todoproto.DeleteItemRequest) (*todoproto.ErrorResponse, error) {
	return s.itemsService.DeleteItem(ctx, req)
}
