# Fluxo de Delivery Reorder: Handler, Service e Repository

## Visao Geral

Este documento segue o mesmo padrao dos modulos anteriores, mas agora entra em um fluxo mais algoritmico do sistema.

O dominio `delivery reorder` cuida de:

- recalcular a ordem sugerida das entregas de um entregador
- persistir a sequencia calculada em `deliveries.current_sequence`
- registrar historico da reclassificacao em `delivery_reorder_history`
- expor a sequencia atual do entregador

Fluxo:

```text
HTTP /couriers/:courier_id/deliveries/* -> handler/routing -> service/routing -> repositories -> MySQL
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

Esse modulo e importante porque ele conecta:

- entregador
- localizacao atual
- fila de entregas
- heuristica de ordenacao
- historico operacional

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar um fluxo de reorder com:

- validacao de entrada no `handler`
- regra de negocio e algoritmo no `service`
- persistencia de sequencia e historico nos `repository`
- retorno de uma lista ordenada para o app do entregador

---

## Contrato HTTP Inicial

Para a fase inicial de reorder, o contrato sugerido e:

- `POST /couriers/:courier_id/deliveries/reorder`
- `GET /couriers/:courier_id/deliveries/sequence`

Payload do reorder:

```json
{}
```

Resposta do reorder:

```json
{
  "courier_id": 3,
  "generated_at": "2026-06-01T12:00:00Z",
  "deliveries": [
    {
      "delivery_id": 100,
      "order_id": 10,
      "sequence": 1,
      "zip_code": "01001000",
      "number": "120",
      "score": 1000.012
    },
    {
      "delivery_id": 101,
      "order_id": 11,
      "sequence": 2,
      "zip_code": "01002000",
      "number": "85",
      "score": 2000.085
    }
  ]
}
```

Resposta da consulta da sequencia:

```json
{
  "courier_id": 3,
  "deliveries": [
    {
      "delivery_id": 100,
      "order_id": 10,
      "sequence": 1
    },
    {
      "delivery_id": 101,
      "order_id": 11,
      "sequence": 2
    }
  ]
}
```

---

## Regras De Negocio Sugeridas

Para a fase inicial:

- o entregador precisa existir
- o entregador precisa ter localizacao mais recente disponivel
- apenas entregas `assigned` participam do reorder
- pedidos cancelados ou ja concluidos nao entram
- a heuristica inicial usa `CEP + numero`
- ao final da reclassificacao, a ordem e persistida em `deliveries.current_sequence`
- cada recalculo gera historico em `delivery_reorder_history`

### Observacao Importante

Mesmo usando localizacao atual do entregador como parte do contexto, o algoritmo MVP ainda pode usar apenas a heuristica:

- ordenar por `zip_code`
- em empate, ordenar pelo numero do endereco convertido para inteiro

Isso ja e suficiente como primeiro passo didatico.

---

## Estrutura Recomendada

```text
internal/
  dto/
    routing_dto.go
  handler/
    routing/
      get_delivery_sequence.go
      handler.go
      reorder_deliveries.go
  repository/
    delivery/
      get_courier_pending_deliveries.go
      update_delivery_sequence.go
      update_delivery_last_reordered_at.go
    reorderhistory/
      repository.go
      store_reorder_history.go
  service/
    routing/
      get_delivery_sequence.go
      reorder_deliveries.go
      service.go
```

Arquivos compartilhados ja existentes:

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `internal/repository/courier/repository.go`
- `internal/repository/location/repository.go`
- `internal/repository/delivery/repository.go`
- `internal/repository/order/repository.go`
- `internal/repository/address/repository.go`

---

## `internal/dto/routing_dto.go`

Digite em `internal/dto/routing_dto.go`:

```go
package dto

import "time"

type (
	DeliverySequenceItemResponse struct {
		DeliveryID int64   `json:"delivery_id"`
		OrderID    int64   `json:"order_id"`
		Sequence   int     `json:"sequence"`
		ZipCode    string  `json:"zip_code,omitempty"`
		Number     string  `json:"number,omitempty"`
		Score      float64 `json:"score,omitempty"`
	}

	ReorderDeliveriesResponse struct {
		CourierID   int64                          `json:"courier_id"`
		GeneratedAt time.Time                      `json:"generated_at"`
		Deliveries  []DeliverySequenceItemResponse `json:"deliveries"`
	}

	GetDeliverySequenceResponse struct {
		CourierID  int64                          `json:"courier_id"`
		Deliveries []DeliverySequenceItemResponse `json:"deliveries"`
	}
)
```

---

## `internal/repository/reorderhistory/repository.go`

Digite em `internal/repository/reorderhistory/repository.go`:

```go
package reorderhistory

import (
	"context"
	"database/sql"
	"time"
)

type ReorderHistoryRepository interface {
	StoreReorderHistory(ctx context.Context, courierID int64, deliveryID int64, sequencePosition int, score float64, reason string, generatedAt time.Time) error
}

type reorderHistoryRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) ReorderHistoryRepository {
	return &reorderHistoryRepository{
		db: db,
	}
}
```

---

## `internal/repository/reorderhistory/store_reorder_history.go`

Digite em `internal/repository/reorderhistory/store_reorder_history.go`:

```go
package reorderhistory

import (
	"context"
	"time"
)

func (r *reorderHistoryRepository) StoreReorderHistory(
	ctx context.Context,
	courierID int64,
	deliveryID int64,
	sequencePosition int,
	score float64,
	reason string,
	generatedAt time.Time,
) error {
	query := `INSERT INTO delivery_reorder_history (
		courier_id, delivery_id, sequence_position, score, reason, generated_at
	) VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		courierID,
		deliveryID,
		sequencePosition,
		score,
		reason,
		generatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
```

---

## Extensao de `internal/repository/delivery/repository.go`

Adicione estes metodos na interface `DeliveryRepository`:

```go
GetCourierPendingDeliveries(ctx context.Context, courierID int64) ([]model.DeliveryModel, error)
UpdateDeliverySequence(ctx context.Context, deliveryID int64, sequence int) error
UpdateDeliveryLastReorderedAt(ctx context.Context, deliveryID int64, lastReorderedAt sql.NullTime) error
```

### Observacao

Esses metodos sao especificos do fluxo de reorder e podem ficar no mesmo repository de `delivery`.

Isso evita criar um repository duplicado so para a mesma tabela.

---

## `internal/repository/delivery/get_courier_pending_deliveries.go`

Digite em `internal/repository/delivery/get_courier_pending_deliveries.go`:

```go
package delivery

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *deliveryRepository) GetCourierPendingDeliveries(ctx context.Context, courierID int64) ([]model.DeliveryModel, error) {
	query := `SELECT id, order_id, courier_id, status, current_sequence, last_reordered_at, created_at, updated_at
	FROM deliveries
	WHERE courier_id = ? AND status = 'assigned'
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

## `internal/repository/delivery/update_delivery_sequence.go`

Digite em `internal/repository/delivery/update_delivery_sequence.go`:

```go
package delivery

import "context"

func (r *deliveryRepository) UpdateDeliverySequence(ctx context.Context, deliveryID int64, sequence int) error {
	query := `UPDATE deliveries
	SET current_sequence = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, sequence, deliveryID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/repository/delivery/update_delivery_last_reordered_at.go`

Digite em `internal/repository/delivery/update_delivery_last_reordered_at.go`:

```go
package delivery

import (
	"context"
	"database/sql"
)

func (r *deliveryRepository) UpdateDeliveryLastReorderedAt(ctx context.Context, deliveryID int64, lastReorderedAt sql.NullTime) error {
	query := `UPDATE deliveries
	SET last_reordered_at = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, lastReorderedAt, lastReorderedAt, deliveryID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/service/routing/service.go`

Digite em `internal/service/routing/service.go`:

```go
package routing

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	addressrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	courierrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	deliveryrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/delivery"
	locationrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/location"
	orderrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
	reorderhistoryrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/reorderhistory"
)

type RoutingService interface {
	ReorderDeliveries(ctx context.Context, courierID int64) (*dto.ReorderDeliveriesResponse, error)
	GetDeliverySequence(ctx context.Context, courierID int64) (*dto.GetDeliverySequenceResponse, error)
}

type routingService struct {
	courierRepo        courierrepo.CourierRepository
	locationRepo       locationrepo.LocationRepository
	deliveryRepo       deliveryrepo.DeliveryRepository
	orderRepo          orderrepo.OrderRepository
	addressRepo        addressrepo.AddressRepository
	reorderHistoryRepo reorderhistoryrepo.ReorderHistoryRepository
}

func NewService(
	courierRepo courierrepo.CourierRepository,
	locationRepo locationrepo.LocationRepository,
	deliveryRepo deliveryrepo.DeliveryRepository,
	orderRepo orderrepo.OrderRepository,
	addressRepo addressrepo.AddressRepository,
	reorderHistoryRepo reorderhistoryrepo.ReorderHistoryRepository,
) RoutingService {
	return &routingService{
		courierRepo:        courierRepo,
		locationRepo:       locationRepo,
		deliveryRepo:       deliveryRepo,
		orderRepo:          orderRepo,
		addressRepo:        addressRepo,
		reorderHistoryRepo: reorderHistoryRepo,
	}
}
```

---

## Algoritmo MVP

Antes do codigo, vale fixar a heuristica.

Para cada entrega pendente do entregador:

1. carregar a entrega
2. carregar o pedido vinculado
3. carregar o endereco do pedido
4. montar um score simples baseado em `zip_code + numero`
5. ordenar pelo score crescente
6. persistir `current_sequence`

### Exemplo de score

Uma forma didatica simples:

```text
score = zip_code_normalizado * 100000 + numero_endereco
```

Se o numero nao puder ser convertido com seguranca:

- usar `99999` como fallback

Isso nao e uma rota otima do mundo real.

Mas e uma heuristica operacional valida para MVP.

---

## `internal/service/routing/reorder_deliveries.go`

Digite em `internal/service/routing/reorder_deliveries.go`:

```go
package routing

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

type reorderCandidate struct {
	DeliveryID int64
	OrderID    int64
	ZipCode    string
	Number     string
	Score      float64
}

func (s *routingService) ReorderDeliveries(ctx context.Context, courierID int64) (*dto.ReorderDeliveriesResponse, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before reorder", err)
	}
	if courierModel == nil {
		return nil, apperror.NotFound("courier not found")
	}

	latestLocation, err := s.locationRepo.GetLatestCourierLocation(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get latest courier location before reorder", err)
	}
	if latestLocation == nil {
		return nil, apperror.BadRequest("courier must have a latest location before reorder")
	}

	deliveries, err := s.deliveryRepo.GetCourierPendingDeliveries(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier pending deliveries", err)
	}

	candidates := make([]reorderCandidate, 0, len(deliveries))
	for _, delivery := range deliveries {
		orderModel, err := s.orderRepo.GetOrderByID(ctx, delivery.OrderID)
		if err != nil {
			return nil, apperror.Internal("failed to get order during reorder", err)
		}
		if orderModel == nil {
			continue
		}

		addressModel, err := s.addressRepo.GetAddressByID(ctx, orderModel.DeliveryAddressID)
		if err != nil {
			return nil, apperror.Internal("failed to get address during reorder", err)
		}
		if addressModel == nil {
			continue
		}

		score := buildDeliveryScore(addressModel.ZipCode, addressModel.Number)
		candidates = append(candidates, reorderCandidate{
			DeliveryID: delivery.ID,
			OrderID:    delivery.OrderID,
			ZipCode:    addressModel.ZipCode,
			Number:     addressModel.Number,
			Score:      score,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score < candidates[j].Score
	})

	now := time.Now()
	responseItems := make([]dto.DeliverySequenceItemResponse, 0, len(candidates))
	for index, candidate := range candidates {
		sequence := index + 1

		if err := s.deliveryRepo.UpdateDeliverySequence(ctx, candidate.DeliveryID, sequence); err != nil {
			return nil, apperror.Internal("failed to update delivery sequence", err)
		}

		if err := s.deliveryRepo.UpdateDeliveryLastReorderedAt(
			ctx,
			candidate.DeliveryID,
			sql.NullTime{Time: now, Valid: true},
		); err != nil {
			return nil, apperror.Internal("failed to update last_reordered_at", err)
		}

		if err := s.reorderHistoryRepo.StoreReorderHistory(
			ctx,
			courierID,
			candidate.DeliveryID,
			sequence,
			candidate.Score,
			"zip_code_plus_number",
			now,
		); err != nil {
			return nil, apperror.Internal("failed to store reorder history", err)
		}

		responseItems = append(responseItems, dto.DeliverySequenceItemResponse{
			DeliveryID: candidate.DeliveryID,
			OrderID:    candidate.OrderID,
			Sequence:   sequence,
			ZipCode:    candidate.ZipCode,
			Number:     candidate.Number,
			Score:      candidate.Score,
		})
	}

	return &dto.ReorderDeliveriesResponse{
		CourierID:   courierID,
		GeneratedAt: now,
		Deliveries:  responseItems,
	}, nil
}

func buildDeliveryScore(zipCode string, number string) float64 {
	normalizedZipCode := strings.ReplaceAll(zipCode, "-", "")
	zipValue, err := strconv.Atoi(normalizedZipCode)
	if err != nil {
		zipValue = 99999999
	}

	numberValue, err := strconv.Atoi(number)
	if err != nil {
		numberValue = 99999
	}

	return float64(zipValue*100000 + numberValue)
}
```

### O Que Esse Metodo Ensina

- coordenacao entre varios repositories
- persistencia do resultado do algoritmo
- separacao entre heuristica e camada HTTP
- historico operacional para auditoria

---

## `internal/service/routing/get_delivery_sequence.go`

Digite em `internal/service/routing/get_delivery_sequence.go`:

```go
package routing

import (
	"context"
	"sort"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *routingService) GetDeliverySequence(ctx context.Context, courierID int64) (*dto.GetDeliverySequenceResponse, error) {
	courierModel, err := s.courierRepo.GetCourierByID(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier before loading sequence", err)
	}
	if courierModel == nil {
		return nil, apperror.NotFound("courier not found")
	}

	deliveries, err := s.deliveryRepo.GetCourierPendingDeliveries(ctx, courierID)
	if err != nil {
		return nil, apperror.Internal("failed to get courier deliveries for sequence", err)
	}

	sort.Slice(deliveries, func(i, j int) bool {
		left := 999999
		right := 999999
		if deliveries[i].CurrentSequence != nil {
			left = *deliveries[i].CurrentSequence
		}
		if deliveries[j].CurrentSequence != nil {
			right = *deliveries[j].CurrentSequence
		}

		return left < right
	})

	items := make([]dto.DeliverySequenceItemResponse, 0, len(deliveries))
	for _, delivery := range deliveries {
		sequence := 0
		if delivery.CurrentSequence != nil {
			sequence = *delivery.CurrentSequence
		}

		items = append(items, dto.DeliverySequenceItemResponse{
			DeliveryID: delivery.ID,
			OrderID:    delivery.OrderID,
			Sequence:   sequence,
		})
	}

	return &dto.GetDeliverySequenceResponse{
		CourierID:  courierID,
		Deliveries: items,
	}, nil
}
```

---

## `internal/handler/routing/handler.go`

Digite em `internal/handler/routing/handler.go`:

```go
package routing

import (
	routingservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/routing"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router         gin.IRouter
	validate       *validator.Validate
	routingService routingservice.RoutingService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, routingService routingservice.RoutingService) *Handler {
	return &Handler{
		router:         router,
		validate:       validate,
		routingService: routingService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/couriers/:courier_id/deliveries/reorder", h.ReorderDeliveries)
	h.router.GET("/couriers/:courier_id/deliveries/sequence", h.GetDeliverySequence)
}
```

---

## `internal/handler/routing/reorder_deliveries.go`

Digite em `internal/handler/routing/reorder_deliveries.go`:

```go
package routing

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ReorderDeliveries(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	response, err := h.routingService.ReorderDeliveries(c.Request.Context(), courierID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/routing/get_delivery_sequence.go`

Digite em `internal/handler/routing/get_delivery_sequence.go`:

```go
package routing

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDeliverySequence(c *gin.Context) {
	courierID, ok := httpresponse.ParseInt64Param(c, "courier_id")
	if !ok {
		return
	}

	response, err := h.routingService.GetDeliverySequence(c.Request.Context(), courierID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/server/router.go`

Agora o `router` central passa a ligar tambem `routing`.

Adicione nos imports de `internal/server/router.go`:

```go
import (
	addressrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/address"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	courierrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/courier"
	deliveryrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/delivery"
	locationrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/location"
	orderrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/order"
	reorderhistoryrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/reorderhistory"
	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	courierhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/courier"
	deliveryhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/delivery"
	locationhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/location"
	orderhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/order"
	routinghandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/routing"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	courierservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/courier"
	deliveryservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/delivery"
	locationservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/location"
	orderservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/order"
	routingservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/routing"
)
```

E, dentro de `NewRouter(...)`:

```go
reorderHistoryRepo := reorderhistoryrepository.NewRepository(db)
locationRepo := locationrepository.NewRepository(db)
deliveryRepo := deliveryrepository.NewRepository(db)

routingService := routingservice.NewService(
	courierRepo,
	locationRepo,
	deliveryRepo,
	orderRepo,
	addressRepo,
	reorderHistoryRepo,
)

routingHandler := routinghandler.NewHandler(router, validate, routingService)
routingHandler.RouteList()
```

### Observacao

Se voce ja criou `locationRepo` e `deliveryRepo` antes no `main`, apenas reaproveite as mesmas instancias.

---

## Resumo Do Fluxo De Reorder

### Reclassificacao

1. handler le `courier_id`
2. service verifica se o entregador existe
3. service valida se existe localizacao atual
4. service carrega entregas pendentes do entregador
5. service carrega pedido e endereco de cada entrega
6. service calcula score `CEP + numero`
7. service ordena as entregas
8. service persiste `current_sequence`
9. service persiste `last_reordered_at`
10. service registra historico
11. handler responde `200`

### Consulta De Sequencia

1. handler le `courier_id`
2. service valida se o entregador existe
3. service carrega entregas pendentes
4. service ordena por `current_sequence`
5. handler responde `200`

---

## Pontos De Log Recomendados Em Delivery Reorder

Quando voce for incrementar logging nesse modulo, a distribuicao mais valiosa costuma ser:

- `handler`: logar erro de bind, validacao e falha inesperada na borda HTTP
- `service`: logar entregador inexistente, ausencia de localizacao, ausencia de entregas pendentes e resultado da reclassificacao
- `repository`: logar falha inesperada ao consultar entregas, atualizar sequencia ou gravar historico

Por operacao:

- reclassificacao: logar quantidade de entregas reordenadas, inicio e fim do processo e falhas nas dependencias usadas no calculo
- consulta de sequencia: logar ausencia de entregas apenas quando isso ajudar no suporte

Campos uteis nesse contexto:

- `courier_id`
- quantidade de entregas consideradas
- quantidade de entregas reordenadas
- horario da ultima reordenacao

Evite:

- logar endereco completo do cliente sem necessidade
- logar todos os detalhes do algoritmo em nivel alto
- poluir o log com sequencias completas em toda chamada

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos:

- `internal/dto/routing_dto.go`
- `internal/repository/reorderhistory/repository.go`
- `internal/repository/reorderhistory/store_reorder_history.go`
- `internal/repository/delivery/get_courier_pending_deliveries.go`
- `internal/repository/delivery/update_delivery_sequence.go`
- `internal/repository/delivery/update_delivery_last_reordered_at.go`
- `internal/service/routing/service.go`
- `internal/service/routing/reorder_deliveries.go`
- `internal/service/routing/get_delivery_sequence.go`
- `internal/handler/routing/handler.go`
- `internal/handler/routing/reorder_deliveries.go`
- `internal/handler/routing/get_delivery_sequence.go`

---

## Observacao Final

Este modulo e onde o projeto comeca a ir alem de CRUD de mercado comum.

Aqui voce demonstra:

- coordenacao entre varios modulos
- logica operacional real
- heuristica de ordenacao
- rastreabilidade do resultado

Mesmo sendo uma heuristica simples, ela ja comunica muito bem a proposta do sistema.

---

## Proximo Passo

Depois deste documento, os proximos passos mais coerentes sao:

- criar o guia de testes da fase 1
- evoluir a heuristica para usar localizacao e distancia com mais peso
- documentar observabilidade do fluxo operacional

Se quiser, no proximo passo eu posso gerar:

- um documento unico de testes para `buyers`, `couriers`, `orders`, `deliveries`, `locations` e `reorder`
- ou um documento de observabilidade e logs do projeto
