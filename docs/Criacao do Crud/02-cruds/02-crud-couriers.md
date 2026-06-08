# Fluxo de Couriers: Handler, Service e Repository

## Visao Geral

Este documento segue o mesmo padrao usado em `01-crud-buyers.md` e foi organizado no estilo do projeto `go-tweets`.

O dominio `courier` cuida de:

- criacao de entregador
- busca por ID
- listagem
- atualizacao de dados cadastrais
- atualizacao de status
- exclusao

Fluxo:

```text
HTTP /couriers/* -> handler/courier -> service/courier -> repository/courier -> MySQL
```

## Atualizacao Do Bootstrap

No estado atual do projeto, a base HTTP ja foi padronizada com:

- `Gin`
- `validator`
- `internal/httpresponse`
- `internal/apperror`
- `internal/server/router.go`

Entao, se algum trecho antigo abaixo ainda mencionar `ServeMux`, `net/http` puro ou registro principal em `cmd/main.go`, adapte mentalmente para o padrao vigente:

- handlers com `*gin.Context`
- construtores usando `gin.IRouter`
- bind e validacao via `internal/httpresponse`
- ligacao das rotas em `internal/server/router.go`

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar o CRUD de `couriers` com:

- validacao de entrada no `handler`
- regra de negocio no `service`
- SQL isolado no `repository`
- contrato HTTP coerente com a fase 1

Este guia mostra:

- quais arquivos criar
- onde digitar cada trecho
- como dividir responsabilidades
- como tratar o status do entregador sem poluir o `handler`

---

## Estrutura Recomendada

Para manter o mesmo padrao do `buyers`, a estrutura de `courier` pode ficar assim:

```text
internal/
  dto/
    courier_dto.go
  model/
    courier_model.go
  handler/
    courier/
      create_courier.go
      delete_courier.go
      get_all_couriers.go
      get_courier_by_id.go
      handler.go
      update_courier.go
      update_courier_status.go
  repository/
    courier/
      create_courier.go
      delete_courier.go
      get_all_couriers.go
      get_courier_by_id.go
      get_courier_by_phone.go
      repository.go
      update_courier.go
      update_courier_status.go
  service/
    courier/
      create_courier.go
      delete_courier.go
      get_all_couriers.go
      get_courier_by_id.go
      service.go
      update_courier.go
      update_courier_status.go
```

Arquivos compartilhados ja existentes do fluxo anterior:

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`

---

## Contrato HTTP Inicial

Para o CRUD de `couriers`, o contrato sugerido e:

- `POST /couriers`
- `GET /couriers/:courier_id`
- `GET /couriers`
- `PUT /couriers/:courier_id`
- `PATCH /couriers/:courier_id/status`
- `DELETE /couriers/:courier_id`

Payload de criacao:

```json
{
  "name": "Carlos Souza",
  "phone": "11977777777",
  "vehicle_type": "motorcycle",
  "status": "available"
}
```

Payload de atualizacao:

```json
{
  "name": "Carlos Souza Lima",
  "phone": "11977777777",
  "vehicle_type": "motorcycle"
}
```

Payload de atualizacao de status:

```json
{
  "status": "busy"
}
```

### Status Sugeridos

Para simplificar a fase 1, voce pode trabalhar com:

- `available`
- `busy`
- `inactive`

Se quiser, no futuro voce pode evoluir esse campo para enum de dominio mais rico.

---

## `internal/dto/courier_dto.go`

No mesmo estilo do `go-tweets`, concentre os DTOs do dominio em um unico arquivo.

Digite em `internal/dto/courier_dto.go`:

```go
package dto

import "time"

type (
	CreateCourierRequest struct {
		Name        string `json:"name" validate:"required,min=3"`
		Phone       string `json:"phone" validate:"required"`
		VehicleType string `json:"vehicle_type" validate:"required"`
		Status      string `json:"status" validate:"required,oneof=available busy inactive"`
	}

	UpdateCourierRequest struct {
		Name        string `json:"name" validate:"required,min=3"`
		Phone       string `json:"phone" validate:"required"`
		VehicleType string `json:"vehicle_type" validate:"required"`
	}

	UpdateCourierStatusRequest struct {
		Status string `json:"status" validate:"required,oneof=available busy inactive"`
	}

	CourierResponse struct {
		ID          int64     `json:"id"`
		Name        string    `json:"name"`
		Phone       string    `json:"phone"`
		VehicleType string    `json:"vehicle_type"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
)
```

### Observacao

Assim como no documento de `buyers`, a validacao basica fica no DTO e no `handler`.

O `service` deve ficar limpo de validacoes como:

- campo obrigatorio
- `min`
- `oneof`

### Normalizacao Recomendada Neste CRUD

Para `couriers`, a normalizacao mais util deve acontecer no inicio de `Create`, `Update` e `UpdateStatus`, antes de:

- verificar telefone duplicado
- comparar status
- montar o `CourierModel`

Aplicacao sugerida:

- `name`: aplicar `strings.TrimSpace(...)`
- `phone`: remover mascara e manter apenas digitos
- `vehicle_type`: aplicar `strings.TrimSpace(...)` e, se fizer sentido no projeto, `strings.ToLower(...)`
- `status`: aplicar `strings.TrimSpace(...)` e `strings.ToLower(...)` quando vier do cliente

Boa regra pratica:

- o `handler` continua fazendo bind e validate
- o `service` normaliza o que sera comparado ou persistido
- o banco protege a unicidade final do telefone com `UNIQUE`

---

## `internal/model/courier_model.go`

Digite em `internal/model/courier_model.go`:

```go
package model

import "time"

type CourierModel struct {
	ID          int64
	Name        string
	Phone       string
	VehicleType string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

### Papel Do Model

- representar o registro persistido
- trafegar entre `service` e `repository`
- manter o DTO desacoplado da persistencia

---

## `internal/repository/courier/repository.go`

Digite em `internal/repository/courier/repository.go`:

```go
package courier

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type CourierRepository interface {
	CreateCourier(ctx context.Context, model *model.CourierModel) (int64, error)
	GetCourierByID(ctx context.Context, courierID int64) (*model.CourierModel, error)
	GetAllCouriers(ctx context.Context) ([]model.CourierModel, error)
	UpdateCourier(ctx context.Context, model *model.CourierModel) error
	UpdateCourierStatus(ctx context.Context, courierID int64, status string, updatedAt sql.NullTime) error
	DeleteCourier(ctx context.Context, courierID int64) error
}

type courierRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) CourierRepository {
	return &courierRepository{
		db: db,
	}
}
```

### Observacao Sobre Unicidade

Como o schema passa a garantir `UNIQUE` para `couriers.phone`, este CRUD nao precisa de uma consulta previa de telefone duplicado no `repository`.

O fluxo recomendado fica assim:

- o `service` normaliza os valores
- o `repository` tenta persistir
- o banco barra duplicidade
- o `service` traduz o erro tecnico em `409 Conflict`

---

## `internal/repository/courier/create_courier.go`

Digite em `internal/repository/courier/create_courier.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *courierRepository) CreateCourier(ctx context.Context, model *model.CourierModel) (int64, error) {
	query := `INSERT INTO couriers (name, phone, vehicle_type, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.Phone,
		model.VehicleType,
		model.Status,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	courierID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return courierID, nil
}
```

---

## `internal/repository/courier/get_courier_by_id.go`

Digite em `internal/repository/courier/get_courier_by_id.go`:

```go
package courier

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *courierRepository) GetCourierByID(ctx context.Context, courierID int64) (*model.CourierModel, error) {
	query := `SELECT id, name, phone, vehicle_type, status, created_at, updated_at
	FROM couriers WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, courierID)

	var result model.CourierModel
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Phone,
		&result.VehicleType,
		&result.Status,
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

## `internal/repository/courier/get_all_couriers.go`

Digite em `internal/repository/courier/get_all_couriers.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *courierRepository) GetAllCouriers(ctx context.Context) ([]model.CourierModel, error) {
	query := `SELECT id, name, phone, vehicle_type, status, created_at, updated_at
	FROM couriers
	ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.CourierModel, 0)
	for rows.Next() {
		var courier model.CourierModel
		if err := rows.Scan(
			&courier.ID,
			&courier.Name,
			&courier.Phone,
			&courier.VehicleType,
			&courier.Status,
			&courier.CreatedAt,
			&courier.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, courier)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
```

---

## `internal/repository/courier/update_courier.go`

Digite em `internal/repository/courier/update_courier.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *courierRepository) UpdateCourier(ctx context.Context, model *model.CourierModel) error {
	query := `UPDATE couriers
	SET name = ?, phone = ?, vehicle_type = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.Phone,
		model.VehicleType,
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

## `internal/repository/courier/update_courier_status.go`

Digite em `internal/repository/courier/update_courier_status.go`:

```go
package courier

import (
	"context"
	"database/sql"
)

func (r *courierRepository) UpdateCourierStatus(ctx context.Context, courierID int64, status string, updatedAt sql.NullTime) error {
	query := `UPDATE couriers
	SET status = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, status, updatedAt, courierID)
	if err != nil {
		return err
	}

	return nil
}
```

### Nota

Se voce preferir simplificar, pode trocar `sql.NullTime` por `time.Time`.

Mantive essa variacao aqui apenas para te lembrar que o repository decide o formato exato de persistencia, nao o `handler`.

---

## `internal/repository/courier/delete_courier.go`

Digite em `internal/repository/courier/delete_courier.go`:

```go
package courier

import "context"

func (r *courierRepository) DeleteCourier(ctx context.Context, courierID int64) error {
	query := `DELETE FROM couriers WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, courierID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/service/courier/service.go`

Digite em `internal/service/courier/service.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	courierrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
)

type CourierService interface {
	Create(ctx context.Context, req *dto.CreateCourierRequest) (int64, error)
	GetByID(ctx context.Context, courierID int64) (*dto.CourierResponse, error)
	GetAll(ctx context.Context) ([]dto.CourierResponse, error)
	Update(ctx context.Context, courierID int64, req *dto.UpdateCourierRequest) (*dto.CourierResponse, error)
	UpdateStatus(ctx context.Context, courierID int64, req *dto.UpdateCourierStatusRequest) (*dto.CourierResponse, error)
	Delete(ctx context.Context, courierID int64) error
}

type courierService struct {
	courierRepo courierrepo.CourierRepository
}

func NewService(courierRepo courierrepo.CourierRepository) CourierService {
	return &courierService{
		courierRepo: courierRepo,
	}
}
```

---

## `internal/service/courier/create_courier.go`

Digite em `internal/service/courier/create_courier.go`:

```go
package courier

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

func (s *courierService) Create(ctx context.Context, req *dto.CreateCourierRequest) (int64, error) {
	normalizedName := strings.TrimSpace(req.Name)
	normalizedPhone := onlyDigits(req.Phone)
	normalizedVehicleType := strings.TrimSpace(strings.ToLower(req.VehicleType))
	normalizedStatus := strings.TrimSpace(strings.ToLower(req.Status))

	now := time.Now()
	courierModel := &model.CourierModel{
		Name:        normalizedName,
		Phone:       normalizedPhone,
		VehicleType: normalizedVehicleType,
		Status:      normalizedStatus,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	courierID, err := s.courierRepo.CreateCourier(ctx, courierModel)
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uq_couriers_phone") {
			return 0, apperror.Conflict("courier phone already exists")
		}

		return 0, apperror.Internal("failed to create courier", err)
	}

	return courierID, nil
}
```

### Regra De Negocio Deste Metodo

O `service` nao valida `required` nem `oneof`.

Aqui ele decide uma regra de negocio:

- telefone precisa ser persistido de forma normalizada
- conflito de telefone deve virar `409 Conflict` quando o banco rejeitar a operacao

### Ordem Recomendada Dentro Do Fluxo

1. handler faz bind e validate do request
2. service normaliza `name`, `phone`, `vehicle_type` e `status`
3. service monta o `CourierModel`
4. repository tenta persistir
5. banco garante unicidade com `UNIQUE`
6. service traduz erro de duplicidade em `409 Conflict`

Esse e exatamente o tipo de responsabilidade que deve ficar no `service`.

Os helpers `onlyDigits(...)` e um eventual `mapCourierConflictError(...)` podem ficar em um arquivo auxiliar da pasta `internal/service/courier/`.

---

## `internal/service/courier/get_courier_by_id.go`

Digite em `internal/service/courier/get_courier_by_id.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *courierService) GetByID(ctx context.Context, courierID int64) (*dto.CourierResponse, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier", err)
	}
	if courierModel == nil {
		return nil, apperror.NotFound("courier not found")
	}

	return &dto.CourierResponse{
		ID:          courierModel.ID,
		Name:        courierModel.Name,
		Phone:       courierModel.Phone,
		VehicleType: courierModel.VehicleType,
		Status:      courierModel.Status,
		CreatedAt:   courierModel.CreatedAt,
		UpdatedAt:   courierModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/courier/get_all_couriers.go`

Digite em `internal/service/courier/get_all_couriers.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *courierService) GetAll(ctx context.Context) ([]dto.CourierResponse, error) {
	couriers, err := s.courierRepo.GetAllCouriers(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to get couriers", err)
	}

	response := make([]dto.CourierResponse, 0, len(couriers))
	for _, courier := range couriers {
		response = append(response, dto.CourierResponse{
			ID:          courier.ID,
			Name:        courier.Name,
			Phone:       courier.Phone,
			VehicleType: courier.VehicleType,
			Status:      courier.Status,
			CreatedAt:   courier.CreatedAt,
			UpdatedAt:   courier.UpdatedAt,
		})
	}

	return response, nil
}
```

---

## `internal/service/courier/update_courier.go`

Digite em `internal/service/courier/update_courier.go`:

```go
package courier

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
	mysqlDriver "github.com/go-sql-driver/mysql"
)

func (s *courierService) Update(ctx context.Context, courierID int64, req *dto.UpdateCourierRequest) (*dto.CourierResponse, error) {
	existingCourier, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before update", err)
	}
	if existingCourier == nil {
		return nil, apperror.NotFound("courier not found")
	}

	existingCourier.Name = strings.TrimSpace(req.Name)
	existingCourier.Phone = onlyDigits(req.Phone)
	existingCourier.VehicleType = strings.TrimSpace(strings.ToLower(req.VehicleType))
	existingCourier.UpdatedAt = time.Now()

	if err := s.courierRepo.UpdateCourier(ctx, &model.CourierModel{
		ID:          existingCourier.ID,
		Name:        existingCourier.Name,
		Phone:       existingCourier.Phone,
		VehicleType: existingCourier.VehicleType,
		Status:      existingCourier.Status,
		CreatedAt:   existingCourier.CreatedAt,
		UpdatedAt:   existingCourier.UpdatedAt,
	}); err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "uq_couriers_phone") {
			return nil, apperror.Conflict("courier phone already exists")
		}

		return nil, apperror.Internal("failed to update courier", err)
	}

	return &dto.CourierResponse{
		ID:          existingCourier.ID,
		Name:        existingCourier.Name,
		Phone:       existingCourier.Phone,
		VehicleType: existingCourier.VehicleType,
		Status:      existingCourier.Status,
		CreatedAt:   existingCourier.CreatedAt,
		UpdatedAt:   existingCourier.UpdatedAt,
	}, nil
}
```

### Regra De Negocio Deste Metodo

Mesmo com a validacao basica no `handler`, ainda ficam no `service`:

- verificacao de existencia do recurso
- normalizacao dos campos antes de persistir
- traducao do erro de `UNIQUE` para `409 Conflict`
- coordenacao entre leitura e escrita

---

## `internal/service/courier/update_courier_status.go`

Digite em `internal/service/courier/update_courier_status.go`:

```go
package courier

import (
	"context"
	"database/sql"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *courierService) UpdateStatus(ctx context.Context, courierID int64, req *dto.UpdateCourierStatusRequest) (*dto.CourierResponse, error) {
	existingCourier, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before status update", err)
	}
	if existingCourier == nil {
		return nil, apperror.NotFound("courier not found")
	}

	// Example of business restriction:
	// once inactive, the courier can only go back to available.
	if existingCourier.Status == "inactive" && req.Status == "busy" {
		return nil, apperror.BadRequest("inactive courier cannot move directly to busy")
	}

	now := time.Now()
	if err := s.courierRepo.UpdateCourierStatus(
		ctx,
		courierID,
		req.Status,
		sql.NullTime{Time: now, Valid: true},
	); err != nil {
		return nil, apperror.Internal("failed to update courier status", err)
	}

	existingCourier.Status = req.Status
	existingCourier.UpdatedAt = now

	return &dto.CourierResponse{
		ID:          existingCourier.ID,
		Name:        existingCourier.Name,
		Phone:       existingCourier.Phone,
		VehicleType: existingCourier.VehicleType,
		Status:      existingCourier.Status,
		CreatedAt:   existingCourier.CreatedAt,
		UpdatedAt:   existingCourier.UpdatedAt,
	}, nil
}
```

### Regra De Negocio Deste Metodo

Aqui aparece um exemplo melhor de regra de negocio do que `required`:

- um entregador `inactive` nao vai direto para `busy`

Voce pode mudar essa regra depois, mas o importante e perceber que esse tipo de decisao deve morar no `service`.

---

## `internal/service/courier/delete_courier.go`

Digite em `internal/service/courier/delete_courier.go`:

```go
package courier

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
)

func (s *courierService) Delete(ctx context.Context, courierID int64) error {
	existingCourier, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return apperror.Internal("failed to get courier before delete", err)
	}
	if existingCourier == nil {
		return apperror.NotFound("courier not found")
	}

	if err := s.courierRepo.DeleteCourier(ctx, courierID); err != nil {
		return apperror.Internal("failed to delete courier", err)
	}

	return nil
}
```

---

## `internal/handler/courier/handler.go`

Digite em `internal/handler/courier/handler.go`:

```go
package courier

import (
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router         gin.IRouter
	validate       *validator.Validate
	courierService courierservice.CourierService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, courierService courierservice.CourierService) *Handler {
	return &Handler{
		router:         router,
		validate:       validate,
		courierService: courierService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/couriers", h.CreateCourier)
	h.router.GET("/couriers", h.GetAllCouriers)
	h.router.GET("/couriers/:courier_id", h.GetCourierByID)
	h.router.PUT("/couriers/:courier_id", h.UpdateCourier)
	h.router.PATCH("/couriers/:courier_id/status", h.UpdateCourierStatus)
	h.router.DELETE("/couriers/:courier_id", h.DeleteCourier)
}
```

---

## `internal/handler/courier/create_courier.go`

Digite em `internal/handler/courier/create_courier.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateCourier(c *gin.Context) {
	var req dto.CreateCourierRequest

	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	courierID, err := h.courierService.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": courierID,
	})
}
```

---

## `internal/handler/courier/get_courier_by_id.go`

Digite em `internal/handler/courier/get_courier_by_id.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCourierByID(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	response, err := h.courierService.GetByID(c.Request.Context(), courierID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/courier/get_all_couriers.go`

Digite em `internal/handler/courier/get_all_couriers.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllCouriers(c *gin.Context) {
	response, err := h.courierService.GetAll(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/courier/update_courier.go`

Digite em `internal/handler/courier/update_courier.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateCourier(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	var req dto.UpdateCourierRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	response, err := h.courierService.Update(c.Request.Context(), courierID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/courier/update_courier_status.go`

Digite em `internal/handler/courier/update_courier_status.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateCourierStatus(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	var req dto.UpdateCourierStatusRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	response, err := h.courierService.UpdateStatus(c.Request.Context(), courierID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/courier/delete_courier.go`

Digite em `internal/handler/courier/delete_courier.go`:

```go
package courier

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteCourier(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	if err := h.courierService.Delete(c.Request.Context(), courierID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
```

---

## `internal/server/router.go`

Se voce ja montou `buyers`, agora basta estender o bootstrap central do projeto.

Adicione em `internal/server/router.go`:

```go
import (
	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	courierhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/courier"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	courierrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
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
```

---

## Resumo Do Fluxo De Couriers

### Criacao

1. handler le e valida o JSON
2. service verifica duplicidade de telefone
3. service monta `CourierModel`
4. repository executa `INSERT`
5. handler responde `201`

### Busca Por ID

1. handler le `courier_id`
2. service chama repository
3. repository faz `SELECT`
4. service transforma ausencia em `404`
5. handler responde `200`

### Listagem

1. handler chama `GetAll`
2. repository faz `SELECT`
3. service monta a resposta
4. handler responde `200`

### Atualizacao

1. handler le `courier_id`
2. handler le e valida o body
3. service verifica existencia
4. service verifica telefone duplicado
5. repository executa `UPDATE`
6. handler responde `200`

### Atualizacao De Status

1. handler le `courier_id`
2. handler valida o body
3. service verifica existencia
4. service aplica regra de transicao de status
5. repository executa `UPDATE`
6. handler responde `200`

### Exclusao

1. handler le `courier_id`
2. service verifica existencia
3. repository executa `DELETE`
4. handler responde `204`

---

## Pontos De Log Recomendados No CRUD De Couriers

Quando voce for implementar logging nesse modulo, a distribuicao mais equilibrada costuma ser:

- `handler`: logar erro de bind, validacao e falha inesperada na borda HTTP
- `service`: logar conflito de telefone, transicao invalida de status e ausencia relevante de recurso
- `repository`: logar falha inesperada de banco em consultas e comandos de persistencia

Para cada operacao, vale pensar assim:

- criacao: logar conflito de telefone e sucesso com `courier_id`
- busca por ID: logar ausencia quando isso ajudar o diagnostico
- listagem: normalmente o middleware HTTP ja cobre bem
- atualizacao: logar conflito de telefone e tentativa sobre recurso inexistente
- atualizacao de status: logar transicao invalida ou mudanca importante de estado
- exclusao: logar ausencia e falha de persistencia

Evite expor em log:

- telefone completo sem necessidade
- payload completo da request
- dados sensiveis ou internos que nao ajudam no suporte

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos:

- `internal/dto/courier_dto.go`
- `internal/model/courier_model.go`
- `internal/repository/courier/repository.go`
- `internal/repository/courier/create_courier.go`
- `internal/repository/courier/get_courier_by_id.go`
- `internal/repository/courier/get_courier_by_phone.go`
- `internal/repository/courier/get_all_couriers.go`
- `internal/repository/courier/update_courier.go`
- `internal/repository/courier/update_courier_status.go`
- `internal/repository/courier/delete_courier.go`
- `internal/service/courier/service.go`
- `internal/service/courier/create_courier.go`
- `internal/service/courier/get_courier_by_id.go`
- `internal/service/courier/get_all_couriers.go`
- `internal/service/courier/update_courier.go`
- `internal/service/courier/update_courier_status.go`
- `internal/service/courier/delete_courier.go`
- `internal/handler/courier/handler.go`
- `internal/handler/courier/create_courier.go`
- `internal/handler/courier/get_courier_by_id.go`
- `internal/handler/courier/get_all_couriers.go`
- `internal/handler/courier/update_courier.go`
- `internal/handler/courier/update_courier_status.go`
- `internal/handler/courier/delete_courier.go`

---

## Observacao Final

O ponto mais importante deste capitulo e perceber a diferenca entre:

- validacao de entrada
- regra de negocio

No caso de `courier`, exemplos de validacao de entrada:

- `vehicle_type` obrigatorio
- `status` dentro de `available busy inactive`
- `name` com tamanho minimo

Exemplos de regra de negocio:

- telefone nao pode estar duplicado
- entregador inexistente nao pode ser atualizado
- entregador `inactive` nao pode ir direto para `busy`

Essa separacao deixa o `service` mais limpo e mais facil de testar.

---

## Proximo Passo

Depois deste documento, o proximo mais coerente e:

- `docs/Criacao do Crud/02-cruds/03-crud-orders.md`

No mesmo formato:

- visao geral
- fluxo por camada
- arquivos a criar
- codigo por arquivo
- validacao no `handler`
- regra de negocio no `service`
