# Guia Unico de Testes da Fase 1

## Objetivo

Este documento consolida a estrategia de testes da fase 1 do projeto para servir como referencia futura durante a implementacao.

Ele existe para responder estas perguntas:

1. o que deve ser testado em cada modulo?
2. onde cada tipo de teste deve ficar?
3. como separar teste unitario de teste de integracao HTTP?
4. quais cenarios sao mais importantes para a fase 1?
5. como montar uma base reutilizavel de stubs e helpers?

Este guia cobre os modulos documentados ate aqui:

- `buyers`
- `couriers`
- `orders`
- `deliveries`
- `courier_locations`
- `delivery_reorder`

---

## Visao Geral

Na fase 1, o ideal e trabalhar com dois grandes grupos de teste:

- `test/unit`
- `test/integration`

Separacao conceitual:

- `unit`: testa regra de negocio da camada `service`
- `integration`: testa o fluxo HTTP real dos `handlers`, `dto`, `httpresponse` e roteamento

Fluxo:

```text
Unit: service -> stub repositories
Integration: http request -> router -> handler -> stub services
```

---

## Estrutura Recomendada

Para manter a organizacao simples e escalavel:

```text
test/
  unit/
    service_helpers_test.go
    buyer_service_test.go
    courier_service_test.go
    order_service_test.go
    delivery_service_test.go
    location_service_test.go
    routing_service_test.go
  integration/
    http_helpers_test.go
    buyer_endpoints_test.go
    courier_endpoints_test.go
    order_endpoints_test.go
    delivery_endpoints_test.go
    location_endpoints_test.go
    routing_endpoints_test.go
```

### Ideia Principal

- `service_helpers_test.go`: concentra stubs e helpers compartilhados
- `http_helpers_test.go`: monta o router de teste e helpers de request/resposta
- um arquivo por modulo facilita leitura e manutencao

---

## O Que Testar Em Cada Camada

## Testes Unitarios

Os testes unitarios devem focar em:

- regra de negocio
- cenarios principais
- cenarios alternativos
- coordenacao entre repositories
- transicao de status

Eles **nao** devem depender de:

- banco real
- servidor HTTP real
- serializacao JSON real

### O Que O Unit Test Deve Confirmar

- o `service` chama os repositories corretos
- erros de negocio sao retornados com a semantica esperada
- dados derivados sao montados corretamente
- transicoes invalidas sao bloqueadas

---

## Testes de Integracao HTTP

Os testes de integracao HTTP devem focar em:

- rota correta
- parse de path param
- bind do JSON
- validacao do DTO
- resposta HTTP final
- serializacao JSON

Eles **nao** precisam:

- bater no banco real
- testar a logica interna do repository
- duplicar toda a cobertura do `service`

### O Que O Integration Test Deve Confirmar

- `400` para body invalido
- `400` para path param invalido
- `404` para erros de recurso inexistente propagados pelo service
- `200`, `201` e `204` nos fluxos felizes

---

## Base Compartilhada dos Testes Unitarios

Arquivo sugerido:

- `test/unit/service_helpers_test.go`

Esse arquivo pode concentrar:

- stubs de `BuyerRepository`
- stubs de `CourierRepository`
- stubs de `OrderRepository`
- stubs de `AddressRepository`
- stubs de `DeliveryRepository`
- stubs de `LocationRepository`
- stubs de `ReorderHistoryRepository`
- helpers para verificar `apperror.StatusCode(err)`

### Exemplo de abordagem

Cada stub pode ter:

- funcao customizavel por metodo
- lista de chamadas recebidas
- dados de retorno configuraveis

Exemplo didatico:

```go
type buyerRepositoryStub struct {
	getBuyerByIDFn func(ctx context.Context, buyerID int64) (*model.BuyerModel, error)
}

func (s *buyerRepositoryStub) GetBuyerByID(ctx context.Context, buyerID int64) (*model.BuyerModel, error) {
	if s.getBuyerByIDFn != nil {
		return s.getBuyerByIDFn(ctx, buyerID)
	}

	return nil, nil
}
```

### Beneficio

Com essa base:

- os testes ficam menores
- o foco vai para o comportamento
- voce evita repeticao desnecessaria

---

## Base Compartilhada dos Testes de Integracao

Arquivo sugerido:

- `test/integration/http_helpers_test.go`

Esse arquivo pode concentrar:

- stubs de `BuyerService`
- stubs de `CourierService`
- stubs de `OrderService`
- stubs de `DeliveryService`
- stubs de `LocationService`
- stubs de `RoutingService`
- helper para request JSON
- helper para decode de resposta
- helper para montar `Gin` de teste

### O Que O Router de Teste Deve Fazer

Ele deve montar:

- `gin.New()`
- `validator.New()`
- handlers reais do projeto
- services stubados

Assim, quando o teste fizer request com `httptest`, ele passa pelo fluxo HTTP real do projeto:

```text
httptest -> gin router -> handler -> dto -> httpresponse
```

---

## Estrategia Por Modulo

## Buyers

### Unitario

Arquivo sugerido:

- `test/unit/buyer_service_test.go`

Fluxos principais:

- criar comprador com sucesso
- buscar comprador existente por ID
- listar compradores
- atualizar comprador existente
- remover comprador existente

Fluxos alternativos:

- comprador nao encontrado no `GetByID`
- comprador nao encontrado no `Update`
- comprador nao encontrado no `Delete`
- erro tecnico do repository no `Create`

### Integracao HTTP

Arquivo sugerido:

- `test/integration/buyer_endpoints_test.go`

Fluxos principais:

- `POST /buyers` retorna `201`
- `GET /buyers/{buyer_id}` retorna `200`
- `GET /buyers` retorna `200`
- `PUT /buyers/{buyer_id}` retorna `200`
- `DELETE /buyers/{buyer_id}` retorna `204`

Fluxos alternativos:

- `POST /buyers` com JSON invalido retorna `400`
- `GET /buyers/abc` retorna `400`
- `GET /buyers/{buyer_id}` com service retornando `NotFound` responde `404`

---

## Couriers

### Unitario

Arquivo sugerido:

- `test/unit/courier_service_test.go`

Fluxos principais:

- criar entregador com telefone unico
- buscar entregador por ID
- listar entregadores
- atualizar entregador
- atualizar status
- deletar entregador

Fluxos alternativos:

- criar com telefone duplicado
- atualizar com telefone ja usado por outro entregador
- atualizar status de entregador inexistente
- bloquear transicao de status invalida

### Integracao HTTP

Arquivo sugerido:

- `test/integration/courier_endpoints_test.go`

Fluxos principais:

- `POST /couriers` retorna `201`
- `GET /couriers/{courier_id}` retorna `200`
- `GET /couriers` retorna `200`
- `PUT /couriers/{courier_id}` retorna `200`
- `PATCH /couriers/{courier_id}/status` retorna `200`
- `DELETE /couriers/{courier_id}` retorna `204`

Fluxos alternativos:

- `POST /couriers` com `status` invalido retorna `400`
- `PATCH /couriers/{courier_id}/status` com body invalido retorna `400`
- `GET /couriers/abc` retorna `400`

---

## Orders

### Unitario

Arquivo sugerido:

- `test/unit/order_service_test.go`

Fluxos principais:

- criar pedido com comprador existente
- buscar pedido por ID
- listar pedidos
- atualizar pedido em status `created`
- marcar pedido como `ready_for_delivery`
- cancelar pedido em status `created`

Fluxos alternativos:

- criar pedido com comprador inexistente
- atualizar pedido inexistente
- atualizar pedido que nao esta em `created`
- mover para `ready_for_delivery` com status invalido
- cancelar pedido com status invalido
- endereco nao encontrado durante leitura ou update

### Integracao HTTP

Arquivo sugerido:

- `test/integration/order_endpoints_test.go`

Fluxos principais:

- `POST /orders` retorna `201`
- `GET /orders/{order_id}` retorna `200`
- `GET /orders` retorna `200`
- `PUT /orders/{order_id}` retorna `200`
- `PATCH /orders/{order_id}/ready-for-delivery` retorna `200`
- `PATCH /orders/{order_id}/cancel` retorna `200`

Fluxos alternativos:

- `POST /orders` com `buyer_id` ausente retorna `400`
- `PUT /orders/abc` retorna `400`
- `PATCH /orders/{order_id}/ready-for-delivery` com service retornando `BadRequest` responde `400`

---

## Deliveries

### Unitario

Arquivo sugerido:

- `test/unit/delivery_service_test.go`

Fluxos principais:

- atribuir pedido pronto para entrega
- listar entregas do entregador
- iniciar entrega em status `assigned`
- concluir entrega em status `out_for_delivery`
- falhar entrega em status `out_for_delivery`

Fluxos alternativos:

- atribuir pedido inexistente
- atribuir entregador inexistente
- atribuir pedido que nao esta `ready_for_delivery`
- atribuir pedido que ja possui entrega
- iniciar entrega com status invalido
- concluir entrega com status invalido
- falhar entrega com status invalido

### Integracao HTTP

Arquivo sugerido:

- `test/integration/delivery_endpoints_test.go`

Fluxos principais:

- `POST /deliveries/assign` retorna `201`
- `GET /couriers/{courier_id}/deliveries` retorna `200`
- `PATCH /deliveries/{delivery_id}/start` retorna `200`
- `PATCH /deliveries/{delivery_id}/complete` retorna `200`
- `PATCH /deliveries/{delivery_id}/fail` retorna `200`

Fluxos alternativos:

- `POST /deliveries/assign` com body invalido retorna `400`
- `PATCH /deliveries/abc/start` retorna `400`
- service retornando `NotFound` propaga `404`
- service retornando `BadRequest` propaga `400`

---

## Courier Locations

### Unitario

Arquivo sugerido:

- `test/unit/location_service_test.go`

Fluxos principais:

- registrar localizacao para entregador existente
- buscar ultima localizacao registrada

Fluxos alternativos:

- registrar localizacao para entregador inexistente
- latitude fora da faixa valida
- longitude fora da faixa valida
- consultar ultima localizacao inexistente

### Integracao HTTP

Arquivo sugerido:

- `test/integration/location_endpoints_test.go`

Fluxos principais:

- `POST /couriers/{courier_id}/location` retorna `201`
- `GET /couriers/{courier_id}/location/latest` retorna `200`

Fluxos alternativos:

- `POST /couriers/{courier_id}/location` com body invalido retorna `400`
- `POST /couriers/abc/location` retorna `400`
- `GET /couriers/{courier_id}/location/latest` com `NotFound` retorna `404`

---

## Delivery Reorder

### Unitario

Arquivo sugerido:

- `test/unit/routing_service_test.go`

Fluxos principais:

- reorder com entregador existente e localizacao valida
- consulta de sequencia atual

Fluxos alternativos:

- reorder com entregador inexistente
- reorder sem localizacao atual
- reorder sem entregas pendentes
- falha ao carregar pedido
- falha ao carregar endereco
- falha ao persistir `current_sequence`
- falha ao gravar historico

### O Que Confirmar no Unitario

- score e calculado corretamente
- entregas sao ordenadas corretamente
- `current_sequence` e persistido na ordem correta
- `delivery_reorder_history` recebe os registros esperados

### Integracao HTTP

Arquivo sugerido:

- `test/integration/routing_endpoints_test.go`

Fluxos principais:

- `POST /couriers/{courier_id}/deliveries/reorder` retorna `200`
- `GET /couriers/{courier_id}/deliveries/sequence` retorna `200`

Fluxos alternativos:

- `POST /couriers/abc/deliveries/reorder` retorna `400`
- service retornando `BadRequest` por falta de localizacao retorna `400`
- service retornando `NotFound` para entregador retorna `404`

---

## Ordem Recomendada de Implementacao dos Testes

Quando voce comecar a codificar, uma ordem eficiente e:

1. `buyers`
2. `couriers`
3. `orders`
4. `deliveries`
5. `courier_locations`
6. `delivery_reorder`
7. testes de integracao HTTP

### Motivo

- `buyers` e `couriers` sao os modulos mais simples
- `orders` introduz coordenacao entre tabelas
- `deliveries` introduz transicao de status operacional
- `locations` fecha a base de dados operacionais
- `reorder` reaproveita tudo o que veio antes

---

## Padrao de Nome dos Testes

Use nomes de teste que deixem claro:

- o comportamento
- o contexto
- o resultado esperado

Exemplos:

```go
func TestBuyerService_Create_ShouldCreateBuyerWhenInputIsValid(t *testing.T)
func TestCourierService_UpdateStatus_ShouldReturnBadRequestWhenTransitionIsInvalid(t *testing.T)
func TestOrderService_Create_ShouldReturnNotFoundWhenBuyerDoesNotExist(t *testing.T)
func TestDeliveryService_Assign_ShouldReturnBadRequestWhenOrderIsNotReadyForDelivery(t *testing.T)
func TestRoutingService_ReorderDeliveries_ShouldPersistSequenceInScoreOrder(t *testing.T)
```

Se preferir um estilo mais enxuto:

```go
func TestCreateBuyer_Success(t *testing.T)
func TestAssignDelivery_OrderNotReady(t *testing.T)
```

O importante e manter consistencia.

---

## O Que Verificar em Cada Teste Unitario

Checklist minimo por teste:

- arrange claro
- act unico
- assert objetivo

### Bons asserts

- status do erro via `apperror.StatusCode(err)`
- campos principais do retorno
- efeitos sobre stubs
- quantidade de chamadas quando fizer sentido

### Evite

- testar muitos comportamentos em um unico teste
- asserts redundantes demais
- dependencia entre testes

---

## O Que Verificar nos Testes de Integracao

Checklist minimo por teste HTTP:

- status code
- corpo JSON
- mensagem de erro quando aplicavel
- campo de resposta mais importante do fluxo

### Exemplos

No `POST /buyers`:

- `201`
- body contem `id`

No `GET /buyers/abc`:

- `400`
- body contem `message`

No `PATCH /orders/{order_id}/cancel`:

- `200`
- body contem `status = cancelled`

---

## Helpers Recomendados

## `test/unit/service_helpers_test.go`

Esse arquivo pode concentrar:

- factory de `time.Time`
- stubs compartilhados
- helper `assertStatusCode`

Exemplo:

```go
func assertStatusCode(t *testing.T, err error, expected int) {
	t.Helper()

	got := apperror.StatusCode(err)
	if got != expected {
		t.Fatalf("expected status %d, got %d", expected, got)
	}
}
```

## `test/integration/http_helpers_test.go`

Esse arquivo pode concentrar:

- `newTestMux(...)`
- helper para enviar JSON
- helper para decodificar resposta

Exemplo:

```go
func newJSONRequest(t *testing.T, method string, url string, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}
```

---

## Comando de Validacao

Quando comecar a implementar os testes, a validacao minima deve ser:

```bash
go test ./...
```

Quando o projeto crescer, vale adicionar tambem:

```bash
golangci-lint run
```

E, durante a escrita:

- verificar diagnosticos do editor
- confirmar que interfaces continuam satisfeitas
- revisar imports nao usados

---

## O Que Nao Vale a Pena Testar Agora

Na fase 1, nao precisa exagerar em:

- benchmark
- carga
- concorrencia pesada
- comparacao fina de performance
- integrações reais com mapas

O foco da fase 1 e:

- estabilidade dos fluxos principais
- previsibilidade das regras de negocio
- contrato HTTP correto

---

## Mapa Final de Cobertura

Se voce seguir este guia, a cobertura conceitual da fase 1 fica assim:

- `buyers`: CRUD basico validado
- `couriers`: CRUD + status validado
- `orders`: criacao + update + transicoes validado
- `deliveries`: atribuicao + fluxo operacional validado
- `courier_locations`: localizacao atual validada
- `delivery_reorder`: heuristica e persistencia da sequencia validada

---

## Checklist Final de Referencia

Antes de considerar a fase 1 bem coberta, confirme se existem:

- testes unitarios para todos os `services`
- ao menos um fluxo alternativo por caso de uso
- testes HTTP para todos os endpoints principais
- validacao de `400`, `404`, `200`, `201` e `204`
- helpers compartilhados para evitar repeticao
- comandos de execucao documentados

---

## Encerramento do Estudo

Com este documento, a trilha de estudo da fase 1 fica organizada em:

- `01-crud-buyers.md`
- `02-crud-couriers.md`
- `03-crud-orders.md`
- `04-crud-deliveries.md`
- `05-crud-courier-locations.md`
- `06-delivery-reorder.md`
- `07-guia-unico-testes-fase1.md`

Isso ja forma uma base muito boa para voce comecar a implementacao depois com referencia clara de:

- arquitetura
- responsabilidades por camada
- fluxo operacional
- estrategia de testes

---

## Proximo Passo Futuro

Quando voce voltar a codificar, os caminhos mais naturais serao:

- comecar pelo modulo `buyers`
- criar a base de `apperror` e `httpresponse`
- montar o `main.go`
- implementar os testes junto com cada modulo

Este documento pode ficar como a referencia principal de testes da fase 1.
