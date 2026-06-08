# Criando O `main.go`

Este documento mostra como criar o ponto de entrada da aplicacao.

O `main.go` existe para orquestrar o bootstrap, nao para concentrar regra de negocio.

---

## Onde Criar

Digite em:

- `cmd/main.go`

---

## Responsabilidade Do Arquivo

O `main.go` deve:

- carregar configuracao
- abrir conexao com banco
- montar o `router`
- subir o servidor HTTP
- desligar a aplicacao com seguranca

O `main.go` nao deve:

- escrever SQL
- validar request
- implementar endpoint
- conhecer detalhes do CRUD

---

## Estrutura Sugerida

Digite em `cmd/main.go`:

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/config"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/internal/server"
	"github.com/MarcosHRFerreira/go-delivery-routing-lab/pkg/internalsql"
	"github.com/go-playground/validator/v10"
)

const (
	serverReadHeaderTimeout = 2 * time.Second
	serverReadTimeout       = 10 * time.Second
	serverWriteTimeout      = 15 * time.Second
	serverIdleTimeout       = 60 * time.Second
	shutdownTimeout         = 10 * time.Second
)

func main() {
	validate := validator.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := internalsql.ConnectMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	router := server.NewRouter(validate, server.Dependencies{
		DB:            db,
		HealthChecker: db,
	})

	httpServer := &http.Server{
		Addr:              cfg.ServerAddress(),
		Handler:           router,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("starting http server on %s", cfg.ServerAddress())
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		close(serverErrors)
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("http server stopped unexpectedly: %v", err)
		}
	case <-shutdownSignal.Done():
		log.Print("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("failed to shutdown http server: %v", err)
	}

	log.Print("server stopped gracefully")
}
```

---

## Leitura Do Fluxo

1. monta `validator`
2. carrega `config`
3. conecta no banco
4. cria o `router`
5. monta `http.Server`
6. sobe a API em goroutine
7. aguarda erro ou sinal do sistema
8. executa `Shutdown(...)`

---

## Por Que Isso E Melhor Do Que Um `main.go` Simples

- separa bootstrap de negocio
- facilita testes
- permite graceful shutdown
- configura timeouts do servidor
- prepara o projeto para crescer

---

## Onde Vale Adicionar Log No `main.go`

Quando voce for incrementar logging no codigo, o `main.go` e um dos primeiros pontos que merecem log.

Aqui normalmente vale registrar:

- inicio do bootstrap da aplicacao
- falha ao carregar configuracao
- falha ao conectar no banco
- inicio do servidor HTTP com endereco e porta
- recebimento de sinal de desligamento
- falha ou sucesso no shutdown

Boa regra pratica:

- `main.go` loga eventos de ciclo de vida da aplicacao
- `main.go` nao deve tentar logar cada request HTTP

Esse detalhamento de request entra melhor em middleware.

---

## Erros Comuns

- esquecer de passar `DB: db` para o `router`
- usar `router.Run(...)` e perder parte do controle do servidor
- esquecer `defer db.Close()`
- nao tratar `http.ErrServerClosed`
- nao usar timeout no desligamento

---

## Proximo Documento

Depois de `main.go`, vale documentar um material de apoio com:

- arquitetura base
- ordem de criacao dos arquivos
- checklist de bootstrap

Esse sera o guia de referencia para a apostila.
