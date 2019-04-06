package items

import (
	"context"

	// "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
)

// GetItem implements our gRPC server's interface
func (s *Service) GetItem(ctx context.Context, in *todoproto.GetItemRequest) (*todoproto.GetItemResponse, error) {
	item, err := s.itemDatabase.GetItem(ctx, in.ItemId, in.UserId)
	if err != nil {
		return nil, err
	}

	return &todoproto.GetItemResponse{Item: todoproto.ProtoItemFromModel(item)}, nil
}

// GetItemCount implements our gRPC server's interface
func (s *Service) GetItemCount(ctx context.Context, in *todoproto.ItemListRequest) (*todoproto.ItemCountResponse, error) {
	count, err := s.itemDatabase.GetItemCount(ctx, in.Filter.ToModelQueryFilter(), in.UserId)
	if err != nil {
		return nil, err
	}

	return &todoproto.ItemCountResponse{
		Count: count,
	}, nil
}

// GetItems implements our gRPC server's interface
func (s *Service) GetItems(ctx context.Context, in *todoproto.ItemListRequest) (*todoproto.ItemListResponse, error) {
	items, err := s.itemDatabase.GetItems(ctx, in.Filter.ToModelQueryFilter(), in.UserId)
	if err != nil {
		return nil, err
	}

	res := &todoproto.ItemListResponse{
		Items: todoproto.ProtoItemsFromModels(items.Items),
	}

	return res, nil
}

// CreateItem implements our gRPC server's interface
func (s *Service) CreateItem(ctx context.Context, in *todoproto.CreateItemRequest) (*todoproto.CreateItemResponse, error) {
	item, err := s.itemDatabase.CreateItem(ctx, in.ToItemInput())
	if err != nil {
		return nil, err
	}

	res := &todoproto.CreateItemResponse{
		Item: todoproto.ProtoItemFromModel(item),
	}

	return res, nil
}

// UpdateItem implements our gRPC server's interface
func (s *Service) UpdateItem(ctx context.Context, in *todoproto.UpdateItemRequest) (*todoproto.UpdateItemResponse, error) {
	err := s.itemDatabase.UpdateItem(ctx, in.Item.ToModelItem())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// DeleteItem implements our gRPC server's interface
func (s *Service) DeleteItem(ctx context.Context, in *todoproto.DeleteItemRequest) (*todoproto.ErrorResponse, error) {

	err := s.itemDatabase.DeleteItem(ctx, in.ItemId, in.UserId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
