# Fluxo de Orders: Handler, Service e Repository

## Visao Geral

Este documento segue o mesmo padrao de `01-crud-buyers.md` e `02-crud-couriers.md`, mantendo a organizacao inspirada no projeto `go-tweets`.

O dominio `order` cuida de:

- criacao de pedido
- busca por ID
- listagem
- atualizacao de dados do pedido
- mudanca para `ready_for_delivery`
- cancelamento

Fluxo:

```text
HTTP /orders/* -> handler/order -> service/order -> repository/order -> MySQL
```

## Atualizacao Do Bootstrap

No estado atual do projeto, a base HTTP ja foi padronizada com:

- `Gin`
- `validator`
- `internal/httpresponse`
- `internal/apperror`
- `internal/server/router.go`

Entao, se algum trecho antigo abaixo ainda mencionar `ServeMux`, `net/http` puro ou registro principal em `cmd/main.go`, adapte para o padrao vigente:

- handlers com `*gin.Context`
- construtores usando `gin.IRouter`
- bind e validacao via `internal/httpresponse`
- ligacao das rotas em `internal/server/router.go`

Diferenca importante em relacao a `buyers` e `couriers`:

- `order` depende de `buyer`
- `order` depende de `address`
- o `service` de `order` passa a coordenar mais de um repository

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar um fluxo de `orders` com:

- validacao de entrada no `handler`
- `service` concentrando regra de negocio
- persistencia separada entre `address` e `order`
- dependencia explicita de `buyer` para validar existencia

Este e o primeiro CRUD da fase 1 em que o `service` realmente orquestra multiplas partes do dominio.

---

## Contrato HTTP Inicial

Para o CRUD de `orders`, o contrato sugerido e:

- `POST /orders`
- `GET /orders/:order_id`
- `GET /orders`
- `PUT /orders/:order_id`
- `PATCH /orders/:order_id/ready-for-delivery`
- `PATCH /orders/:order_id/cancel`

Payload de criacao:

```json
{
  "buyer_id": 1,
  "delivery_address": {
    "zip_code": "01001000",
    "street": "Rua Exemplo",
    "number": "120",
    "complement": "Apto 12",
    "district": "Centro",
    "city": "Sao Paulo",
    "state": "SP"
  },
  "total_amount": 89.9
}
```

Payload de atualizacao:

```json
{
  "delivery_address": {
    "zip_code": "01001000",
    "street": "Rua Exemplo Atualizada",
    "number": "220",
    "complement": "Casa 2",
    "district": "Centro",
    "city": "Sao Paulo",
    "state": "SP"
  },
  "total_amount": 99.9
}
```

Payload para `ready-for-delivery`:

```json
{}
```

Payload para `cancel`:

```json
{}
```

---

## Status Sugeridos

Para a fase 1, voce pode trabalhar com estes status:

- `created`
- `ready_for_delivery`
- `cancelled`

### Regras De Negocio Sugeridas

- todo pedido nasce como `created`
- apenas pedido `created` pode virar `ready_for_delivery`
- apenas pedido `created` pode ser cancelado
- pedido `cancelled` nao volta de status

Essas regras ficam no `service`, nao no `handler`.

---

## Estrutura Recomendada

Para manter o mesmo padrao dos documentos anteriores, a estrutura pode ficar assim:

```text
internal/
  dto/
    order_dto.go
  model/
    address_model.go
    order_model.go
  handler/
    order/
      create_order.go
      get_all_orders.go
      get_order_by_id.go
      handler.go
      mark_order_ready_for_delivery.go
      cancel_order.go
      update_order.go
  repository/
    address/
      create_address.go
      get_address_by_id.go
      repository.go
      update_address.go
    order/
      create_order.go
      get_all_orders.go
      get_order_by_id.go
      repository.go
      update_order.go
      update_order_status.go
  service/
    order/
      create_order.go
      get_all_orders.go
      get_order_by_id.go
      service.go
      update_order.go
      mark_order_ready_for_delivery.go
      cancel_order.go
```

Arquivos compartilhados ja existentes:

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `internal/repository/buyer/repository.go`
- `internal/model/buyer_model.go`

---

## `internal/dto/order_dto.go`

Concentre os DTOs do dominio em um unico arquivo, como foi feito nos outros modulos.

Digite em `internal/dto/order_dto.go`:

```go
package dto

import "time"

type (
	DeliveryAddressRequest struct {
		ZipCode    string `json:"zip_code" validate:"required"`
		Street     string `json:"street" validate:"required"`
		Number     string `json:"number" validate:"required"`
		Complement string `json:"complement"`
		District   string `json:"district"`
		City       string `json:"city" validate:"required"`
		State      string `json:"state" validate:"required"`
	}

	CreateOrderRequest struct {
		BuyerID         int64                  `json:"buyer_id" validate:"required"`
		DeliveryAddress DeliveryAddressRequest `json:"delivery_address" validate:"required"`
		TotalAmount     float64                `json:"total_amount" validate:"required,gt=0"`
	}

	UpdateOrderRequest struct {
		DeliveryAddress DeliveryAddressRequest `json:"delivery_address" validate:"required"`
		TotalAmount     float64                `json:"total_amount" validate:"required,gt=0"`
	}

	OrderAddressResponse struct {
		ID         int64     `json:"id"`
		ZipCode    string    `json:"zip_code"`
		Street     string    `json:"street"`
		Number     string    `json:"number"`
		Complement string    `json:"complement"`
		District   string    `json:"district"`
		City       string    `json:"city"`
		State      string    `json:"state"`
		Latitude   *float64  `json:"latitude,omitempty"`
		Longitude  *float64  `json:"longitude,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	OrderResponse struct {
		ID              int64                `json:"id"`
		Code            string               `json:"code"`
		BuyerID         int64                `json:"buyer_id"`
		Status          string               `json:"status"`
		TotalAmount     float64              `json:"total_amount"`
		DeliveryAddress OrderAddressResponse `json:"delivery_address"`
		CreatedAt       time.Time            `json:"created_at"`
		UpdatedAt       time.Time            `json:"updated_at"`
	}
)
```

### Observacao

Repare que o DTO de `order` ja carrega um objeto aninhado para endereco.

Isso e natural porque:

- o cliente HTTP envia endereco junto com o pedido
- mas no banco o endereco fica em tabela separada

Essa traducao entre request aninhado e persistencia separada e responsabilidade do `service`.

---

## `internal/model/address_model.go`

Digite em `internal/model/address_model.go`:

```go
package model

import "time"

type AddressModel struct {
	ID         int64
	ZipCode    string
	Street     string
	Number     string
	Complement string
	District   string
	City       string
	State      string
	Latitude   *float64
	Longitude  *float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
```

---

## `internal/model/order_model.go`

Digite em `internal/model/order_model.go`:

```go
package model

import "time"

type OrderModel struct {
	ID                int64
	Code              string
	BuyerID           int64
	DeliveryAddressID int64
	Status            string
	TotalAmount       float64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
```

---

## `internal/repository/address/repository.go`

Digite em `internal/repository/address/repository.go`:

```go
package address

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, model *model.AddressModel) (int64, error)
	GetAddressByID(ctx context.Context, addressID int64) (*model.AddressModel, error)
	UpdateAddress(ctx context.Context, model *model.AddressModel) error
}

type addressRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) AddressRepository {
	return &addressRepository{
		db: db,
	}
}
```

---

## `internal/repository/address/create_address.go`

Digite em `internal/repository/address/create_address.go`:

```go
package address

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *addressRepository) CreateAddress(ctx context.Context, model *model.AddressModel) (int64, error) {
	query := `INSERT INTO addresses (
		zip_code, street, number, complement, district, city, state, latitude, longitude, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.ZipCode,
		model.Street,
		model.Number,
		model.Complement,
		model.District,
		model.City,
		model.State,
		model.Latitude,
		model.Longitude,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	addressID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return addressID, nil
}
```

---

## `internal/repository/address/get_address_by_id.go`

Digite em `internal/repository/address/get_address_by_id.go`:

```go
package address

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *addressRepository) GetAddressByID(ctx context.Context, addressID int64) (*model.AddressModel, error) {
	query := `SELECT id, zip_code, street, number, complement, district, city, state, latitude, longitude, created_at, updated_at
	FROM addresses
	WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, addressID)

	var result model.AddressModel
	err := row.Scan(
		&result.ID,
		&result.ZipCode,
		&result.Street,
		&result.Number,
		&result.Complement,
		&result.District,
		&result.City,
		&result.State,
		&result.Latitude,
		&result.Longitude,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}
```

---

## `internal/repository/address/update_address.go`

Digite em `internal/repository/address/update_address.go`:

```go
package address

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *addressRepository) UpdateAddress(ctx context.Context, model *model.AddressModel) error {
	query := `UPDATE addresses
	SET zip_code = ?, street = ?, number = ?, complement = ?, district = ?, city = ?, state = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(
		ctx,
		query,
		model.ZipCode,
		model.Street,
		model.Number,
		model.Complement,
		model.District,
		model.City,
		model.State,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/repository/order/repository.go`

Digite em `internal/repository/order/repository.go`:

```go
package order

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, model *model.OrderModel) (int64, error)
	GetOrderByID(ctx context.Context, orderID int64) (*model.OrderModel, error)
	GetAllOrders(ctx context.Context) ([]model.OrderModel, error)
	UpdateOrder(ctx context.Context, model *model.OrderModel) error
	UpdateOrderStatus(ctx context.Context, orderID int64, status string, updatedAt sql.NullTime) error
}

type orderRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}
```

---

## `internal/repository/order/create_order.go`

Digite em `internal/repository/order/create_order.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *orderRepository) CreateOrder(ctx context.Context, model *model.OrderModel) (int64, error) {
	query := `INSERT INTO orders (
		code, buyer_id, delivery_address_id, status, total_amount, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.Code,
		model.BuyerID,
		model.DeliveryAddressID,
		model.Status,
		model.TotalAmount,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return orderID, nil
}
```

---

## `internal/repository/order/get_order_by_id.go`

Digite em `internal/repository/order/get_order_by_id.go`:

```go
package order

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *orderRepository) GetOrderByID(ctx context.Context, orderID int64) (*model.OrderModel, error) {
	query := `SELECT id, code, buyer_id, delivery_address_id, status, total_amount, created_at, updated_at
	FROM orders
	WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, orderID)

	var result model.OrderModel
	err := row.Scan(
		&result.ID,
		&result.Code,
		&result.BuyerID,
		&result.DeliveryAddressID,
		&result.Status,
		&result.TotalAmount,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}
```

---

## `internal/repository/order/get_all_orders.go`

Digite em `internal/repository/order/get_all_orders.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]model.OrderModel, error) {
	query := `SELECT id, code, buyer_id, delivery_address_id, status, total_amount, created_at, updated_at
	FROM orders
	ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.OrderModel, 0)
	for rows.Next() {
		var order model.OrderModel
		if err := rows.Scan(
			&order.ID,
			&order.Code,
			&order.BuyerID,
			&order.DeliveryAddressID,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
```

---

## `internal/repository/order/update_order.go`

Digite em `internal/repository/order/update_order.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *orderRepository) UpdateOrder(ctx context.Context, model *model.OrderModel) error {
	query := `UPDATE orders
	SET total_amount = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(
		ctx,
		query,
		model.TotalAmount,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/repository/order/update_order_status.go`

Digite em `internal/repository/order/update_order_status.go`:

```go
package order

import (
	"context"
	"database/sql"
)

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string, updatedAt sql.NullTime) error {
	query := `UPDATE orders
	SET status = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, status, updatedAt, orderID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/service/order/service.go`

O `service` de `order` e o primeiro que depende de mais de um repository.

Digite em `internal/service/order/service.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	addressrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	buyerrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	orderrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
)

type OrderService interface {
	Create(ctx context.Context, req *dto.CreateOrderRequest) (int64, error)
	GetByID(ctx context.Context, orderID int64) (*dto.OrderResponse, error)
	GetAll(ctx context.Context) ([]dto.OrderResponse, error)
	Update(ctx context.Context, orderID int64, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	MarkReadyForDelivery(ctx context.Context, orderID int64) (*dto.OrderResponse, error)
	Cancel(ctx context.Context, orderID int64) (*dto.OrderResponse, error)
}

type orderService struct {
	orderRepo   orderrepo.OrderRepository
	buyerRepo   buyerrepo.BuyerRepository
	addressRepo addressrepo.AddressRepository
}

func NewService(
	orderRepo orderrepo.OrderRepository,
	buyerRepo buyerrepo.BuyerRepository,
	addressRepo addressrepo.AddressRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		buyerRepo:   buyerRepo,
		addressRepo: addressRepo,
	}
}
```

### Por Que Esse Arquivo E Importante

Aqui fica clara a ideia de orquestracao:

- `order` depende de `buyer`
- `order` depende de `address`
- o `service` monta o fluxo inteiro

Esse e exatamente o tipo de responsabilidade que nao cabe no `handler`.

---

## `internal/service/order/create_order.go`

Digite em `internal/service/order/create_order.go`:

```go
package order

import (
	"context"
	"fmt"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *orderService) Create(ctx context.Context, req *dto.CreateOrderRequest) (int64, error) {
	buyerModel, err := s.buyerRepo.GetBuyerByID(ctx, req.BuyerID)
	if err != nil {
		return 0, apperror.Internal("failed to get buyer before order creation", err)
	}
	if buyerModel == nil {
		return 0, apperror.NotFound("buyer not found")
	}

	now := time.Now()

	addressModel := &model.AddressModel{
		ZipCode:    req.DeliveryAddress.ZipCode,
		Street:     req.DeliveryAddress.Street,
		Number:     req.DeliveryAddress.Number,
		Complement: req.DeliveryAddress.Complement,
		District:   req.DeliveryAddress.District,
		City:       req.DeliveryAddress.City,
		State:      req.DeliveryAddress.State,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	addressID, err := s.addressRepo.CreateAddress(ctx, addressModel)
	if err != nil {
		return 0, apperror.Internal("failed to create delivery address", err)
	}

	orderModel := &model.OrderModel{
		Code:              fmt.Sprintf("ORD-%d", now.UnixNano()),
		BuyerID:           req.BuyerID,
		DeliveryAddressID: addressID,
		Status:            "created",
		TotalAmount:       req.TotalAmount,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	orderID, err := s.orderRepo.CreateOrder(ctx, orderModel)
	if err != nil {
		return 0, apperror.Internal("failed to create order", err)
	}

	return orderID, nil
}
```

### Regra De Negocio Deste Metodo

O `service` decide:

- se o `buyer` existe
- que o `status` inicial e `created`
- como gerar o `code`
- que o endereco precisa ser persistido antes do pedido

Essas decisoes nao pertencem ao `handler`.

---

## `internal/service/order/get_order_by_id.go`

Digite em `internal/service/order/get_order_by_id.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *orderService) GetByID(ctx context.Context, orderID int64) (*dto.OrderResponse, error) {
	orderModel, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("failed to get order", err)
	}
	if orderModel == nil {
		return nil, apperror.NotFound("order not found")
	}

	addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
	if err != nil {
		return nil, apperror.Internal("failed to get order address", err)
	}
	if addressModel == nil {
		return nil, apperror.Internal("order address not found", nil)
	}

	return &dto.OrderResponse{
		ID:          orderModel.ID,
		Code:        orderModel.Code,
		BuyerID:     orderModel.BuyerID,
		Status:      orderModel.Status,
		TotalAmount: orderModel.TotalAmount,
		DeliveryAddress: dto.OrderAddressResponse{
			ID:         addressModel.ID,
			ZipCode:    addressModel.ZipCode,
			Street:     addressModel.Street,
			Number:     addressModel.Number,
			Complement: addressModel.Complement,
			District:   addressModel.District,
			City:       addressModel.City,
			State:      addressModel.State,
			Latitude:   addressModel.Latitude,
			Longitude:  addressModel.Longitude,
			CreatedAt:  addressModel.CreatedAt,
			UpdatedAt:  addressModel.UpdatedAt,
		},
		CreatedAt: orderModel.CreatedAt,
		UpdatedAt: orderModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/order/get_all_orders.go`

Digite em `internal/service/order/get_all_orders.go`:

```go
package order

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *orderService) GetAll(ctx context.Context) ([]dto.OrderResponse, error) {
	orders, err := s.orderRepo.GetAllOrders(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to get orders", err)
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for _, orderModel := range orders {
		addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
		if err != nil {
			return nil, apperror.Internal("failed to get address while listing orders", err)
		}
		if addressModel == nil {
			return nil, apperror.Internal("address not found while listing orders", nil)
		}

		response = append(response, dto.OrderResponse{
			ID:          orderModel.ID,
			Code:        orderModel.Code,
			BuyerID:     orderModel.BuyerID,
			Status:      orderModel.Status,
			TotalAmount: orderModel.TotalAmount,
			DeliveryAddress: dto.OrderAddressResponse{
				ID:         addressModel.ID,
				ZipCode:    addressModel.ZipCode,
				Street:     addressModel.Street,
				Number:     addressModel.Number,
				Complement: addressModel.Complement,
				District:   addressModel.District,
				City:       addressModel.City,
				State:      addressModel.State,
				Latitude:   addressModel.Latitude,
				Longitude:  addressModel.Longitude,
				CreatedAt:  addressModel.CreatedAt,
				UpdatedAt:  addressModel.UpdatedAt,
			},
			CreatedAt: orderModel.CreatedAt,
			UpdatedAt: orderModel.UpdatedAt,
		})
	}

	return response, nil
}
```

### Observacao

Esse exemplo e simples e didatico, mas pode gerar um `N+1` ao listar muitos pedidos.

Para a fase 1 esta aceitavel.

Mais tarde, voce pode otimizar isso com:

- `JOIN`
- query dedicada para listagem

---

## `internal/service/order/update_order.go`

Digite em `internal/service/order/update_order.go`:

```go
package order

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *orderService) Update(ctx context.Context, orderID int64, req *dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	orderModel, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("failed to get order before update", err)
	}
	if orderModel == nil {
		return nil, apperror.NotFound("order not found")
	}

	if orderModel.Status != "created" {
		return nil, apperror.BadRequest("only created orders can be updated")
	}

	addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
	if err != nil {
		return nil, apperror.Internal("failed to get order address before update", err)
	}
	if addressModel == nil {
		return nil, apperror.Internal("order address not found", nil)
	}

	now := time.Now()

	addressModel.ZipCode = req.DeliveryAddress.ZipCode
	addressModel.Street = req.DeliveryAddress.Street
	addressModel.Number = req.DeliveryAddress.Number
	addressModel.Complement = req.DeliveryAddress.Complement
	addressModel.District = req.DeliveryAddress.District
	addressModel.City = req.DeliveryAddress.City
	addressModel.State = req.DeliveryAddress.State
	addressModel.UpdatedAt = now

	if err := s.addressRepo.UpdateAddress(ctx, addressModel); err != nil {
		return nil, apperror.Internal("failed to update order address", err)
	}

	orderModel.TotalAmount = req.TotalAmount
	orderModel.UpdatedAt = now

	if err := s.orderRepo.UpdateOrder(ctx, &model.OrderModel{
		ID:                orderModel.ID,
		Code:              orderModel.Code,
		BuyerID:           orderModel.BuyerID,
		DeliveryAddressID: orderModel.DeliveryAddressID,
		Status:            orderModel.Status,
		TotalAmount:       orderModel.TotalAmount,
		CreatedAt:         orderModel.CreatedAt,
		UpdatedAt:         orderModel.UpdatedAt,
	}); err != nil {
		return nil, apperror.Internal("failed to update order", err)
	}

	return &dto.OrderResponse{
		ID:          orderModel.ID,
		Code:        orderModel.Code,
		BuyerID:     orderModel.BuyerID,
		Status:      orderModel.Status,
		TotalAmount: orderModel.TotalAmount,
		DeliveryAddress: dto.OrderAddressResponse{
			ID:         addressModel.ID,
			ZipCode:    addressModel.ZipCode,
			Street:     addressModel.Street,
			Number:     addressModel.Number,
			Complement: addressModel.Complement,
			District:   addressModel.District,
			City:       addressModel.City,
			State:      addressModel.State,
			Latitude:   addressModel.Latitude,
			Longitude:  addressModel.Longitude,
			CreatedAt:  addressModel.CreatedAt,
			UpdatedAt:  addressModel.UpdatedAt,
		},
		CreatedAt: orderModel.CreatedAt,
		UpdatedAt: orderModel.UpdatedAt,
	}, nil
}
```

### Regra De Negocio Deste Metodo

Aqui o `service` decide:

- pedido precisa existir
- apenas pedido `created` pode ser atualizado
- endereco e pedido precisam ser atualizados juntos

Esse e um excelente exemplo de coordenacao de negocio entre tabelas diferentes.

---

## `internal/service/order/mark_order_ready_for_delivery.go`

Digite em `internal/service/order/mark_order_ready_for_delivery.go`:

```go
package order

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *orderService) MarkReadyForDelivery(ctx context.Context, orderID int64) (*dto.OrderResponse, error) {
	orderModel, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("failed to get order before ready-for-delivery", err)
	}
	if orderModel == nil {
		return nil, apperror.NotFound("order not found")
	}

	if orderModel.Status != "created" {
		return nil, apperror.BadRequest("only created orders can move to ready_for_delivery")
	}

	now := time.Now()
	if err := s.orderRepo.UpdateOrderStatus(
		ctx,
		orderID,
		"ready_for_delivery",
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed to update order status to ready_for_delivery", err)
	}

	orderModel.Status = "ready_for_delivery"
	orderModel.UpdatedAt = now

	addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
	if err != nil {
		return nil, apperror.Internal("failed to get order address after ready-for-delivery", err)
	}
	if addressModel == nil {
		return nil, apperror.Internal("order address not found", nil)
	}

	return &dto.OrderResponse{
		ID:          orderModel.ID,
		Code:        orderModel.Code,
		BuyerID:     orderModel.BuyerID,
		Status:      orderModel.Status,
		TotalAmount: orderModel.TotalAmount,
		DeliveryAddress: dto.OrderAddressResponse{
			ID:         addressModel.ID,
			ZipCode:    addressModel.ZipCode,
			Street:     addressModel.Street,
			Number:     addressModel.Number,
			Complement: addressModel.Complement,
			District:   addressModel.District,
			City:       addressModel.City,
			State:      addressModel.State,
			Latitude:   addressModel.Latitude,
			Longitude:  addressModel.Longitude,
			CreatedAt:  addressModel.CreatedAt,
			UpdatedAt:  addressModel.UpdatedAt,
		},
		CreatedAt: orderModel.CreatedAt,
		UpdatedAt: orderModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/order/cancel_order.go`

Digite em `internal/service/order/cancel_order.go`:

```go
package order

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *orderService) Cancel(ctx context.Context, orderID int64) (*dto.OrderResponse, error) {
	orderModel, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal("failed to get order before cancel", err)
	}
	if orderModel == nil {
		return nil, apperror.NotFound("order not found")
	}

	if orderModel.Status != "created" {
		return nil, apperror.BadRequest("only created orders can be cancelled")
	}

	now := time.Now()
	if err := s.orderRepo.UpdateOrderStatus(
		ctx,
		orderID,
		"cancelled",
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed to cancel order", err)
	}

	orderModel.Status = "cancelled"
	orderModel.UpdatedAt = now

	addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
	if err != nil {
		return nil, apperror.Internal("failed to get order address after cancel", err)
	}
	if addressModel == nil {
		return nil, apperror.Internal("order address not found", nil)
	}

	return &dto.OrderResponse{
		ID:          orderModel.ID,
		Code:        orderModel.Code,
		BuyerID:     orderModel.BuyerID,
		Status:      orderModel.Status,
		TotalAmount: orderModel.TotalAmount,
		DeliveryAddress: dto.OrderAddressResponse{
			ID:         addressModel.ID,
			ZipCode:    addressModel.ZipCode,
			Street:     addressModel.Street,
			Number:     addressModel.Number,
			Complement: addressModel.Complement,
			District:   addressModel.District,
			City:       addressModel.City,
			State:      addressModel.State,
			Latitude:   addressModel.Latitude,
			Longitude:  addressModel.Longitude,
			CreatedAt:  addressModel.CreatedAt,
			UpdatedAt:  addressModel.UpdatedAt,
		},
		CreatedAt: orderModel.CreatedAt,
		UpdatedAt: orderModel.UpdatedAt,
	}, nil
}
```

---

## `internal/handler/order/handler.go`

Digite em `internal/handler/order/handler.go`:

```go
package order

import (
	orderservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/order"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router       gin.IRouter
	validate     *validator.Validate
	orderService orderservice.OrderService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, orderService orderservice.OrderService) *Handler {
	return &Handler{
		router:       router,
		validate:     validate,
		orderService: orderService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/orders", h.CreateOrder)
	h.router.GET("/orders", h.GetAllOrders)
	h.router.GET("/orders/:order_id", h.GetOrderByID)
	h.router.PUT("/orders/:order_id", h.UpdateOrder)
	h.router.PATCH("/orders/:order_id/ready-for-delivery", h.MarkReadyForDelivery)
	h.router.PATCH("/orders/:order_id/cancel", h.CancelOrder)
}
```

---

## `internal/handler/order/create_order.go`

Digite em `internal/handler/order/create_order.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest

	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	orderID, err := h.orderService.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": orderID,
	})
}
```

---

## `internal/handler/order/get_order_by_id.go`

Digite em `internal/handler/order/get_order_by_id.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetOrderByID(c *gin.Context) {
	orderID, ok := httpresponse.ParseInt64Param(c, "order_id")
	if !ok {
		return
	}

	response, err := h.orderService.GetByID(c.Request.Context(), orderID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/order/get_all_orders.go`

Digite em `internal/handler/order/get_all_orders.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllOrders(c *gin.Context) {
	response, err := h.orderService.GetAll(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/order/update_order.go`

Digite em `internal/handler/order/update_order.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateOrder(c *gin.Context) {
	orderID, ok := httpresponse.ParseInt64Param(c, "order_id")
	if !ok {
		return
	}

	var req dto.UpdateOrderRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	response, err := h.orderService.Update(c.Request.Context(), orderID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/order/mark_order_ready_for_delivery.go`

Digite em `internal/handler/order/mark_order_ready_for_delivery.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) MarkReadyForDelivery(c *gin.Context) {
	orderID, ok := httpresponse.ParseInt64Param(c, "order_id")
	if !ok {
		return
	}

	response, err := h.orderService.MarkReadyForDelivery(c.Request.Context(), orderID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/order/cancel_order.go`

Digite em `internal/handler/order/cancel_order.go`:

```go
package order

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CancelOrder(c *gin.Context) {
	orderID, ok := httpresponse.ParseInt64Param(c, "order_id")
	if !ok {
		return
	}

	response, err := h.orderService.Cancel(c.Request.Context(), orderID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/server/router.go`

Agora o `router` central precisa ligar quatro blocos:

- `buyer`
- `address`
- `order`
- `courier`

Adicione nos imports de `internal/server/router.go`:

```go
import (
	addressrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	courierhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/courier"
	orderhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/order"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	courierrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	orderrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
	orderservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/order"
)
```

E, dentro de `NewRouter(...)`:

```go
db := deps.DB
buyerRepo := buyerrepository.NewRepository(db)
buyerService := buyerservice.NewService(buyerRepo)
buyerHandler := buyerhandler.NewHandler(router, validate, buyerService)
buyerHandler.RouteList()

courierRepo := courierrepository.NewRepository(db)
courierService := courierservice.NewService(courierRepo)
courierHandler := courierhandler.NewHandler(router, validate, courierService)
courierHandler.RouteList()

addressRepo := addressrepository.NewRepository(db)
orderRepo := orderrepository.NewRepository(db)
orderService := orderservice.NewService(orderRepo, buyerRepo, addressRepo)
orderHandler := orderhandler.NewHandler(router, validate, orderService)
orderHandler.RouteList()
```

---

## Resumo Do Fluxo De Orders

### Criacao

1. handler le e valida o JSON
2. service verifica se o `buyer` existe
3. service cria o endereco
4. service gera `code` e define status `created`
5. repository persiste o pedido
6. handler responde `201`

### Busca Por ID

1. handler le `order_id`
2. service busca o pedido
3. service busca o endereco vinculado
4. service monta a resposta composta
5. handler responde `200`

### Listagem

1. handler chama `GetAll`
2. service lista pedidos
3. service busca endereco de cada pedido
4. service monta a resposta
5. handler responde `200`

### Atualizacao

1. handler le `order_id`
2. handler valida o body
3. service verifica existencia
4. service verifica se o pedido ainda esta `created`
5. service atualiza endereco e total
6. handler responde `200`

### Ready For Delivery

1. handler le `order_id`
2. service verifica existencia
3. service valida transicao de status
4. repository atualiza `status`
5. handler responde `200`

### Cancelamento

1. handler le `order_id`
2. service verifica existencia
3. service valida se ainda pode cancelar
4. repository atualiza `status`
5. handler responde `200`

---

## Pontos De Log Recomendados No CRUD De Orders

Quando voce for incrementar logs nesse modulo, a recomendacao mais util costuma ser:

- `handler`: logar erro de bind, validacao e contexto HTTP relevante
- `service`: logar ausencia de `buyer`, transicao invalida de status, tentativa de atualizacao indevida e cancelamento importante
- `repository`: logar falha inesperada ao persistir pedido, endereco ou atualizar status

Por operacao:

- criacao: logar falha ao localizar `buyer`, falha ao criar pedido e sucesso com `order_id` e `code`
- busca por ID: logar ausencia quando isso ajudar no diagnostico
- listagem: geralmente o middleware HTTP ja cobre bem
- atualizacao: logar tentativa de atualizar pedido fora do status permitido
- ready for delivery: logar transicao de status relevante
- cancelamento: logar motivo operacional quando houver bloqueio de cancelamento

Evite registrar sem necessidade:

- endereco completo em texto puro
- payload bruto inteiro
- dados pessoais do comprador sem mascaramento

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos:

- `internal/dto/order_dto.go`
- `internal/model/address_model.go`
- `internal/model/order_model.go`
- `internal/repository/address/repository.go`
- `internal/repository/address/create_address.go`
- `internal/repository/address/get_address_by_id.go`
- `internal/repository/address/update_address.go`
- `internal/repository/order/repository.go`
- `internal/repository/order/create_order.go`
- `internal/repository/order/get_order_by_id.go`
- `internal/repository/order/get_all_orders.go`
- `internal/repository/order/update_order.go`
- `internal/repository/order/update_order_status.go`
- `internal/service/order/service.go`
- `internal/service/order/create_order.go`
- `internal/service/order/get_order_by_id.go`
- `internal/service/order/get_all_orders.go`
- `internal/service/order/update_order.go`
- `internal/service/order/mark_order_ready_for_delivery.go`
- `internal/service/order/cancel_order.go`
- `internal/handler/order/handler.go`
- `internal/handler/order/create_order.go`
- `internal/handler/order/get_order_by_id.go`
- `internal/handler/order/get_all_orders.go`
- `internal/handler/order/update_order.go`
- `internal/handler/order/mark_order_ready_for_delivery.go`
- `internal/handler/order/cancel_order.go`

---

## Observacao Final

Este capitulo e muito importante porque mostra a diferenca entre um CRUD simples e um CRUD com coordenacao de dominio.

Em `order`, o `service` ja precisa:

- validar recurso relacionado
- criar registros em mais de uma tabela
- aplicar transicao de status
- montar resposta composta

Isso aproxima bastante o projeto de um backend real de mercado.

---

## Proximo Passo

Depois deste documento, os proximos passos naturais sao:

- criar testes unitarios dos `services`
- criar testes de integracao HTTP
- partir para `deliveries`

Se quiser, no proximo passo eu posso gerar:

- um `04-crud-deliveries.md`
- ou um documento focado so em testes para `buyers`, `couriers` e `orders`
