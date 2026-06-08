# Criando O `router.go` E O `health check`

Este documento mostra como ligar os modulos da aplicacao no Gin.

Ele cobre:

- `internal/server/router.go`
- `internal/handler/health/handler.go`

Sem esses arquivos, `main.go` sobe o servidor, mas a API nao ganha rotas de verdade.

---

## `internal/server/router.go`

### Responsabilidade

Esse arquivo monta as dependencias da camada HTTP.

Ele nao deve:

- carregar configuracao
- abrir conexao com banco
- chamar `ListenAndServe`

Ele deve:

- criar o `gin.Engine`
- registrar middlewares
- registrar handlers
- montar repositories e services

---

## Estrutura Sugerida

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

	db := deps.DB
	buyerRepo := buyerrepository.NewRepository(db)
	buyerService := buyerservice.NewService(buyerRepo)
	buyerHandler := buyerhandler.NewHandler(router, validate, buyerService)
	buyerHandler.RouteList()

	return router
}
```

### Leitura Do Fluxo

1. cria `validator` se ele nao foi injetado
2. cria `gin.New()`
3. registra `gin.Recovery()`
4. registra `health check`
5. se existir banco, monta os modulos reais
6. devolve o `router`

---

## Por Que O `Dependencies` E Importante

Essa struct facilita:

- teste de integracao
- bootstrap por ambiente
- crescimento do projeto sem acoplamento no `main.go`

Com o tempo, ela pode receber:

- logger
- jwt service
- clock
- clients externos

---

## Onde O Log Entra No `router.go`

Quando voce for evoluir a aplicacao com logging, o `router.go` vira o ponto central para registrar middlewares transversais.

O caminho mais natural e:

1. injetar um logger em `Dependencies`
2. registrar middleware de `request id`
3. registrar middleware de logging HTTP
4. manter recovery na cadeia
5. so depois ligar os handlers de dominio

Na pratica, o `router.go` costuma ser o lugar ideal para:

- ligar logger compartilhado
- registrar logging middleware
- garantir que todos os endpoints passem pela mesma estrategia de log

Evite colocar no `router.go`:

- logica de negocio
- log por operacao de CRUD
- log de SQL especifico

O papel dele e montar a infraestrutura HTTP comum.

---

## `internal/handler/health/handler.go`

### Responsabilidade

Esse handler serve para verificar se a aplicacao e o banco estao disponiveis.

Ele e simples, mas muito importante para:

- monitoramento
- ambiente Docker
- deploy futuro

---

## Estrutura Sugerida

Digite em `internal/handler/health/handler.go`:

```go
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/httpresponse"
	"github.com/gin-gonic/gin"
)

type Checker interface {
	PingContext(ctx context.Context) error
}

type Handler struct {
	router  gin.IRouter
	checker Checker
	timeout time.Duration
}

func NewHandler(router gin.IRouter, checker Checker, timeout time.Duration) *Handler {
	return &Handler{
		router:  router,
		checker: checker,
		timeout: timeout,
	}
}

func (h *Handler) RouteList() {
	h.router.GET("/check-health", h.Check)
}

func (h *Handler) Check(c *gin.Context) {
	if h.checker == nil {
		httpresponse.JSONError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.checker.PingContext(ctx); err != nil {
		httpresponse.JSONError(c, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	httpresponse.JSON(c, http.StatusOK, gin.H{
		"message": "service is healthy",
	})
}
```

---

## O Que Esse Handler Ensina Para A Apostila

Mesmo um handler simples ja mostra o padrao completo:

- dependencia por interface
- uso de `gin.IRouter`
- uso de `httpresponse`
- uso de `context.WithTimeout(...)`

---

## Log Recomendado Para O Health Check

Quando voce for implementar logs, o `health check` nao precisa ser verboso em toda chamada saudavel.

Uma estrategia equilibrada costuma ser:

- logar falha de dependencia quando `PingContext(...)` retornar erro
- logar indisponibilidade quando `checker` estiver ausente
- evitar poluir o log com `200 OK` de health check a cada request

Se quiser registrar chamadas saudaveis, prefira deixar isso no middleware HTTP com baixo nivel de log.

---

## Erros Comuns

- esquecer de registrar `RouteList()`
- subir o `router` sem passar `DB` em `Dependencies`
- colocar SQL direto no `handler`
- fazer `main.go` registrar rotas manualmente

---

## Proximo Documento

Depois de `router.go` e `health check`, o passo natural e fechar o bootstrap com:

- `cmd/main.go`
