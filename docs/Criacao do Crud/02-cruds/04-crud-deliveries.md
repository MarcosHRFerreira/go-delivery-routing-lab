# Fluxo de Deliveries: Handler, Service e Repository

## Visao Geral

Este documento segue o mesmo padrao de `buyers`, `couriers` e `orders`, mantendo a organizacao inspirada no projeto `go-tweets`.

O dominio `delivery` cuida de:

- atribuicao de pedido a entregador
- consulta da fila do entregador
- inicio da entrega
- conclusao da entrega
- falha de entrega

Fluxo:

```text
HTTP /deliveries/* -> handler/delivery -> service/delivery -> repository/delivery -> MySQL
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

Mas aqui existe uma orquestracao mais rica:

- `delivery` depende de `order`
- `delivery` depende de `courier`
- uma entrega nasce a partir de um pedido pronto para entrega

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar o fluxo de `deliveries` com:

- validacao de entrada no `handler`
- regra de negocio no `service`
- SQL isolado no `repository`
- integracao com `orders` e `couriers`

Esse modulo marca a transicao entre CRUD basico e operacao de entrega.

---

## Contrato HTTP Inicial

Para `deliveries`, o contrato sugerido e:

- `POST /deliveries/assign`
- `GET /couriers/:courier_id/deliveries`
- `PATCH /deliveries/:delivery_id/start`
- `PATCH /deliveries/:delivery_id/complete`
- `PATCH /deliveries/:delivery_id/fail`

Payload de atribuicao:

```json
{
  "order_id": 10,
  "courier_id": 3
}
```

Payload de inicio:

```json
{}
```

Payload de conclusao:

```json
{}
```

Payload de falha:

```json
{}
```

---

## Status Sugeridos

Para a fase operacional inicial, voce pode usar estes status:

- `assigned`
- `out_for_delivery`
- `completed`
- `failed`

### Regras De Negocio Sugeridas

- apenas pedido `ready_for_delivery` pode ser atribuido
- um pedido nao pode ter mais de uma entrega
- apenas entrega `assigned` pode ir para `out_for_delivery`
- apenas entrega `out_for_delivery` pode ir para `completed`
- apenas entrega `out_for_delivery` pode ir para `failed`

Essas decisoes ficam no `service`, nao no `handler`.

---

## Estrutura Recomendada

Para manter o mesmo padrao dos documentos anteriores:

```text
internal/
  dto/
    delivery_dto.go
  model/
    delivery_model.go
  handler/
    delivery/
      assign_delivery.go
      get_courier_deliveries.go
      handler.go
      start_delivery.go
      complete_delivery.go
      fail_delivery.go
  repository/
    delivery/
      create_delivery.go
      get_courier_deliveries.go
      get_delivery_by_id.go
      get_delivery_by_order_id.go
      repository.go
      update_delivery_status.go
  service/
    delivery/
      assign_delivery.go
      complete_delivery.go
      fail_delivery.go
      get_courier_deliveries.go
      service.go
      start_delivery.go
```

Arquivos compartilhados ja existentes:

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `internal/repository/order/repository.go`
- `internal/repository/courier/repository.go`
- `internal/model/order_model.go`
- `internal/model/courier_model.go`

---

## `internal/dto/delivery_dto.go`

Centralize os DTOs do dominio em um unico arquivo.

Digite em `internal/dto/delivery_dto.go`:

```go
package dto

import "time"

type (
	AssignDeliveryRequest struct {
		OrderID   int64 `json:"order_id" validate:"required"`
		CourierID int64 `json:"courier_id" validate:"required"`
	}

	DeliveryResponse struct {
		ID              int64     `json:"id"`
		OrderID         int64     `json:"order_id"`
		CourierID       int64     `json:"courier_id"`
		Status          string    `json:"status"`
		CurrentSequence *int      `json:"current_sequence,omitempty"`
		LastReorderedAt time.Time `json:"last_reordered_at,omitempty"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}
)
```

### Observacao

Aqui a validacao basica continua no DTO e no `handler`.

O `service` continua responsavel por:

- verificar se pedido existe
- verificar se entregador existe
- validar status do pedido
- validar transicoes da entrega

---

## `internal/model/delivery_model.go`

Digite em `internal/model/delivery_model.go`:

```go
package model

import "time"

type DeliveryModel struct {
	ID              int64
	OrderID         int64
	CourierID       int64
	Status          string
	CurrentSequence *int
	LastReorderedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

---

## `internal/repository/delivery/repository.go`

Digite em `internal/repository/delivery/repository.go`:

```go
package delivery

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type DeliveryRepository interface {
	CreateDelivery(ctx context.Context, model *model.DeliveryModel) (int64, error)
	GetDeliveryByID(ctx context.Context, deliveryID int64) (*model.DeliveryModel, error)
	GetDeliveryByOrderID(ctx context.Context, orderID int64) (*model.DeliveryModel, error)
	GetCourierDeliveries(ctx context.Context, courierID int64) ([]model.DeliveryModel, error)
	UpdateDeliveryStatus(ctx context.Context, deliveryID int64, status string, updatedAt sql.NullTime) error
}

type deliveryRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) DeliveryRepository {
	return &deliveryRepository{
		db: db,
	}
}
```

---

## `internal/repository/delivery/create_delivery.go`

Digite em `internal/repository/delivery/create_delivery.go`:

```go
package delivery

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *deliveryRepository) CreateDelivery(ctx context.Context, model *model.DeliveryModel) (int64, error) {
	query := `INSERT INTO deliveries (
		order_id, courier_id, status, current_sequence, last_reordered_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.OrderID,
		model.CourierID,
		model.Status,
		model.CurrentSequence,
		model.LastReorderedAt,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	deliveryID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return deliveryID, nil
}
```

---

## `internal/repository/delivery/get_delivery_by_id.go`

Digite em `internal/repository/delivery/get_delivery_by_id.go`:

```go
package delivery

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *deliveryRepository) GetDeliveryByID(ctx context.Context, deliveryID int64) (*model.DeliveryModel, error) {
	query := `SELECT id, order_id, courier_id, status, current_sequence, last_reordered_at, created_at, updated_at
	FROM deliveries
	WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, deliveryID)

	var result model.DeliveryModel
	err := row.Scan(
		&result.ID,
		&result.OrderID,
		&result.CourierID,
		&result.Status,
		&result.CurrentSequence,
		&result.LastReorderedAt,
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

## `internal/repository/delivery/get_delivery_by_order_id.go`

Digite em `internal/repository/delivery/get_delivery_by_order_id.go`:

```go
package delivery

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *deliveryRepository) GetDeliveryByOrderID(ctx context.Context, orderID int64) (*model.DeliveryModel, error) {
	query := `SELECT id, order_id, courier_id, status, current_sequence, last_reordered_at, created_at, updated_at
	FROM deliveries
	WHERE order_id = ?`

	row := r.db.QueryRowContext(ctx, query, orderID)

	var result model.DeliveryModel
	err := row.Scan(
		&result.ID,
		&result.OrderID,
		&result.CourierID,
		&result.Status,
		&result.CurrentSequence,
		&result.LastReorderedAt,
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

## `internal/repository/delivery/get_courier_deliveries.go`

Digite em `internal/repository/delivery/get_courier_deliveries.go`:

```go
package delivery

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *deliveryRepository) GetCourierDeliveries(ctx context.Context, courierID int64) ([]model.DeliveryModel, error) {
	query := `SELECT id, order_id, courier_id, status, current_sequence, last_reordered_at, created_at, updated_at
	FROM deliveries
	WHERE courier_id = ?
	ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.DeliveryModel, 0)
	for rows.Next() {
		var delivery model.DeliveryModel
		if err := rows.Scan(
			&delivery.ID,
			&delivery.OrderID,
			&delivery.CourierID,
			&delivery.Status,
			&delivery.CurrentSequence,
			&delivery.LastReorderedAt,
			&delivery.CreatedAt,
			&delivery.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, delivery)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
```

---

## `internal/repository/delivery/update_delivery_status.go`

Digite em `internal/repository/delivery/update_delivery_status.go`:

```go
package delivery

import (
	"context"
	"database/sql"
)

func (r *deliveryRepository) UpdateDeliveryStatus(ctx context.Context, deliveryID int64, status string, updatedAt sql.NullTime) error {
	query := `UPDATE deliveries
	SET status = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, status, updatedAt, deliveryID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/service/delivery/service.go`

O `service` de `delivery` depende de tres repositories.

Digite em `internal/service/delivery/service.go`:

```go
package delivery

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	courierrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	deliveryrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/delivery"
	orderrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
)

type DeliveryService interface {
	Assign(ctx context.Context, req *dto.AssignDeliveryRequest) (int64, error)
	GetCourierDeliveries(ctx context.Context, courierID int64) ([]dto.DeliveryResponse, error)
	Start(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error)
	Complete(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error)
	Fail(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error)
}

type deliveryService struct {
	deliveryRepo deliveryrepo.DeliveryRepository
	orderRepo    orderrepo.OrderRepository
	courierRepo  courierrepo.CourierRepository
}

func NewService(
	deliveryRepo deliveryrepo.DeliveryRepository,
	orderRepo orderrepo.OrderRepository,
	courierRepo courierrepo.CourierRepository,
) DeliveryService {
	return &deliveryService{
		deliveryRepo: deliveryRepo,
		orderRepo:    orderRepo,
		courierRepo:  courierRepo,
	}
}
```

---

## `internal/service/delivery/assign_delivery.go`

Digite em `internal/service/delivery/assign_delivery.go`:

```go
package delivery

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *deliveryService) Assign(ctx context.Context, req *dto.AssignDeliveryRequest) (int64, error) {
	orderModel, err := s.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		return 0, apperror.Internal("failed to get order before delivery assignment", err)
	}
	if orderModel == nil {
		return 0, apperror.NotFound("order not found")
	}

	courierModel, err := s.courierRepo.GetCourierByID(ctx, req.CourierID)
	if err != nil {
		return 0, apperror.Internal("failed to get courier before delivery assignment", err)
	}
	if courierModel == nil {
		return 0, apperror.NotFound("courier not found")
	}

	if orderModel.Status != "ready_for_delivery" {
		return 0, apperror.BadRequest("only ready_for_delivery orders can be assigned")
	}

	if courierModel.Status != "available" && courierModel.Status != "busy" {
		return 0, apperror.BadRequest("courier is not eligible for assignment")
	}

	existingDelivery, err := s.deliveryRepo.GetDeliveryByOrderID(ctx, req.OrderID)
	if err != nil {
		return 0, apperror.Internal("failed to check existing delivery for order", err)
	}
	if existingDelivery != nil {
		return 0, apperror.BadRequest("order already has a delivery")
	}

	now := time.Now()
	deliveryModel := &model.DeliveryModel{
		OrderID:         req.OrderID,
		CourierID:       req.CourierID,
		Status:          "assigned",
		CurrentSequence: nil,
		LastReorderedAt: nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	deliveryID, err := s.deliveryRepo.CreateDelivery(ctx, deliveryModel)
	if err != nil {
		return 0, apperror.Internal("failed to assign delivery", err)
	}

	return deliveryID, nil
}
```

### Regra De Negocio Deste Metodo

O `service` decide:

- se o pedido existe
- se o entregador existe
- se o pedido esta `ready_for_delivery`
- se o pedido ja possui entrega

Esse e o coracao do fluxo operacional.

---

## `internal/service/delivery/get_courier_deliveries.go`

Digite em `internal/service/delivery/get_courier_deliveries.go`:

```go
package delivery

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *deliveryService) GetCourierDeliveries(ctx context.Context, courierID int64) ([]dto.DeliveryResponse, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before listing deliveries", err)
	}
	if courierModel == nil {
		return nil, apperror.NotFound("courier not found")
	}

	deliveries, err := s.deliveryRepo.GetCourierDeliveries(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier deliveries", err)
	}

	response := make([]dto.DeliveryResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		response = append(response, dto.DeliveryResponse{
			ID:              delivery.ID,
			OrderID:         delivery.OrderID,
			CourierID:       delivery.CourierID,
			Status:          delivery.Status,
			CurrentSequence: delivery.CurrentSequence,
			CreatedAt:       delivery.CreatedAt,
			UpdatedAt:       delivery.UpdatedAt,
		})
	}

	return response, nil
}
```

---

## `internal/service/delivery/start_delivery.go`

Digite em `internal/service/delivery/start_delivery.go`:

```go
package delivery

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *deliveryService) Start(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error) {
	deliveryModel, err := s.deliveryRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return nil, apperror.Internal("failed to get delivery before start", err)
	}
	if deliveryModel == nil {
		return nil, apperror.NotFound("delivery not found")
	}

	if deliveryModel.Status != "assigned" {
		return nil, apperror.BadRequest("only assigned deliveries can be started")
	}

	now := time.Now()
	if err := s.deliveryRepo.UpdateDeliveryStatus(
		ctx,
		deliveryID,
		"out_for_delivery",
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed to start delivery", err)
	}

	deliveryModel.Status = "out_for_delivery"
	deliveryModel.UpdatedAt = now

	return &dto.DeliveryResponse{
		ID:              deliveryModel.ID,
		OrderID:         deliveryModel.OrderID,
		CourierID:       deliveryModel.CourierID,
		Status:          deliveryModel.Status,
		CurrentSequence: deliveryModel.CurrentSequence,
		CreatedAt:       deliveryModel.CreatedAt,
		UpdatedAt:       deliveryModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/delivery/complete_delivery.go`

Digite em `internal/service/delivery/complete_delivery.go`:

```go
package delivery

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *deliveryService) Complete(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error) {
	deliveryModel, err := s.deliveryRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return nil, apperror.Internal("failed to get delivery before completion", err)
	}
	if deliveryModel == nil {
		return nil, apperror.NotFound("delivery not found")
	}

	if deliveryModel.Status != "out_for_delivery" {
		return nil, apperror.BadRequest("only out_for_delivery deliveries can be completed")
	}

	now := time.Now()
	if err := s.deliveryRepo.UpdateDeliveryStatus(
		ctx,
		deliveryID,
		"completed",
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed to complete delivery", err)
	}

	deliveryModel.Status = "completed"
	deliveryModel.UpdatedAt = now

	return &dto.DeliveryResponse{
		ID:              deliveryModel.ID,
		OrderID:         deliveryModel.OrderID,
		CourierID:       deliveryModel.CourierID,
		Status:          deliveryModel.Status,
		CurrentSequence: deliveryModel.CurrentSequence,
		CreatedAt:       deliveryModel.CreatedAt,
		UpdatedAt:       deliveryModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/delivery/fail_delivery.go`

Digite em `internal/service/delivery/fail_delivery.go`:

```go
package delivery

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *deliveryService) Fail(ctx context.Context, deliveryID int64) (*dto.DeliveryResponse, error) {
	deliveryModel, err := s.deliveryRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return nil, apperror.Internal("failed to get delivery before fail", err)
	}
	if deliveryModel == nil {
		return nil, apperror.NotFound("delivery not found")
	}

	if deliveryModel.Status != "out_for_delivery" {
		return nil, apperror.BadRequest("only out_for_delivery deliveries can fail")
	}

	now := time.Now()
	if err := s.deliveryRepo.UpdateDeliveryStatus(
		ctx,
		deliveryID,
		"failed",
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed delivery status update", err)
	}

	deliveryModel.Status = "failed"
	deliveryModel.UpdatedAt = now

	return &dto.DeliveryResponse{
		ID:              deliveryModel.ID,
		OrderID:         deliveryModel.OrderID,
		CourierID:       deliveryModel.CourierID,
		Status:          deliveryModel.Status,
		CurrentSequence: deliveryModel.CurrentSequence,
		CreatedAt:       deliveryModel.CreatedAt,
		UpdatedAt:       deliveryModel.UpdatedAt,
	}, nil
}
```

---

## `internal/handler/delivery/handler.go`

Digite em `internal/handler/delivery/handler.go`:

```go
package delivery

import (
	deliveryservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/delivery"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router          gin.IRouter
	validate        *validator.Validate
	deliveryService deliveryservice.DeliveryService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, deliveryService deliveryservice.DeliveryService) *Handler {
	return &Handler{
		router:          router,
		validate:        validate,
		deliveryService: deliveryService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/deliveries/assign", h.AssignDelivery)
	h.router.GET("/couriers/:courier_id/deliveries", h.GetCourierDeliveries)
	h.router.PATCH("/deliveries/:delivery_id/start", h.StartDelivery)
	h.router.PATCH("/deliveries/:delivery_id/complete", h.CompleteDelivery)
	h.router.PATCH("/deliveries/:delivery_id/fail", h.FailDelivery)
}
```

---

## `internal/handler/delivery/assign_delivery.go`

Digite em `internal/handler/delivery/assign_delivery.go`:

```go
package delivery

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AssignDelivery(c *gin.Context) {
	var req dto.AssignDeliveryRequest

	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	deliveryID, err := h.deliveryService.Assign(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": deliveryID,
	})
}
```

---

## `internal/handler/delivery/get_courier_deliveries.go`

Digite em `internal/handler/delivery/get_courier_deliveries.go`:

```go
package delivery

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCourierDeliveries(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	response, err := h.deliveryService.GetCourierDeliveries(c.Request.Context(), courierID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/delivery/start_delivery.go`

Digite em `internal/handler/delivery/start_delivery.go`:

```go
package delivery

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) StartDelivery(c *gin.Context) {
	deliveryID, ok := httpresponse.ParseInt64Param(c, "delivery_id")
	if !ok {
		return
	}

	response, err := h.deliveryService.Start(c.Request.Context(), deliveryID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/delivery/complete_delivery.go`

Digite em `internal/handler/delivery/complete_delivery.go`:

```go
package delivery

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CompleteDelivery(c *gin.Context) {
	deliveryID, ok := httpresponse.ParseInt64Param(c, "delivery_id")
	if !ok {
		return
	}

	response, err := h.deliveryService.Complete(c.Request.Context(), deliveryID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/delivery/fail_delivery.go`

Digite em `internal/handler/delivery/fail_delivery.go`:

```go
package delivery

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) FailDelivery(c *gin.Context) {
	deliveryID, ok := httpresponse.ParseInt64Param(c, "delivery_id")
	if !ok {
		return
	}

	response, err := h.deliveryService.Fail(c.Request.Context(), deliveryID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/server/router.go`

Agora o `router` central passa a ligar tambem `delivery`.

Adicione nos imports de `internal/server/router.go`:

```go
import (
	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	courierhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/courier"
	deliveryhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/delivery"
	orderhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/order"
	addressrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	courierrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	deliveryrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/delivery"
	orderrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
	deliveryservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/delivery"
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

deliveryRepo := deliveryrepository.NewRepository(db)
deliveryService := deliveryservice.NewService(deliveryRepo, orderRepo, courierRepo)
deliveryHandler := deliveryhandler.NewHandler(router, validate, deliveryService)
deliveryHandler.RouteList()
```

---

## Resumo Do Fluxo De Deliveries

### Atribuicao

1. handler le e valida o JSON
2. service verifica se o pedido existe
3. service verifica se o entregador existe
4. service valida se o pedido esta `ready_for_delivery`
5. service valida que o pedido ainda nao tem entrega
6. repository cria a entrega
7. handler responde `201`

### Fila Do Entregador

1. handler le `courier_id`
2. service verifica se o entregador existe
3. repository lista entregas do entregador
4. handler responde `200`

### Inicio

1. handler le `delivery_id`
2. service verifica se a entrega existe
3. service valida se o status e `assigned`
4. repository atualiza para `out_for_delivery`
5. handler responde `200`

### Conclusao

1. handler le `delivery_id`
2. service verifica se a entrega existe
3. service valida se o status e `out_for_delivery`
4. repository atualiza para `completed`
5. handler responde `200`

### Falha

1. handler le `delivery_id`
2. service verifica se a entrega existe
3. service valida se o status e `out_for_delivery`
4. repository atualiza para `failed`
5. handler responde `200`

---

## Pontos De Log Recomendados No CRUD De Deliveries

Quando voce for implementar logs nesse modulo, vale distribuir assim:

- `handler`: logar erro de bind, validacao e falha inesperada de request
- `service`: logar ausencia de pedido ou entregador, tentativa de atribuicao invalida e mudancas importantes de status
- `repository`: logar falha inesperada ao criar entrega, listar fila ou atualizar status

Por operacao:

- atribuicao: logar bloqueio por pedido ja entregue, pedido sem status correto ou entregador inexistente
- fila do entregador: normalmente o middleware HTTP ja cobre bem a parte tecnica
- inicio: logar transicao para `out_for_delivery`
- conclusao: logar transicao para `completed`
- falha: logar transicao para `failed` e contexto operacional quando fizer sentido

Evite em log:

- repetir o mesmo erro em todas as camadas
- payload completo com dados sensiveis
- detalhes desnecessarios do endereco do cliente

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos:

- `internal/dto/delivery_dto.go`
- `internal/model/delivery_model.go`
- `internal/repository/delivery/repository.go`
- `internal/repository/delivery/create_delivery.go`
- `internal/repository/delivery/get_delivery_by_id.go`
- `internal/repository/delivery/get_delivery_by_order_id.go`
- `internal/repository/delivery/get_courier_deliveries.go`
- `internal/repository/delivery/update_delivery_status.go`
- `internal/service/delivery/service.go`
- `internal/service/delivery/assign_delivery.go`
- `internal/service/delivery/get_courier_deliveries.go`
- `internal/service/delivery/start_delivery.go`
- `internal/service/delivery/complete_delivery.go`
- `internal/service/delivery/fail_delivery.go`
- `internal/handler/delivery/handler.go`
- `internal/handler/delivery/assign_delivery.go`
- `internal/handler/delivery/get_courier_deliveries.go`
- `internal/handler/delivery/start_delivery.go`
- `internal/handler/delivery/complete_delivery.go`
- `internal/handler/delivery/fail_delivery.go`

---

## Observacao Final

Este modulo consolida a ideia de backend orientado a operacao.

Agora o sistema ja demonstra:

- transicao de status
- coordenacao entre modulos
- restricoes de negocio reais
- fila operacional por entregador

Depois dele, o passo mais natural deixa de ser CRUD puro e passa a ser:

- localizacao
- reordenacao
- testes de integracao

---

## Proximo Passo

Depois deste documento, os proximos passos mais coerentes sao:

- documentar `courier_locations`
- documentar roteirizacao e reorder
- gerar um guia de testes para todos os fluxos

Se quiser, no proximo passo eu posso gerar:

- `05-crud-courier-locations.md`
- ou um documento so de testes da fase 1
