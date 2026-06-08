# Criando O `error.go`

Este documento mostra como criar o pacote de erro compartilhado da aplicacao.

O objetivo dele e evitar:

- strings soltas espalhadas pelo projeto
- `http.Status...` repetido em todo lugar
- handlers decidindo regras de erro por conta propria

---

## Onde Criar

Digite em:

- `internal/apperror/error.go`

---

## Responsabilidade Do Arquivo

O pacote `apperror` representa erros da aplicacao com semantica HTTP.

Ele nao escreve resposta HTTP.

Ele apenas carrega:

- `status code`
- `message`
- `cause`

---

## Estrutura Sugerida

```go
package apperror

import "net/http"

type Error struct {
	statusCode int
	message    string
	cause      error
}
```

### Leitura Da Estrutura

- `statusCode` diz qual status o handler deve devolver
- `message` e o texto seguro para o cliente
- `cause` guarda o erro tecnico original

---

## Construtores

Digite em `internal/apperror/error.go`:

```go
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

func Internal(message string, cause error) *Error {
	return Wrap(http.StatusInternalServerError, message, cause)
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, message)
}
```

---

## Metodos Da Struct

Digite tambem:

```go
func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) Unwrap() error {
	return e.cause
}
```

### Por Que Esses Metodos Existem

- `Error()` faz a struct implementar a interface `error`
- `StatusCode()` permite o `httpresponse` extrair o status certo
- `Unwrap()` preserva a causa para debug futuro

---

## Helper Global

Digite por fim:

```go
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if asStatusCode, ok := err.(interface{ StatusCode() int }); ok {
		return asStatusCode.StatusCode()
	}

	return http.StatusInternalServerError
}
```

Esse helper e importante porque o `handler` ou o `httpresponse` nao precisa conhecer o tipo concreto do erro.

---

## Exemplo De Uso No Service

```go
if conflictErr := mapBuyerConflictError(err); conflictErr != nil {
	return 0, conflictErr
}
```

```go
buyerID, err := s.buyerRepo.CreateBuyer(ctx, buyerModel)
if err != nil {
	return 0, apperror.Internal("failed to create buyer", err)
}
```

---

## Regra De Ouro

Use `apperror` no `service` e em outras camadas de aplicacao.

Nao use `gin.Context` aqui.

O pacote precisa continuar agnostico de framework HTTP.

---

## Proximo Documento

Depois de `error.go`, o proximo passo natural e criar:

- `internal/httpresponse/response.go`

Esse pacote vai traduzir `apperror` e validacao em resposta HTTP padronizada.
