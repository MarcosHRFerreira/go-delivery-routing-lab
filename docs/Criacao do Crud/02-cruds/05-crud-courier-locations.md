# Fluxo de Courier Locations: Handler, Service e Repository

## Visao Geral

Este documento segue o mesmo padrao dos modulos anteriores e prepara a base para a futura roteirizacao.

O dominio `courier location` cuida de:

- registrar a localizacao atual do entregador
- consultar a ultima localizacao registrada

Fluxo:

```text
HTTP /couriers/:courier_id/location/* -> handler/location -> service/location -> repository/location -> MySQL
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

Esse modulo parece simples, mas e muito importante porque:

- fornece a posicao atual do entregador
- apoia futuras regras de reordenacao
- introduce validacao geografica basica

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar o fluxo de localizacao com:

- validacao de entrada no `handler`
- regra de negocio no `service`
- SQL isolado no `repository`
- integracao com `couriers`

---

## Contrato HTTP Inicial

Para `courier_locations`, o contrato sugerido e:

- `POST /couriers/:courier_id/location`
- `GET /couriers/:courier_id/location/latest`

Payload de registro:

```json
{
  "latitude": -23.55052,
  "longitude": -46.633308
}
```

Resposta de consulta da ultima localizacao:

```json
{
  "id": 10,
  "courier_id": 3,
  "latitude": -23.55052,
  "longitude": -46.633308,
  "recorded_at": "2026-06-01T12:00:00Z"
}
```

---

## Regras De Negocio Sugeridas

Para a fase inicial:

- o entregador precisa existir para registrar localizacao
- `latitude` precisa estar entre `-90` e `90`
- `longitude` precisa estar entre `-180` e `180`
- a consulta de ultima localizacao deve retornar o registro mais recente

Essas regras ficam no `service`.

As validacoes estruturais de payload continuam no `handler` via DTO.

---

## Estrutura Recomendada

```text
internal/
  dto/
    courier_location_dto.go
  model/
    courier_location_model.go
  handler/
    location/
      get_latest_courier_location.go
      handler.go
      register_courier_location.go
  repository/
    location/
      get_latest_courier_location.go
      repository.go
      store_courier_location.go
  service/
    location/
      get_latest_courier_location.go
      register_courier_location.go
      service.go
```

Arquivos compartilhados ja existentes:

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `internal/repository/courier/repository.go`

---

## `internal/dto/courier_location_dto.go`

Digite em `internal/dto/courier_location_dto.go`:

```go
package dto

import "time"

type (
	RegisterCourierLocationRequest struct {
		Latitude  float64 `json:"latitude" validate:"required"`
		Longitude float64 `json:"longitude" validate:"required"`
	}

	CourierLocationResponse struct {
		ID         int64     `json:"id"`
		CourierID  int64     `json:"courier_id"`
		Latitude   float64   `json:"latitude"`
		Longitude  float64   `json:"longitude"`
		RecordedAt time.Time `json:"recorded_at"`
	}
)
```

### Observacao

O `validate:"required"` em `float64` nem sempre e suficiente para cenarios mais rigorosos.

Por isso, a validacao geografica real continua no `service`:

- faixa de latitude
- faixa de longitude

---

## `internal/model/courier_location_model.go`

Digite em `internal/model/courier_location_model.go`:

```go
package model

import "time"

type CourierLocationModel struct {
	ID         int64
	CourierID  int64
	Latitude   float64
	Longitude  float64
	RecordedAt time.Time
}
```

---

## `internal/repository/location/repository.go`

Digite em `internal/repository/location/repository.go`:

```go
package location

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type LocationRepository interface {
	StoreCourierLocation(ctx context.Context, model *model.CourierLocationModel) (int64, error)
	GetLatestCourierLocation(ctx context.Context, courierID int64) (*model.CourierLocationModel, error)
}

type locationRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) LocationRepository {
	return &locationRepository{
		db: db,
	}
}
```

---

## `internal/repository/location/store_courier_location.go`

Digite em `internal/repository/location/store_courier_location.go`:

```go
package location

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *locationRepository) StoreCourierLocation(ctx context.Context, model *model.CourierLocationModel) (int64, error) {
	query := `INSERT INTO courier_locations (courier_id, latitude, longitude, recorded_at)
	VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.CourierID,
		model.Latitude,
		model.Longitude,
		model.RecordedAt,
	)
	if err != nil {
		return 0, err
	}

	locationID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return locationID, nil
}
```

---

## `internal/repository/location/get_latest_courier_location.go`

Digite em `internal/repository/location/get_latest_courier_location.go`:

```go
package location

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *locationRepository) GetLatestCourierLocation(ctx context.Context, courierID int64) (*model.CourierLocationModel, error) {
	query := `SELECT id, courier_id, latitude, longitude, recorded_at
	FROM courier_locations
	WHERE courier_id = ?
	ORDER BY recorded_at DESC, id DESC
	LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, courierID)

	var result model.CourierLocationModel
	err := row.Scan(
		&result.ID,
		&result.CourierID,
		&result.Latitude,
		&result.Longitude,
		&result.RecordedAt,
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

### Ponto Importante

Ordenar por `recorded_at DESC, id DESC` ajuda quando dois registros tiverem timestamps muito proximos.

---

## `internal/service/location/service.go`

Digite em `internal/service/location/service.go`:

```go
package location

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	courierrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	locationrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/location"
)

type LocationService interface {
	Register(ctx context.Context, courierID int64, req *dto.RegisterCourierLocationRequest) (int64, error)
	GetLatest(ctx context.Context, courierID int64) (*dto.CourierLocationResponse, error)
}

type locationService struct {
	locationRepo locationrepo.LocationRepository
	courierRepo  courierrepo.CourierRepository
}

func NewService(
	locationRepo locationrepo.LocationRepository,
	courierRepo courierrepo.CourierRepository,
) LocationService {
	return &locationService{
		locationRepo: locationRepo,
		courierRepo:  courierRepo,
	}
}
```

---

## `internal/service/location/register_courier_location.go`

Digite em `internal/service/location/register_courier_location.go`:

```go
package location

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *locationService) Register(ctx context.Context, courierID int64, req *dto.RegisterCourierLocationRequest) (int64, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return 0, apperror.Internal("failed to get courier before location register", err)
	}
	if courierModel == nil {
		return 0, apperror.NotFound("courier not found")
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		return 0, apperror.BadRequest("latitude must be between -90 and 90")
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		return 0, apperror.BadRequest("longitude must be between -180 and 180")
	}

	locationModel := &model.CourierLocationModel{
		CourierID:  courierID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		RecordedAt: time.Now(),
	}

	locationID, err := s.locationRepo.StoreCourierLocation(ctx, locationModel)
	if err != nil {
		return 0, apperror.Internal("failed to store courier location", err)
	}

	return locationID, nil
}
```

### Regra De Negocio Deste Metodo

O `service` decide:

- se o entregador existe
- se latitude esta em faixa valida
- se longitude esta em faixa valida

Esse tipo de regra nao pertence ao `handler`.

---

## `internal/service/location/get_latest_courier_location.go`

Digite em `internal/service/location/get_latest_courier_location.go`:

```go
package location

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *locationService) GetLatest(ctx context.Context, courierID int64) (*dto.CourierLocationResponse, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before loading latest location", err)
	}
	if courierModel == nil {
		return nil, apperror.NotFound("courier not found")
	}

	locationModel, err := s.locationRepo.GetLatestCourierLocation(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get latest courier location", err)
	}
	if locationModel == nil {
		return nil, apperror.NotFound("latest courier location not found")
	}

	return &dto.CourierLocationResponse{
		ID:         locationModel.ID,
		CourierID:  locationModel.CourierID,
		Latitude:   locationModel.Latitude,
		Longitude:  locationModel.Longitude,
		RecordedAt: locationModel.RecordedAt,
	}, nil
}
```

---

## `internal/handler/location/handler.go`

Digite em `internal/handler/location/handler.go`:

```go
package location

import (
	locationservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/location"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router          gin.IRouter
	validate        *validator.Validate
	locationService locationservice.LocationService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, locationService locationservice.LocationService) *Handler {
	return &Handler{
		router:          router,
		validate:        validate,
		locationService: locationService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/couriers/:courier_id/location", h.RegisterCourierLocation)
	h.router.GET("/couriers/:courier_id/location/latest", h.GetLatestCourierLocation)
}
```

---

## `internal/handler/location/register_courier_location.go`

Digite em `internal/handler/location/register_courier_location.go`:

```go
package location

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterCourierLocation(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	var req dto.RegisterCourierLocationRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	locationID, err := h.locationService.Register(c.Request.Context(), courierID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": locationID,
	})
}
```

---

## `internal/handler/location/get_latest_courier_location.go`

Digite em `internal/handler/location/get_latest_courier_location.go`:

```go
package location

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLatestCourierLocation(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	response, err := h.locationService.GetLatest(c.Request.Context(), courierID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/server/router.go`

Agora o `router` central passa a ligar tambem `location`.

Adicione nos imports de `internal/server/router.go`:

```go
import (
	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	courierhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/courier"
	deliveryhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/delivery"
	locationhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/location"
	orderhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/order"
	addressrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	courierrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	deliveryrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/delivery"
	locationrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/location"
	orderrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
	deliveryservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/delivery"
	locationservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/location"
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

locationRepo := locationrepository.NewRepository(db)
locationService := locationservice.NewService(locationRepo, courierRepo)
locationHandler := locationhandler.NewHandler(router, validate, locationService)
locationHandler.RouteList()
```

---

## Resumo Do Fluxo De Courier Locations

### Registro

1. handler le `courier_id`
2. handler valida o body
3. service verifica se o entregador existe
4. service valida latitude e longitude
5. repository grava a localizacao
6. handler responde `201`

### Ultima Localizacao

1. handler le `courier_id`
2. service verifica se o entregador existe
3. repository busca a ultima localizacao
4. handler responde `200`

---

## Pontos De Log Recomendados Em Courier Locations

Quando voce for adicionar logs nesse modulo, a estrategia mais util costuma ser:

- `handler`: logar erro de bind, validacao e falha inesperada de request
- `service`: logar entregador inexistente, coordenada invalida e ausencia de localizacao quando isso for relevante
- `repository`: logar falha inesperada ao gravar ou consultar localizacao

Por operacao:

- registro: logar tentativa com latitude ou longitude invalida e sucesso com `courier_id`
- ultima localizacao: logar ausencia de registro apenas quando isso ajudar no diagnostico

Evite:

- logar coordenadas com precisao excessiva sem necessidade
- logar cada consulta saudavel em alto volume fora do middleware
- duplicar o mesmo contexto em varias camadas

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos:

- `internal/dto/courier_location_dto.go`
- `internal/model/courier_location_model.go`
- `internal/repository/location/repository.go`
- `internal/repository/location/store_courier_location.go`
- `internal/repository/location/get_latest_courier_location.go`
- `internal/service/location/service.go`
- `internal/service/location/register_courier_location.go`
- `internal/service/location/get_latest_courier_location.go`
- `internal/handler/location/handler.go`
- `internal/handler/location/register_courier_location.go`
- `internal/handler/location/get_latest_courier_location.go`

---

## Observacao Final

Esse modulo parece pequeno, mas ele fecha uma parte critica do dominio:

- operacionalmente, o sistema passa a conhecer a posicao do entregador
- arquiteturalmente, ele prepara a base para reorder e roteirizacao

Depois dele, o proximo passo deixa de ser CRUD tradicional e passa a ser logica de classificacao e sequenciamento.

---

## Proximo Passo

Depois deste documento, os proximos passos mais coerentes sao:

- documentar reorder de entregas
- documentar `delivery_reorder_history`
- gerar guia de testes da fase 1

Se quiser, no proximo passo eu posso gerar:

- `06-delivery-reorder.md`
- ou um documento unico de testes para `buyers`, `couriers`, `orders`, `deliveries` e `locations`
