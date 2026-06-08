# Criando O `response.go`

Este documento mostra como criar um pacote compartilhado para respostas HTTP.

O objetivo dele e padronizar:

- JSON de sucesso
- JSON de erro
- bind de request
- validacao com `validator`
- parse de parametros

---

## Onde Criar

Digite em:

- `internal/httpresponse/response.go`

---

## Responsabilidade Do Arquivo

Esse arquivo conhece `Gin`, mas nao conhece regra de negocio.

Ele e uma camada de apoio para os handlers.

---

## Estruturas Basicas

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
```

---

## Helpers De Resposta Simples

Digite em `internal/httpresponse/response.go`:

```go
func JSON(c *gin.Context, statusCode int, payload interface{}) {
	c.JSON(statusCode, payload)
}

func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func JSONErrorFromErr(c *gin.Context, statusCode int, err error) {
	JSONError(c, statusCode, err.Error())
}

func JSONAppError(c *gin.Context, err error) {
	JSONError(c, apperror.StatusCode(err), err.Error())
}

func AbortJSONError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Message: message,
	})
}
```

### Como Pensar

- `JSON(...)` responde sucesso
- `JSONError(...)` responde erro simples
- `JSONAppError(...)` usa o `apperror`
- `AbortJSONError(...)` interrompe o pipeline do Gin

---

## Bind E Validacao

Digite tambem:

```go
func BindAndValidateJSON(c *gin.Context, validate *validator.Validate, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, "invalid request body")
		return false
	}

	if validate == nil {
		validate = validator.New()
	}

	if err := validate.Struct(req); err != nil {
		JSONValidationError(c, err, req)
		return false
	}

	return true
}
```

### Fluxo

1. faz bind do JSON
2. se falhar, devolve `400`
3. valida struct tags
4. se falhar, devolve `400` com detalhes
5. retorna `true` quando esta tudo certo

---

## Erro De Validacao

Digite:

```go
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
```

---

## Parse De Parametros

Digite:

```go
func ParseIntParam(c *gin.Context, name string) (int, bool) {
	value := c.Param(name)
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", name))
		return 0, false
	}

	return parsedValue, true
}

func ParseInt64Param(c *gin.Context, name string) (int64, bool) {
	value := c.Param(name)
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", name))
		return 0, false
	}

	return parsedValue, true
}

func ParseMinInt64Query(c *gin.Context, name string, defaultValue string, min int64) (int64, bool) {
	rawValue := c.DefaultQuery(name, defaultValue)
	parsedValue, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil || parsedValue < min {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer greater than or equal to %d", name, min))
		return 0, false
	}

	return parsedValue, true
}
```

---

## Helpers Internos De Validacao

Digite tambem:

```go
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
	case "max":
		return fmt.Sprintf("must have at most %s characters", fieldError.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fieldError.Param())
	case "eqfield":
		targetField := fieldNames[fieldError.Param()]
		if targetField == "" {
			targetField = strings.ToLower(fieldError.Param())
		}

		return fmt.Sprintf("must match %s", targetField)
	default:
		return "is invalid"
	}
}
```

---

## Exemplo De Uso No Handler

```go
var req dto.CreateBuyerRequest
if ok := httpresponse.BindAndValidateJSON(c, h.validate, &req); !ok {
	return
}

buyerID, err := h.service.Create(c.Request.Context(), &req)
if err != nil {
	httpresponse.JSONAppError(c, err)
	return
}

httpresponse.JSON(c, http.StatusCreated, gin.H{
	"id": buyerID,
})
```

---

## Proximo Documento

Depois de `response.go`, o proximo passo natural e montar:

- `internal/server/router.go`
- `internal/handler/health/handler.go`

Assim a aplicacao passa a ter bootstrap HTTP coerente.
