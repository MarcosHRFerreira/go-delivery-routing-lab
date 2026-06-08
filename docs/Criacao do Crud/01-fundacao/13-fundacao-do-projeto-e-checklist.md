# Fundacao Do Projeto E Checklist

Este documento fecha a trilha base para montar uma API do zero antes dos CRUDs.

Ele serve como referencia de apostila para explicar a ordem de construcao do projeto.

---

## Ordem Recomendada

Antes de criar `buyers`, `couriers`, `orders` e os demais modulos, a base ideal e:

1. `internal/config/config.go`
2. `internal/apperror/error.go`
3. `internal/httpresponse/response.go`
4. `internal/handler/health/handler.go`
5. `internal/server/router.go`
6. `cmd/main.go`
7. conexao com MySQL
8. migrations
9. primeiro CRUD

---

## Papel De Cada Camada

### Bootstrap

- `cmd/main.go`
- `internal/config/config.go`
- `internal/server/router.go`

Essa parte sobe a aplicacao.

### Infra Compartilhada

- `internal/apperror/error.go`
- `internal/httpresponse/response.go`
- `pkg/internalsql`

Essa parte da base para os modulos de negocio.

### Primeiros Handlers

- `internal/handler/health/handler.go`
- `internal/handler/buyer/...`

Aqui comecam os endpoints.

---

## Fluxo Mental Da Aplicacao

Use este fluxo como regra geral da apostila:

1. `main.go` inicia tudo
2. `config.go` carrega ambiente
3. `internalsql` conecta no banco
4. `router.go` registra handlers
5. `handler` recebe HTTP
6. `httpresponse` faz bind/validate/resposta
7. `service` aplica regra de negocio
8. `repository` fala com o banco
9. `apperror` padroniza os erros

---

## Checklist De Bootstrap

Marque cada item quando terminar:

- criar `internal/config/config.go`
- criar `internal/apperror/error.go`
- criar `internal/httpresponse/response.go`
- instalar `gin`
- instalar `validator`
- instalar `godotenv`
- criar `internal/handler/health/handler.go`
- criar `internal/server/router.go`
- criar `cmd/main.go`
- criar conexao MySQL em `pkg/internalsql`
- subir o banco local
- testar `GET /check-health`

---

## Checklist Antes Do Primeiro CRUD

- migrations aplicadas com sucesso
- `buyers` criado no banco
- `router` recebendo `DB`
- `validator` funcionando
- `response.go` padronizado
- `apperror` funcionando
- health check respondendo `200`

---

## Material Complementar Da Apostila

Depois desta base, os proximos capitulos naturais sao:

- `14-ordem-cronologica-da-construcao-da-api.md`
- `01-crud-buyers.md`
- `02-crud-couriers.md`
- `03-crud-orders.md`
- `04-crud-deliveries.md`
- `05-crud-courier-locations.md`
- `06-delivery-reorder.md`
- `07-guia-unico-testes-fase1.md`

---

## Evolucoes Futuras Interessantes

Quando a apostila avancar, vale documentar tambem:

- autenticacao JWT
- middlewares compartilhados
- logger estruturado
- padrao oficial de logs do projeto
- transacoes
- paginacao padrao
- Swagger ou OpenAPI
- testes de integracao com banco real
- observabilidade
- recovery global
- deploy com Docker e ambientes
- CI/CD e pipeline de testes

---

## Fechamento

Se a pessoa souber montar bem esses arquivos base, ela ja consegue sair de um projeto vazio para uma API organizada.

Os CRUDs entram depois como extensao natural da mesma arquitetura.
