# Fluxo de Buyers: Handler, Service e Repository

## Visao Geral

Este documento foi reescrito para seguir o mesmo estilo do projeto `go-tweets`.

Ou seja:

- foco em fluxo por camada
- explicacao arquivo por arquivo
- separacao clara entre `handler`, `service` e `repository`
- codigo dividido por responsabilidade
- contrato HTTP padronizado

O dominio `buyer` cuida de:

- criacao de comprador
- busca por ID
- listagem
- atualizacao
- exclusao

Fluxo:

```text
HTTP /buyers/* -> handler/buyer -> service/buyer -> repository/buyer -> MySQL
```

## Atualizacao Do Bootstrap

Desde a ultima revisao deste guia, o projeto passou a ter bootstrap real com:

- `Gin`
- `validator`
- `internal/httpresponse`
- `internal/apperror`
- `internal/server/router.go`
- base de testes em `test/unit` e `test/integration`

Por isso, sempre que algum trecho mais antigo deste documento mencionar:

- `net/http`
- `ServeMux`
- `cmd/main.go` como ponto principal de registro das rotas

leia no estado atual do projeto desta forma:

- handlers devem receber `*gin.Context`
- construtores de handler devem usar `gin.IRouter`
- bind e validacao devem passar por `internal/httpresponse`
- o bootstrap das rotas deve ser centralizado em `internal/server/router.go`

---

## Objetivo Deste Capitulo

Ao final deste documento, voce deve conseguir montar um CRUD de `buyers` alinhado ao padrao arquitetural do `go-tweets`, adaptado ao contexto atual do `go-delivery-routing-lab`.

Este guia mostra:

- quais arquivos criar
- em qual arquivo digitar cada trecho
- como dividir responsabilidade entre camadas
- como fica o fluxo de request e response

---

## Estrutura Recomendada

Se voce quiser seguir o mesmo padrao do `go-tweets`, a estrutura de `buyer` fica melhor assim:

```text
internal/
  apperror/
    error.go
  dto/
    buyer_dto.go
  handler/
    buyer/
      create_buyer.go
      delete_buyer.go
      get_all_buyers.go
      get_buyer_by_id.go
      handler.go
      update_buyer.go
  httpresponse/
    response.go
  model/
    buyer_model.go
  repository/
    buyer/
      create_buyer.go
      delete_buyer.go
      get_all_buyers.go
      get_buyer_by_id.go
      repository.go
      update_buyer.go
  service/
    buyer/
      create_buyer.go
      delete_buyer.go
      get_all_buyers.go
      get_buyer_by_id.go
      service.go
      update_buyer.go
cmd/
  main.go
```

### Diferenca Em Relacao Ao Documento Anterior

O guia antigo seguia um estilo mais "roteiro de digitacao".

Agora o padrao fica mais proximo do `go-tweets`:

- `dto` centralizado em um arquivo por dominio
- `model` em vez de `entity`
- `service.go` com interface e implementacao concreta
- `repository.go` com interface e struct concreta
- um arquivo por operacao dentro de `handler`, `service` e `repository`

---

## Contrato HTTP Inicial

Para o CRUD de `buyers`, o contrato sugerido e:

- `POST /buyers`
- `GET /buyers/:buyer_id`
- `GET /buyers`
- `PUT /buyers/:buyer_id`
- `DELETE /buyers/:buyer_id`

Payload de criacao:

```json
{
  "name": "Maria Silva",
  "document": "12345678900",
  "phone": "11999999999",
  "email": "maria@email.com"
}
```

Payload de atualizacao:

```json
{
  "name": "Maria Oliveira",
  "document": "12345678900",
  "phone": "11988888888",
  "email": "maria.oliveira@email.com"
}
```

---

## `internal/dto/buyer_dto.go`

No `go-tweets`, os DTOs de um dominio ficam centralizados em um unico arquivo.

Aqui voce pode seguir a mesma ideia.

Digite em `internal/dto/buyer_dto.go`:

```go
package dto

import "time"

type (
	CreateBuyerRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Document string `json:"document" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		Email    string `json:"email" validate:"omitempty,email"`
	}

	UpdateBuyerRequest struct {
		Name     string `json:"name" validate:"required,min=3"`
		Document string `json:"document" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		Email    string `json:"email" validate:"omitempty,email"`
	}

	BuyerResponse struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		Document  string    `json:"document"`
		Phone     string    `json:"phone"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
```

### Por Que Esse Formato

- deixa o dominio `buyer` mais facil de localizar
- segue a mesma ideia de `user_dto.go` no `go-tweets`
- as tags `validate` movem a validacao de entrada para o DTO e para o `handler`

### Regra Pratica

Se voce quer que o `service` fique somente com regra de negocio, entao estas validacoes devem sair do `service`:

- `required`
- `min`
- `email`
- body JSON invalido
- parametro de rota invalido

Esses pontos ficam melhores em:

- `dto`
- `httpresponse`
- `handler`

### Observacao Sobre Espacos Em Branco

As tags `validate:"required"` e `validate:"min=3"` nao tratam sozinhas o caso de `"   "`.

Se depois voce quiser bloquear strings apenas com espacos sem levar isso para o `service`, as opcoes mais comuns sao:

- criar validacao customizada no `validator`
- normalizar os campos antes da validacao

Neste documento, vou manter o `service` limpo e focado em regra de negocio.

### Normalizacao Recomendada Neste CRUD

Para `buyers`, a normalizacao mais importante deve acontecer logo no inicio do fluxo de `Create` e `Update`, antes de:

- verificar duplicidade
- comparar valores
- montar o `BuyerModel`

Aplicacao sugerida:

- `name`: aplicar `strings.TrimSpace(...)`
- `document`: remover mascara e manter apenas digitos
- `phone`: remover mascara e manter apenas digitos
- `email`: aplicar `strings.TrimSpace(...)` e `strings.ToLower(...)`

Boa regra pratica:

- o `handler` continua responsavel por bind e validate
- o `service` pode receber o request e normalizar os campos antes da regra de unicidade
- o banco continua sendo a garantia final com `UNIQUE`

---

## `internal/model/buyer_model.go`

No `go-tweets`, a camada de persistencia conversa com `model`, nao com DTO.

Digite em `internal/model/buyer_model.go`:

```go
package model

import "time"

type BuyerModel struct {
	ID        int64
	Name      string
	Document  string
	Phone     string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

### Papel Do Model

- representar o dado persistido
- trafegar entre `service` e `repository`
- manter a camada HTTP desacoplada da camada SQL

---

## `internal/apperror/error.go`

Para ficar no mesmo padrao do `go-tweets`, o ideal e usar um `apperror` parecido com este.

Digite em `internal/apperror/error.go`:

```go
package apperror

import "net/http"

type Error struct {
	statusCode int
	message    string
	cause      error
}

func New(statusCode int, message string) *Error {
	return &Error{
		statusCode: statusCode,
		message:    message,
	}
}

func Wrap(statusCode int, message string, cause error) *Error {
	return &Error{
		statusCode: statusCode,
		message:    message,
		cause:      cause,
	}
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, message)
}

func Internal(message string, cause error) *Error {
	return Wrap(http.StatusInternalServerError, message, cause)
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) Unwrap() error {
	return e.cause
}

func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if statusCoder, ok := err.(interface{ StatusCode() int }); ok {
		return statusCoder.StatusCode()
	}

	return http.StatusInternalServerError
}
```

### Ideia Central

No `go-tweets`, o service nao devolve `statusCode` cru.

Em vez disso:

- o service devolve `error`
- o erro ja carrega semantica HTTP
- o handler apenas traduz a resposta final

---

## `internal/httpresponse/response.go`

Para manter o mesmo padrao de centralizacao de resposta, use um helper unico baseado em `Gin`.

Digite em `internal/httpresponse/response.go`:

```go
package httpresponse

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

func JSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func JSONAppError(c *gin.Context, err error) {
	JSONError(c, apperror.StatusCode(err), err.Error())
}

func BindAndValidateJSON(c *gin.Context, validate *validator.Validate, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, "invalid request body")
		return false
	}

	if err := validate.Struct(req); err != nil {
		JSONValidationError(c, err, req)
		return false
	}

	return true
}

func JSONValidationError(c *gin.Context, err error, req interface{}) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		JSONError(c, http.StatusBadRequest, "validation failed")
		return
	}

	fieldNames := jsonFieldNames(req)
	details := make([]ErrorDetail, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		fieldName := fieldNames[validationErr.Field()]
		if fieldName == "" {
			fieldName = strings.ToLower(validationErr.Field())
		}

		details = append(details, ErrorDetail{
			Field:   fieldName,
			Message: validationMessage(validationErr, fieldNames),
		})
	}

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Message: "validation failed",
		Errors:  details,
	})
}

func ParseInt64Param(c *gin.Context, fieldName string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(fieldName), 10, 64)
	if err != nil {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", fieldName))
		return 0, false
	}

	return value, true
}

func jsonFieldNames(req interface{}) map[string]string {
	reqType := reflect.TypeOf(req)
	if reqType == nil {
		return map[string]string{}
	}

	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	if reqType.Kind() != reflect.Struct {
		return map[string]string{}
	}

	fieldNames := make(map[string]string, reqType.NumField())
	for index := 0; index < reqType.NumField(); index++ {
		field := reqType.Field(index)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldNames[field.Name] = strings.Split(jsonTag, ",")[0]
	}

	return fieldNames
}

func validationMessage(fieldError validator.FieldError, fieldNames map[string]string) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must have at least %s characters", fieldError.Param())
	default:
		return "is invalid"
	}
}
```

### O Que Esse Arquivo Resolve

- padroniza resposta JSON
- evita repeticao de `json.NewEncoder(...)`
- centraliza parse de parametro numerico
- centraliza bind e validacao de request
- aproxima o guia do estilo de `internal/httpresponse/response.go` do `go-tweets`

### Estado Atual

Esse helper ja existe na base do projeto e hoje e uma das pecas centrais da camada HTTP.

---

## `internal/repository/buyer/repository.go`

No `go-tweets`, `repository.go` guarda interface e implementacao concreta.

Digite em `internal/repository/buyer/repository.go`:

```go
package buyer

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

type BuyerRepository interface {
	CreateBuyer(ctx context.Context, model *model.BuyerModel) (int64, error)
	GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error)
	GetAllBuyers(ctx context.Context) ([]model.BuyerModel, error)
	UpdateBuyer(ctx context.Context, model *model.BuyerModel) error
	DeleteBuyer(ctx context.Context, buyerID int64) error
}

type buyerRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) BuyerRepository {
	return &buyerRepository{
		db: db,
	}
}
```

### Por Que Esse Padrao E Bom

- o handler nao conhece SQL
- o service depende de interface
- o construtor `NewRepository` repete o mesmo desenho usado no `go-tweets`
- o repository expoe apenas as operacoes realmente necessarias para o CRUD

---

## `internal/repository/buyer/create_buyer.go`

Digite em `internal/repository/buyer/create_buyer.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) CreateBuyer(ctx context.Context, model *model.BuyerModel) (int64, error) {
	query := `INSERT INTO buyers (name, document, phone, email, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.Document,
		model.Phone,
		model.Email,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	buyerID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return buyerID, nil
}
```

### Observacao

No estilo do `go-tweets`, o repository:

- executa SQL
- nao valida regra de negocio
- devolve erro tecnico ou resultado persistido

---

## `internal/repository/buyer/get_buyer_by_id.go`

Digite em `internal/repository/buyer/get_buyer_by_id.go`:

```go
package buyer

import (
	"context"
	"database/sql"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error) {
	query := `SELECT id, name, document, phone, email, created_at, updated_at
	FROM buyers WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, buyerID)

	var result model.BuyerModel
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Document,
		&result.Phone,
		&result.Email,
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

### Ponto Importante

Repare no mesmo padrao do `go-tweets`:

- `nil, nil` significa "nao encontrado, mas sem erro tecnico"
- quem decide se isso vira `404` e o `service`

---

### Observacao Sobre Unicidade

Como o schema agora passa a garantir `UNIQUE` para `document`, `email` e `phone`, este CRUD deixa de depender de consultas previas de duplicidade no `repository`.

O fluxo recomendado fica assim:

- o `service` normaliza os valores
- o `repository` tenta persistir
- o banco barra duplicidade
- o `service` traduz o erro tecnico em `409 Conflict`

---

## `internal/repository/buyer/get_all_buyers.go`

Digite em `internal/repository/buyer/get_all_buyers.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) GetAllBuyers(ctx context.Context) ([]model.BuyerModel, error) {
	query := `SELECT id, name, document, phone, email, created_at, updated_at
	FROM buyers
	ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.BuyerModel, 0)
	for rows.Next() {
		var buyer model.BuyerModel
		if err := rows.Scan(
			&buyer.ID,
			&buyer.Name,
			&buyer.Document,
			&buyer.Phone,
			&buyer.Email,
			&buyer.CreatedAt,
			&buyer.UpdatedAt,
		); err != nil {
			return nil, err
		}

		results = append(results, buyer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
```

---

## `internal/repository/buyer/update_buyer.go`

Digite em `internal/repository/buyer/update_buyer.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (r *buyerRepository) UpdateBuyer(ctx context.Context, model *model.BuyerModel) error {
	query := `UPDATE buyers
	SET name = ?, document = ?, phone = ?, email = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(
		ctx,
		query,
		model.Name,
		model.Document,
		model.Phone,
		model.Email,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return err
	}

	return nil
}
```

### Melhoria Sugerida

Se quiser deixar ainda mais robusto, voce pode usar `RowsAffected()` e retornar erro quando nada for alterado.

---

## `internal/repository/buyer/delete_buyer.go`

Digite em `internal/repository/buyer/delete_buyer.go`:

```go
package buyer

import "context"

func (r *buyerRepository) DeleteBuyer(ctx context.Context, buyerID int64) error {
	query := `DELETE FROM buyers WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, buyerID)
	if err != nil {
		return err
	}

	return nil
}
```

---

## `internal/service/buyer/service.go`

No `go-tweets`, `service.go` define interface, struct concreta e construtor.

Digite em `internal/service/buyer/service.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	buyerrepo "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
)

type BuyerService interface {
	Create(ctx context.Context, req *dto.CreateBuyerRequest) (int64, error)
	GetByID(ctx context.Context, buyerID int64) (*dto.BuyerResponse, error)
	GetAll(ctx context.Context) ([]dto.BuyerResponse, error)
	Update(ctx context.Context, buyerID int64, req *dto.UpdateBuyerRequest) (*dto.BuyerResponse, error)
	Delete(ctx context.Context, buyerID int64) error
}

type buyerService struct {
	buyerRepo buyerrepo.BuyerRepository
}

func NewService(buyerRepo buyerrepo.BuyerRepository) BuyerService {
	return &buyerService{
		buyerRepo: buyerRepo,
	}
}
```

### Beneficio

- o handler depende de interface
- a implementacao concreta fica escondida
- o padrao combina com `internal/service/user/service.go` do `go-tweets`

---

## `internal/service/buyer/create_buyer.go`

Digite em `internal/service/buyer/create_buyer.go`:

```go
package buyer

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *buyerService) Create(ctx context.Context, req *dto.CreateBuyerRequest) (int64, error) {
	normalizedName := strings.TrimSpace(req.Name)
	normalizedDocument := onlyDigits(req.Document)
	normalizedPhone := onlyDigits(req.Phone)
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	now := time.Now()
	buyerModel := &model.BuyerModel{
		Name:      normalizedName,
		Document:  normalizedDocument,
		Phone:     normalizedPhone,
		Email:     normalizedEmail,
		CreatedAt: now,
		UpdatedAt: now,
	}

	buyerID, err := s.buyerRepo.CreateBuyer(ctx, buyerModel)
	if err != nil {
		if conflictErr := mapBuyerConflictError(err); conflictErr != nil {
			return 0, conflictErr
		}
		return 0, apperror.Internal("failed to create buyer", err)
	}

	return buyerID, nil
}
```

### Leitura Do Fluxo

1. handler faz bind e validate do request
2. service normaliza `name`, `document`, `phone` e `email`
3. service monta o `BuyerModel`
4. repository tenta persistir
5. banco garante unicidade com `UNIQUE`
6. service traduz erro de duplicidade em `409 Conflict`
7. handler monta a resposta HTTP

### Regras De Negocio Atuais

- normalizar `name` com `strings.TrimSpace(...)`
- normalizar `document` mantendo apenas digitos
- normalizar `phone` mantendo apenas digitos
- normalizar `email` com `strings.TrimSpace(...)` e `strings.ToLower(...)`
- traduzir violacao de `UNIQUE` do banco em erro HTTP de conflito

Os helpers `onlyDigits(...)` e `mapBuyerConflictError(...)` podem ficar em um arquivo auxiliar da pasta `internal/service/buyer/`.

### Importante

Essas verificacoes no `service` sao importantes, mas o ideal em producao e combinar isso com restricoes `UNIQUE` no banco.

Assim voce protege:

- o fluxo da aplicacao
- e a integridade dos dados no schema

---

## `internal/service/buyer/get_buyer_by_id.go`

Digite em `internal/service/buyer/get_buyer_by_id.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *buyerService) GetByID(ctx context.Context, buyerID int64) (*dto.BuyerResponse, error) {
	buyerModel, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return nil, apperror.Internal("failed to get buyer", err)
	}
	if buyerModel == nil {
		return nil, apperror.NotFound("buyer not found")
	}

	return &dto.BuyerResponse{
		ID:        buyerModel.ID,
		Name:      buyerModel.Name,
		Document:  buyerModel.Document,
		Phone:     buyerModel.Phone,
		Email:     buyerModel.Email,
		CreatedAt: buyerModel.CreatedAt,
		UpdatedAt: buyerModel.UpdatedAt,
	}, nil
}
```

---

## `internal/service/buyer/get_all_buyers.go`

Digite em `internal/service/buyer/get_all_buyers.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
)

func (s *buyerService) GetAll(ctx context.Context) ([]dto.BuyerResponse, error) {
	buyers, err := s.buyerRepo.GetAllBuyers(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to get buyers", err)
	}

	response := make([]dto.BuyerResponse, 0, len(buyers))
	for _, buyer := range buyers {
		response = append(response, dto.BuyerResponse{
			ID:        buyer.ID,
			Name:      buyer.Name,
			Document:  buyer.Document,
			Phone:     buyer.Phone,
			Email:     buyer.Email,
			CreatedAt: buyer.CreatedAt,
			UpdatedAt: buyer.UpdatedAt,
		})
	}

	return response, nil
}
```

---

## `internal/service/buyer/update_buyer.go`

Digite em `internal/service/buyer/update_buyer.go`:

```go
package buyer

import (
	"context"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/model"
)

func (s *buyerService) Update(ctx context.Context, buyerID int64, req *dto.UpdateBuyerRequest) (*dto.BuyerResponse, error) {
	existingBuyer, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return nil, apperror.Internal("failed to get buyer before update", err)
	}
	if existingBuyer == nil {
		return nil, apperror.NotFound("buyer not found")
	}

	existingBuyer.Name = req.Name
	existingBuyer.Document = req.Document
	existingBuyer.Phone = req.Phone
	existingBuyer.Email = req.Email
	existingBuyer.UpdatedAt = time.Now()

	if err := s.buyerRepo.UpdateBuyer(ctx, &model.BuyerModel{
		ID:        existingBuyer.ID,
		Name:      existingBuyer.Name,
		Document:  existingBuyer.Document,
		Phone:     existingBuyer.Phone,
		Email:     existingBuyer.Email,
		CreatedAt: existingBuyer.CreatedAt,
		UpdatedAt: existingBuyer.UpdatedAt,
	}); err != nil {
		return nil, apperror.Internal("failed to update buyer", err)
	}

	return &dto.BuyerResponse{
		ID:        existingBuyer.ID,
		Name:      existingBuyer.Name,
		Document:  existingBuyer.Document,
		Phone:     existingBuyer.Phone,
		Email:     existingBuyer.Email,
		CreatedAt: existingBuyer.CreatedAt,
		UpdatedAt: existingBuyer.UpdatedAt,
	}, nil
}
```

### Regra De Negocio Que Permanece No `service`

Mesmo removendo a validacao de formato e obrigatoriedade, ainda faz sentido manter no `service`:

- verificar se o comprador existe antes de atualizar
- normalizar os campos antes de persistir
- impedir update de recurso inexistente
- traduzir conflito de `document`, `email` e `phone` quando o banco rejeitar o `UPDATE`
- aplicar futuras restricoes de negocio do dominio

---

## `internal/service/buyer/delete_buyer.go`

Digite em `internal/service/buyer/delete_buyer.go`:

```go
package buyer

import (
	"context"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/apperror"
)

func (s *buyerService) Delete(ctx context.Context, buyerID int64) error {
	existingBuyer, err := s.buyerRepo.GetBuyerByID(ctx, buyerID)
	if err != nil {
		return apperror.Internal("failed to get buyer before delete", err)
	}
	if existingBuyer == nil {
		return apperror.NotFound("buyer not found")
	}

	if err := s.buyerRepo.DeleteBuyer(ctx, buyerID); err != nil {
		return apperror.Internal("failed to delete buyer", err)
	}

	return nil
}
```

---

## `internal/handler/buyer/handler.go`

No `go-tweets`, o `handler.go` guarda dependencias e construtor.

No estado atual do seu projeto, o handler de `buyer` deve seguir este formato:

Digite em `internal/handler/buyer/handler.go`:

```go
package buyer

import (
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	router       gin.IRouter
	validate     *validator.Validate
	buyerService buyerservice.BuyerService
}

func NewHandler(router gin.IRouter, validate *validator.Validate, buyerService buyerservice.BuyerService) *Handler {
	return &Handler{
		router:       router,
		validate:     validate,
		buyerService: buyerService,
	}
}

func (h *Handler) RouteList() {
	h.router.POST("/buyers", h.CreateBuyer)
	h.router.GET("/buyers", h.GetAllBuyers)
	h.router.GET("/buyers/:buyer_id", h.GetBuyerByID)
	h.router.PUT("/buyers/:buyer_id", h.UpdateBuyer)
	h.router.DELETE("/buyers/:buyer_id", h.DeleteBuyer)
}
```

### Analogia Com O `go-tweets`

No `go-tweets` o handler guarda:

- roteador
- validador
- service

Aqui, no estado atual do projeto, a adaptacao equivalente e:

- `gin.IRouter`
- `validator.Validate`
- `buyerService`

---

## `internal/handler/buyer/create_buyer.go`

Digite em `internal/handler/buyer/create_buyer.go`:

```go
package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateBuyer(c *gin.Context) {
	var req dto.CreateBuyerRequest

	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	buyerID, err := h.buyerService.Create(c.Request.Context(), &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusCreated, map[string]int64{
		"id": buyerID,
	})
}
```

### Leitura Didatica

- o handler faz bind e validate do JSON
- o handler nao executa SQL
- o handler nao decide regra de negocio
- o handler apenas chama `buyerService.Create(...)`

---

## `internal/handler/buyer/get_buyer_by_id.go`

Digite em `internal/handler/buyer/get_buyer_by_id.go`:

```go
package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetBuyerByID(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}

	response, err := h.buyerService.GetByID(c.Request.Context(), buyerID)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/buyer/get_all_buyers.go`

Digite em `internal/handler/buyer/get_all_buyers.go`:

```go
package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAllBuyers(c *gin.Context) {
	response, err := h.buyerService.GetAll(c.Request.Context())
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/buyer/update_buyer.go`

Digite em `internal/handler/buyer/update_buyer.go`:

```go
package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/dto"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateBuyer(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}

	var req dto.UpdateBuyerRequest
	if !httpresponse.BindAndValidateJSON(c, h.validate, &req) {
		return
	}

	response, err := h.buyerService.Update(c.Request.Context(), buyerID, &req)
	if err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	httpresponse.JSON(c, http.StatusOK, response)
}
```

---

## `internal/handler/buyer/delete_buyer.go`

Digite em `internal/handler/buyer/delete_buyer.go`:

```go
package buyer

import (
	"net/http"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteBuyer(c *gin.Context) {
	buyerID, ok := httpresponse.ParseInt64Param(c, "buyer_id")
	if !ok {
		return
	}

	if err := h.buyerService.Delete(c.Request.Context(), buyerID); err != nil {
		httpresponse.JSONAppError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
```

---

## `internal/server/router.go`

No estado atual do projeto, o registro de rotas deve convergir para `internal/server/router.go`.

Para ligar `buyer`, voce vai estender esse ponto central.

Digite em `internal/server/router.go`:

```go
package server

import (
	"database/sql"
	"time"

	buyerhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/buyer"
	healthhandler "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/handler/health"
	buyerrepository "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/repository/buyer"
	buyerservice "github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/service/buyer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const healthCheckTimeout = 2 * time.Second

type Dependencies struct {
	DB            *sql.DB
	HealthChecker healthhandler.Checker
}

func NewRouter(validate *validator.Validate, deps Dependencies) *gin.Engine {
	if validate == nil {
		validate = validator.New()
	}

	router := gin.New()
	router.Use(gin.Recovery())

	healthhandler.NewHandler(router, deps.HealthChecker, healthCheckTimeout).RouteList()

	if deps.DB == nil {
		return router
	}

	// Buyer
	// O mesmo padrao sera repetido depois para courier, order, delivery, location e routing.
	db := deps.DB
	buyerRepo := buyerrepository.NewRepository(db)
	buyerService := buyerservice.NewService(buyerRepo)
	buyerHandler := buyerhandler.NewHandler(router, validate, buyerService)
	buyerHandler.RouteList()

	return router
}
```

### Fluxo Final

1. `cmd/main.go` instancia `validator`
2. `cmd/main.go` conecta o banco
3. `cmd/main.go` chama `server.NewRouter(...)`
4. `internal/server/router.go` instancia `repository`
5. `internal/server/router.go` injeta no `service`
6. `internal/server/router.go` injeta `service` e `validator` no `handler`
7. `handler.RouteList()` registra rotas
8. requests entram em `/buyers`

---

## Resumo Do Fluxo De Buyers

### Criacao

1. handler le e valida o JSON
2. service executa regra de negocio
3. service monta `BuyerModel`
4. repository executa `INSERT`
5. handler responde `201`

### Busca Por ID

1. handler le `buyer_id`
2. service chama repository
3. repository faz `SELECT`
4. service transforma ausencia em `404`
5. handler responde `200`

### Listagem

1. handler chama `GetAll`
2. repository faz `SELECT ... ORDER BY id ASC`
3. service monta resposta
4. handler responde `200`

### Atualizacao

1. handler le `buyer_id`
2. handler le e valida o body
3. service executa regra de negocio
4. service busca comprador atual
5. repository executa `UPDATE`
6. handler responde `200`

### Exclusao

1. handler le `buyer_id`
2. service verifica existencia
3. repository executa `DELETE`
4. handler responde `204`

---

## Pontos De Log Recomendados No CRUD De Buyers

Quando voce for fazer a incrementacao de logs no codigo, a sugestao mais equilibrada para `buyer` e:

- `handler`: logar erro de bind, validacao ou falha inesperada na borda HTTP
- `service`: logar conflito de `document`, `email` ou `phone` e falhas relevantes de negocio
- `repository`: logar erro inesperado de banco, especialmente em `INSERT`, `UPDATE`, `DELETE` e consultas criticas

Para cada operacao, pense assim:

- criacao: vale logar conflito de unicidade e sucesso com `buyer_id`
- busca por ID: vale logar ausencia quando isso ajudar no diagnostico
- listagem: normalmente o middleware HTTP ja cobre bem
- atualizacao: vale logar tentativa de alterar recurso inexistente ou conflito
- exclusao: vale logar ausencia e falha de persistencia

Evite registrar em log, sem necessidade real:

- `document` completo
- `email` completo
- `phone` completo
- payload bruto inteiro da request

Se quiser manter o projeto limpo, a distribuicao recomendada e:

- middleware para log transversal da request
- handler para contexto HTTP
- service para regra de negocio
- repository para erro de infraestrutura

---

## Checklist De Implementacao

Confirme se voce criou estes arquivos no padrao do `go-tweets`:

- `internal/dto/buyer_dto.go`
- `internal/model/buyer_model.go`
- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `internal/repository/buyer/repository.go`
- `internal/repository/buyer/create_buyer.go`
- `internal/repository/buyer/get_buyer_by_id.go`
- `internal/repository/buyer/get_all_buyers.go`
- `internal/repository/buyer/update_buyer.go`
- `internal/repository/buyer/delete_buyer.go`
- `internal/service/buyer/service.go`
- `internal/service/buyer/create_buyer.go`
- `internal/service/buyer/get_buyer_by_id.go`
- `internal/service/buyer/get_all_buyers.go`
- `internal/service/buyer/update_buyer.go`
- `internal/service/buyer/delete_buyer.go`
- `internal/handler/buyer/handler.go`
- `internal/handler/buyer/create_buyer.go`
- `internal/handler/buyer/get_buyer_by_id.go`
- `internal/handler/buyer/get_all_buyers.go`
- `internal/handler/buyer/update_buyer.go`
- `internal/handler/buyer/delete_buyer.go`
- `internal/server/router.go`
- `cmd/main.go`
- `test/unit/buyer_service_test.go`

## Testes De Regra De Negocio

Para esse modulo, vale muito a pena testar o `service` cobrindo os fluxos de unicidade.

Arquivo sugerido:

- `test/unit/buyer_service_test.go`

Cenarios minimos:

- conflito quando `document` ja existe
- conflito quando `email` ja existe
- conflito quando `phone` ja existe
- sucesso quando `document`, `email` e `phone` sao unicos

Esses cenarios ja existem hoje na base do projeto e servem como protecao para a regra de negocio do `Create`.

---

## Observacao Final

Este documento continua seguindo o padrao conceitual do `go-tweets`, mas agora a base real do projeto tambem ja foi aproximada.

Hoje, o projeto ja possui:

- `Gin`
- `validator`
- `internal/httpresponse`
- `internal/apperror`
- estrutura inicial de testes no mesmo estilo do `go-tweets`

Ou seja, a principal adaptacao que permanece neste guia e arquitetural:

- uso de `sql.DB` direto no repository em vez de uma camada adicional como `QueryExecutor`

---

## Proximo Passo

Depois deste documento, o proximo mais coerente e criar:

- `docs/Criacao do Crud/02-cruds/02-crud-couriers.md`

No mesmo formato:

- visao geral
- fluxo por camada
- arquivos a criar
- codigo por arquivo
- estrutura no padrao do `go-tweets`
