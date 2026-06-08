package unit_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

type buyerRepositoryStub struct {
	createBuyerFunc  func(ctx context.Context, buyer *model.BuyerModel) (int64, error)
	getBuyerByIDFunc func(ctx context.Context, buyerID int64) (*model.BuyerModel, error)
	getAllBuyersFunc func(ctx context.Context) ([]model.BuyerModel, error)
	updateBuyerFunc  func(ctx context.Context, buyer *model.BuyerModel) error
	deleteBuyerFunc  func(ctx context.Context, buyerID int64) error
}

func (s *buyerRepositoryStub) CreateBuyer(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
	if s.createBuyerFunc == nil {
		return 0, nil
	}

	return s.createBuyerFunc(ctx, buyer)
}

func (s *buyerRepositoryStub) GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error) {
	if s.getBuyerByIDFunc == nil {
		return nil, nil
	}

	return s.getBuyerByIDFunc(ctx, buyerID)
}

func (s *buyerRepositoryStub) GetAllBuyers(ctx context.Context) ([]model.BuyerModel, error) {
	if s.getAllBuyersFunc == nil {
		return nil, nil
	}

	return s.getAllBuyersFunc(ctx)
}

func (s *buyerRepositoryStub) UpdateBuyer(ctx context.Context, buyer *model.BuyerModel) error {
	if s.updateBuyerFunc == nil {
		return nil
	}

	return s.updateBuyerFunc(ctx, buyer)
}

func (s *buyerRepositoryStub) DeleteBuyer(ctx context.Context, buyerID int64) error {
	if s.deleteBuyerFunc == nil {
		return nil
	}

	return s.deleteBuyerFunc(ctx, buyerID)
}

func TestBuyerServiceCreateReturnsConflictWhenDatabaseRejectsDuplicateDocument(t *testing.T) {
	t.Parallel()

	service := buyerservice.NewService(&buyerRepositoryStub{
		createBuyerFunc: func(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
			return 0, &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry '12345678900' for key 'uq_buyers_document'"}
		},
	})

	_, err := service.Create(context.Background(), &dto.CreateBuyerRequest{
		Name:     "Marcos Ferreira",
		Document: "12345678900",
		Phone:    "11999999999",
		Email:    "marcos@example.com",
	})

	assertStatusCode(t, err, http.StatusConflict)
	assertErrorMessage(t, err, "buyer document already exists")
}

func TestBuyerServiceCreateReturnsConflictWhenDatabaseRejectsDuplicateEmail(t *testing.T) {
	t.Parallel()

	service := buyerservice.NewService(&buyerRepositoryStub{
		createBuyerFunc: func(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
			return 0, &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry 'marcos@example.com' for key 'uq_buyers_email'"}
		},
	})

	_, err := service.Create(context.Background(), &dto.CreateBuyerRequest{
		Name:     "Marcos Ferreira",
		Document: "12345678900",
		Phone:    "11999999999",
		Email:    "marcos@example.com",
	})

	assertStatusCode(t, err, http.StatusConflict)
	assertErrorMessage(t, err, "buyer email already exists")
}

func TestBuyerServiceCreateReturnsConflictWhenDatabaseRejectsDuplicatePhone(t *testing.T) {
	t.Parallel()

	service := buyerservice.NewService(&buyerRepositoryStub{
		createBuyerFunc: func(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
			return 0, &mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry '11999999999' for key 'uq_buyers_phone'"}
		},
	})

	_, err := service.Create(context.Background(), &dto.CreateBuyerRequest{
		Name:     "Marcos Ferreira",
		Document: "12345678900",
		Phone:    "11999999999",
		Email:    "marcos@example.com",
	})

	assertStatusCode(t, err, http.StatusConflict)
	assertErrorMessage(t, err, "buyer phone already exists")
}

func TestBuyerServiceCreateReturnsInternalWhenCreateFailsWithoutDuplicateKey(t *testing.T) {
	t.Parallel()

	service := buyerservice.NewService(&buyerRepositoryStub{
		createBuyerFunc: func(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
			return 0, errors.New("database unavailable")
		},
	})

	_, err := service.Create(context.Background(), &dto.CreateBuyerRequest{
		Name:     "Marcos Ferreira",
		Document: "12345678900",
		Phone:    "11999999999",
		Email:    "marcos@example.com",
	})

	assertStatusCode(t, err, http.StatusInternalServerError)
}

func TestBuyerServiceCreateCreatesBuyerWhenDatabaseAcceptsUniqueValues(t *testing.T) {
	t.Parallel()

	service := buyerservice.NewService(&buyerRepositoryStub{
		createBuyerFunc: func(ctx context.Context, buyer *model.BuyerModel) (int64, error) {
			return 10, nil
		},
	})

	buyerID, err := service.Create(context.Background(), &dto.CreateBuyerRequest{
		Name:     "Marcos Ferreira",
		Document: "12345678900",
		Phone:    "11999999999",
		Email:    "marcos@example.com",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if buyerID != 10 {
		t.Fatalf("expected buyer id 10, got %d", buyerID)
	}
}
