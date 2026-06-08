# Estudo De Sistema De Roteirizacao De Entregas Em Go

## Objetivo

Este material descreve como projetar um novo backend em `Golang` para gerenciar pedidos e organizar entregas para entregadores, usando como referencia a estrutura arquitetural do projeto atual.

O cenario funcional e o seguinte:

- o sistema recebe pedidos com os dados do comprador
- cada pedido possui um endereco de entrega
- o entregador recebe uma lista de pedidos atribuidos
- o sistema captura a posicao geografica atual do entregador
- o sistema reorganiza as entregas com base em `CEP + numero`
- sempre que o entregador solicitar, a classificacao das entregas e recalculada

O objetivo nao e apenas criar um CRUD de pedidos, mas projetar um sistema com cara de mercado, preparado para evoluir em:

- arquitetura em camadas
- regras de negocio bem isoladas
- roteirizacao incremental
- observabilidade
- testes automatizados
- crescimento futuro para geolocalizacao mais inteligente

---

## Visao Geral Do Problema

O desafio principal deste projeto nao esta somente em armazenar pedidos, mas em organizar a fila de entregas de forma pratica para o entregador em campo.

O sistema precisa responder perguntas como:

- quais pedidos estao atribuidos a este entregador agora?
- em qual ordem ele deveria entregar?
- quando a posicao geografica muda, a ordem deve ser recalculada?
- como usar `CEP + numero` para aproximar uma sequencia de rota sem depender, no primeiro momento, de uma API externa de mapas?
- como manter o processo simples no `MVP`, mas com espaco para evoluir depois?

Esse tipo de projeto e interessante porque mistura:

- modelagem de dominio
- regras operacionais
- localizacao
- ordenacao de fila
- eventos de campo
- consistencia entre pedido, entregador e entrega

---

## Diagnostico Do Dominio

### Entidades Principais

O dominio pode ser dividido em alguns blocos centrais:

1. `comprador`
2. `pedido`
3. `endereco de entrega`
4. `entregador`
5. `lote de entregas`
6. `roteirizacao`
7. `evento de localizacao`

### Conceitos Importantes

#### Pedido

Representa a compra que precisa ser entregue.

Campos essenciais:

- `id`
- `codigo`
- `status`
- `buyer_id`
- `delivery_address_id`
- `assigned_courier_id`
- `created_at`
- `updated_at`

#### Comprador

Representa a pessoa que realizou a compra.

Campos essenciais:

- `id`
- `name`
- `document`
- `phone`
- `email`

#### Endereco De Entrega

Representa o local para onde o pedido sera entregue.

Campos essenciais:

- `id`
- `zip_code`
- `street`
- `number`
- `complement`
- `district`
- `city`
- `state`
- `latitude`
- `longitude`

Mesmo que no `MVP` a classificacao use `CEP + numero`, vale a pena deixar `latitude` e `longitude` previstos no modelo para evolucao futura.

#### Entregador

Representa o profissional responsavel por transportar os pedidos.

Campos essenciais:

- `id`
- `name`
- `phone`
- `vehicle_type`
- `status`

#### Posicao Do Entregador

Representa a localizacao atual do entregador.

Campos essenciais:

- `courier_id`
- `latitude`
- `longitude`
- `recorded_at`

#### Entrega

Representa o vinculo operacional entre pedido e processo de entrega.

Campos essenciais:

- `id`
- `order_id`
- `courier_id`
- `status`
- `current_sequence`
- `last_reordered_at`

#### Classificacao De Entrega

Representa o resultado do algoritmo de ordenacao em determinado momento.

Campos essenciais:

- `delivery_id`
- `sequence`
- `score`
- `reason`
- `generated_at`

---

## Regras De Negocio Iniciais

Para o `MVP`, o sistema pode seguir regras objetivas e simples:

1. um pedido so pode ser entregue por um entregador por vez
2. apenas pedidos com status `ready_for_delivery` podem entrar na fila de entrega
3. o entregador so visualiza pedidos atribuidos a ele
4. a ordenacao pode ser recalculada sob demanda pelo entregador
5. a classificacao usa como base:
   - posicao geografica atual do entregador
   - agrupamento por `CEP`
   - ordenacao interna por `numero`
6. a reclassificacao nao altera o endereco do pedido, apenas a ordem sugerida de visita
7. pedidos ja entregues ou cancelados nao participam da nova classificacao
8. pedidos com dados de endereco incompletos devem ser marcados como `nao roteirizaveis`

### Status Sugeridos

#### Status Do Pedido

- `created`
- `paid`
- `preparing`
- `ready_for_delivery`
- `out_for_delivery`
- `delivered`
- `cancelled`

#### Status Da Entrega

- `pending_assignment`
- `assigned`
- `sequenced`
- `in_progress`
- `completed`
- `failed`

#### Status Do Entregador

- `available`
- `busy`
- `offline`

---

## Importante Sobre A Regra De CEP + Numero

Usar `CEP + numero` e uma heuristica operacional valida para `MVP`, mas ela tem limitacoes claras.

### Vantagens

- simples de implementar
- barata
- nao depende de API externa de mapas
- suficiente para bairros ou rotas urbanas com padrao razoavel de numeracao
- boa para validar o produto rapidamente

### Limitacoes

- `CEP` nem sempre representa proximidade real
- ruas com mesmo `CEP` podem estar em pontos diferentes
- numero do imovel nem sempre cresce na direcao real da rota
- condominios, predios e enderecos irregulares podem quebrar a heuristica
- sem geocodificacao completa, a ordem sera aproximada, nao otima

### Decisao Profissional

Para um projeto inicial, a melhor estrategia e:

1. comecar com `CEP + numero`
2. manter suporte no modelo para `latitude` e `longitude`
3. isolar o algoritmo em um `service`
4. permitir trocar depois por:
   - geocoding
   - matriz de distancia
   - API de mapas
   - algoritmo de menor custo

Essa decisao reduz complexidade cedo demais, sem fechar a porta para maturidade futura.

---

## Fluxo Funcional Principal

### 1. Criacao Do Pedido

1. o sistema recebe os dados do comprador
2. recebe o endereco de entrega
3. cria o pedido com status inicial
4. quando o pedido estiver pronto, ele muda para `ready_for_delivery`

### 2. Atribuicao Ao Entregador

1. um operador ou rotina automatica atribui pedidos ao entregador
2. o sistema cria ou atualiza o registro de entrega
3. os pedidos ficam disponiveis na fila do entregador

### 3. Atualizacao Da Posicao Do Entregador

1. o app do entregador envia `latitude` e `longitude`
2. o sistema salva a localizacao atual
3. essa posicao passa a ser usada como ponto de referencia para reclassificacao

### 4. Reclassificacao Das Entregas

1. o entregador solicita reorganizacao
2. o backend busca:
   - entregador
   - ultima localizacao
   - pedidos ativos atribuidos
3. o algoritmo agrupa por proximidade aproximada e ordena por `CEP + numero`
4. o sistema retorna a nova sequencia sugerida
5. opcionalmente, persiste essa sequencia

### 5. Execucao Da Rota

1. o entregador inicia a entrega
2. confirma cada parada como entregue ou falha
3. o sistema atualiza status e recalcula a fila restante quando solicitado

---

## Arquitetura Recomendada

A estrutura pode seguir a mesma filosofia do seu projeto atual:

```text
cmd/
internal/
  handler/
  service/
  repository/
  dto/
  entity/
  apperror/
  middleware/
  observability/
pkg/
db/
test/
docs/
```

### Camadas

#### Handler

Responsabilidades:

- receber HTTP
- validar entrada
- converter payload em DTO
- delegar para o service
- devolver resposta padronizada

#### Service

Responsabilidades:

- aplicar regra de negocio
- coordenar entidades
- chamar repositories
- executar algoritmo de classificacao
- decidir quando recalcular ou bloquear operacoes

#### Repository

Responsabilidades:

- persistencia
- consultas SQL
- carregamento de pedidos, entregadores, enderecos e localizacao

#### Entity

Responsabilidades:

- representar conceitos do dominio
- encapsular invariantes
- validar transicoes de status quando fizer sentido

#### DTO

Responsabilidades:

- transportar dados entre HTTP e service
- separar contrato externo da modelagem interna

---

## Modulos Recomendados

### 1. Modulo De Compradores

Casos de uso:

- criar comprador
- buscar comprador
- atualizar comprador

### 2. Modulo De Pedidos

Casos de uso:

- criar pedido
- buscar pedido
- listar pedidos
- marcar pedido como pronto para entrega
- cancelar pedido

### 3. Modulo De Entregadores

Casos de uso:

- criar entregador
- consultar entregador
- atualizar status do entregador
- listar fila de entregas do entregador

### 4. Modulo De Localizacao

Casos de uso:

- registrar posicao atual do entregador
- consultar ultima posicao conhecida
- consultar historico de posicoes no futuro

### 5. Modulo De Roteirizacao

Casos de uso:

- classificar entregas
- recalcular ordem manualmente
- persistir sequencia sugerida
- retornar justificativa da classificacao

### 6. Modulo De Entregas

Casos de uso:

- atribuir pedido ao entregador
- iniciar entrega
- concluir entrega
- registrar falha de entrega
- reordenar fila remanescente

---

## Modelagem Inicial De Tabelas

### `buyers`

```sql
CREATE TABLE buyers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    document VARCHAR(50) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `addresses`

```sql
CREATE TABLE addresses (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    zip_code VARCHAR(10) NOT NULL,
    street VARCHAR(255) NOT NULL,
    number VARCHAR(20) NOT NULL,
    complement VARCHAR(255),
    district VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state VARCHAR(10) NOT NULL,
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `couriers`

```sql
CREATE TABLE couriers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL,
    status VARCHAR(30) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `orders`

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL UNIQUE,
    buyer_id BIGINT NOT NULL,
    delivery_address_id BIGINT NOT NULL,
    status VARCHAR(30) NOT NULL,
    total_amount DECIMAL(10, 2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orders_buyer FOREIGN KEY (buyer_id) REFERENCES buyers(id),
    CONSTRAINT fk_orders_address FOREIGN KEY (delivery_address_id) REFERENCES addresses(id)
);
```

### `deliveries`

```sql
CREATE TABLE deliveries (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id BIGINT NOT NULL UNIQUE,
    courier_id BIGINT NOT NULL,
    status VARCHAR(30) NOT NULL,
    current_sequence INT,
    last_reordered_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_deliveries_order FOREIGN KEY (order_id) REFERENCES orders(id),
    CONSTRAINT fk_deliveries_courier FOREIGN KEY (courier_id) REFERENCES couriers(id)
);
```

### `courier_locations`

```sql
CREATE TABLE courier_locations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    courier_id BIGINT NOT NULL,
    latitude DECIMAL(10, 7) NOT NULL,
    longitude DECIMAL(10, 7) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_locations_courier FOREIGN KEY (courier_id) REFERENCES couriers(id)
);
```

### `delivery_reorder_history`

```sql
CREATE TABLE delivery_reorder_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    courier_id BIGINT NOT NULL,
    delivery_id BIGINT NOT NULL,
    sequence_position INT NOT NULL,
    score DECIMAL(10, 4),
    reason VARCHAR(255),
    generated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_reorder_history_courier FOREIGN KEY (courier_id) REFERENCES couriers(id),
    CONSTRAINT fk_reorder_history_delivery FOREIGN KEY (delivery_id) REFERENCES deliveries(id)
);
```

---

## Endpoints Iniciais Sugeridos

### Compradores

- `POST /buyers`
- `GET /buyers/:buyer_id`

### Pedidos

- `POST /orders`
- `GET /orders/:order_id`
- `PATCH /orders/:order_id/ready-for-delivery`
- `PATCH /orders/:order_id/cancel`

### Entregadores

- `POST /couriers`
- `GET /couriers/:courier_id`
- `PATCH /couriers/:courier_id/status`

### Localizacao

- `POST /couriers/:courier_id/location`
- `GET /couriers/:courier_id/location/latest`

### Entregas

- `POST /deliveries/assign`
- `GET /couriers/:courier_id/deliveries`
- `PATCH /deliveries/:delivery_id/start`
- `PATCH /deliveries/:delivery_id/complete`
- `PATCH /deliveries/:delivery_id/fail`

### Roteirizacao

- `POST /couriers/:courier_id/deliveries/reorder`
- `GET /couriers/:courier_id/deliveries/sequence`

---

## DTOs Principais

### Criacao De Pedido

```json
{
  "buyer": {
    "name": "Maria Silva",
    "document": "12345678900",
    "phone": "11999999999",
    "email": "maria@email.com"
  },
  "delivery_address": {
    "zip_code": "01001000",
    "street": "Rua Exemplo",
    "number": "120",
    "complement": "Apto 12",
    "district": "Centro",
    "city": "Sao Paulo",
    "state": "SP"
  },
  "total_amount": 89.90
}
```

### Registro De Localizacao

```json
{
  "latitude": -23.550520,
  "longitude": -46.633308
}
```

### Atribuicao De Entrega

```json
{
  "order_id": 10,
  "courier_id": 3
}
```

### Resposta De Reclassificacao

```json
{
  "courier_id": 3,
  "generated_at": "2026-05-27T10:00:00Z",
  "deliveries": [
    {
      "delivery_id": 100,
      "order_id": 10,
      "sequence": 1,
      "zip_code": "01001000",
      "number": "120",
      "score": 1.2,
      "reason": "same_zip_and_closest_number"
    },
    {
      "delivery_id": 101,
      "order_id": 11,
      "sequence": 2,
      "zip_code": "01001000",
      "number": "150",
      "score": 1.8,
      "reason": "same_zip_and_next_number"
    }
  ]
}
```

---

## Como Projetar O Algoritmo Inicial

### Estrategia De MVP

O algoritmo pode ser simples, deterministico e explicavel.

Passos:

1. carregar a ultima localizacao do entregador
2. carregar as entregas pendentes atribuídas a ele
3. descartar entregas sem endereco utilizavel
4. agrupar por `CEP`
5. definir prioridade aproximada de grupo
6. dentro de cada grupo, ordenar por `numero`
7. gerar uma sequencia final

### Heuristica Sugerida

Uma abordagem inicial:

1. se houver `latitude` e `longitude` do endereco:
   - calcular distancia aproximada entre entregador e endereco
   - ordenar primeiro pelo grupo mais proximo
2. se nao houver coordenadas do endereco:
   - usar `CEP` como agrupador principal
   - ordenar lexicalmente ou numericamente pelo `CEP`
3. dentro do mesmo `CEP`:
   - ordenar pelo `numero` convertido para inteiro quando possivel
   - quando nao for numerico, aplicar fallback por texto

### Exemplo Conceitual Em Go

```go
type DeliveryCandidate struct {
	DeliveryID int64
	OrderID    int64
	ZipCode    string
	Number     string
	Latitude   *float64
	Longitude  *float64
}

func ReorderDeliveries(
	courierLat float64,
	courierLng float64,
	candidates []DeliveryCandidate,
) []DeliveryCandidate {
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]

		if left.ZipCode != right.ZipCode {
			return left.ZipCode < right.ZipCode
		}

		return parseAddressNumber(left.Number) < parseAddressNumber(right.Number)
	})

	return candidates
}
```

Esse exemplo ainda e simplificado demais, mas ja mostra a ideia de deixar a regra em uma funcao dedicada, testavel e independente da camada HTTP.

### Recomendacao Melhor

Crie um service especializado:

```text
internal/service/routing
```

Esse service pode expor algo como:

```go
type ReorderDeliveriesInput struct {
	CourierID int64
}

type ReorderDeliveriesOutput struct {
	GeneratedAt time.Time
	Deliveries  []DeliverySequence
}

type RoutingService interface {
	Reorder(ctx context.Context, input ReorderDeliveriesInput) (*ReorderDeliveriesOutput, error)
}
```

---

## Casos De Uso Principais

### 1. CreateOrder

Responsabilidades:

- validar comprador
- validar endereco
- criar pedido
- deixar pronto para fluxo operacional

### 2. AssignDeliveryToCourier

Responsabilidades:

- verificar se pedido pode ser entregue
- verificar se entregador esta elegivel
- vincular pedido ao entregador

### 3. RegisterCourierLocation

Responsabilidades:

- registrar posicao recebida do app
- manter ultima localizacao disponivel

### 4. ReorderCourierDeliveries

Responsabilidades:

- buscar fila ativa
- aplicar regra de roteirizacao
- retornar ou persistir nova ordem

### 5. CompleteDelivery

Responsabilidades:

- marcar entrega como concluida
- marcar pedido como entregue
- remover item da fila ativa

### 6. FailDelivery

Responsabilidades:

- registrar motivo
- permitir nova tentativa ou retorno

---

## Regras De Validacao Importantes

### Pedido

- comprador deve ter nome e telefone
- endereco deve ter `CEP`, rua, numero, cidade e estado
- pedido cancelado nao pode ser atribuido

### Entrega

- pedido precisa estar `ready_for_delivery`
- entrega concluida nao pode ser reclassificada
- entrega falha precisa de motivo

### Localizacao

- latitude deve estar entre `-90` e `90`
- longitude deve estar entre `-180` e `180`
- pode haver politica de aceitacao de localizacao antiga

### Reclassificacao

- entregador precisa existir
- deve haver localizacao valida recente
- deve haver entregas pendentes atribuidas

---

## Estrategia De Implementacao Em Fases

### Fase 1 - Fundacao

Objetivo:

- bootstrap
- config
- banco
- migrations
- CRUD basico de comprador, pedido e entregador

Entrega:

- projeto sobe
- dados persistem
- contratos HTTP definidos

### Fase 2 - Operacao De Entrega

Objetivo:

- atribuicao de pedidos
- fila do entregador
- atualizacao de status da entrega

Entrega:

- entregador recebe pedidos
- consegue iniciar e concluir entrega

### Fase 3 - Localizacao

Objetivo:

- endpoint de localizacao
- armazenamento da ultima posicao
- validacoes de coordenadas

Entrega:

- sistema conhece a posicao atual do entregador

### Fase 4 - Roteirizacao MVP

Objetivo:

- algoritmo `CEP + numero`
- endpoint de reclassificacao sob demanda
- persistencia do resultado

Entrega:

- entregador pede reorganizacao
- sistema devolve nova ordem sugerida

### Fase 5 - Observabilidade E Maturidade

Objetivo:

- logging estruturado
- request id
- metricas
- tracing
- testes de integracao

Entrega:

- sistema com cara profissional e facil de operar

### Fase 6 - Evolucao Inteligente

Objetivo:

- usar coordenadas do endereco
- melhorar score de proximidade
- considerar tempo de deslocamento

Entrega:

- roteirizacao bem mais realista

---

## Observabilidade Recomendada

Como voce ja tem esse aprendizado no projeto atual, vale replicar desde cedo:

- `slog` para logs estruturados
- `request_id`
- `trace_id`
- access log HTTP
- metricas Prometheus
- tracing com OpenTelemetry

Eventos importantes para log:

- pedido criado
- pedido atribuido
- localizacao recebida
- reclassificacao solicitada
- nova sequencia gerada
- entrega concluida
- falha de entrega

Metricas uteis:

- total de pedidos criados
- total de pedidos prontos para entrega
- total de reclassificacoes
- latencia do endpoint de roteirizacao
- tempo medio de classificacao
- quantidade de entregas por entregador

---

## Estrategia De Testes

### Domain

Testar:

- transicao de status
- regras de atribuicao
- validacao de localizacao
- parse do numero do endereco
- ordenacao por `CEP + numero`

### Use Cases

Testar:

- criacao de pedido
- atribuicao de pedido
- registro de localizacao
- reclassificacao
- conclusao de entrega

### HTTP Integration

Testar:

- `POST /orders`
- `POST /deliveries/assign`
- `POST /couriers/:courier_id/location`
- `POST /couriers/:courier_id/deliveries/reorder`
- `PATCH /deliveries/:delivery_id/complete`

### Testes De Algoritmo

Esse projeto pede bastante teste de mesa do algoritmo.

Exemplos:

- mesmo `CEP`, numeros crescentes
- mesmo `CEP`, numeros mistos
- `CEP`s diferentes
- endereco sem numero valido
- entrega sem coordenada
- entregador sem localizacao recente

---

## Riscos Tecnicos

### 1. Confundir Aproximacao Com Rota Otima

O sistema inicial nao calcula a melhor rota do mundo real. Ele gera uma classificacao operacional aproximada.

### 2. Enderecos Sujos

Campos como `numero`, `CEP` e complemento podem vir mal preenchidos e prejudicar a ordenacao.

### 3. Localizacao Defasada

Se a posicao do entregador estiver antiga, a reclassificacao perde valor.

### 4. Crescimento Da Complexidade

Se o algoritmo ficar embutido no handler ou repository, a manutencao vai degradar rapido.

### 5. Acoplamento Com API Externa Muito Cedo

Usar Google Maps ou similar cedo demais pode aumentar custo e complexidade antes da hora.

---

## Recomendacao De Escopo Para MVP

Se o objetivo for construir um projeto de portifolio forte e realista, eu recomendo este recorte:

1. cadastro de comprador
2. criacao de pedido com endereco
3. cadastro de entregador
4. atribuicao de pedido ao entregador
5. registro de localizacao atual
6. reclassificacao de entregas por `CEP + numero`
7. conclusao de entrega
8. logs estruturados e metricas basicas

Isso ja produz um projeto muito interessante para mercado porque demonstra:

- arquitetura
- dominio
- geolocalizacao
- algoritmo
- operacao de campo
- observabilidade

---

## Nome Conceitual Do Projeto

Alguns nomes possiveis:

- `go-delivery-routing`
- `go-last-mile`
- `go-courier-ops`
- `go-route-dispatch`
- `go-smart-delivery`

Se a ideia for portifolio, `go-last-mile` e `go-delivery-routing` comunicam muito bem a proposta.

---

## Estrutura Recomendada De Inicio

```text
cmd/
  main.go
internal/
  apperror/
  config/
  dto/
  entity/
  handler/
    buyer/
    order/
    courier/
    delivery/
    routing/
  httpresponse/
  middleware/
  observability/
  repository/
    buyer/
    address/
    order/
    courier/
    delivery/
    location/
  service/
    buyer/
    order/
    courier/
    delivery/
    routing/
pkg/
db/
  migrations/
test/
  integration/
  unit/
docs/
```

---

## Conclusao

Esse projeto tem excelente potencial para estudo e portifolio porque vai alem de um CRUD tradicional.

Ele combina:

- modelagem de pedidos
- operacao logistica
- geolocalizacao
- ordenacao de fila
- regras de negocio reais
- observabilidade moderna

Minha recomendacao tecnica e:

1. replicar a arquitetura em camadas do projeto atual
2. separar bem `pedido`, `entrega`, `localizacao` e `roteirizacao`
3. comecar com a heuristica `CEP + numero`
4. projetar o dominio para suportar coordenadas e algoritmos melhores depois
5. implementar em fases, evitando complexidade prematura

Se voce seguir essa linha, tera um projeto em Go com valor real de aprendizado e muito bom apelo para demonstrar capacidade de engenharia no mercado.

---

## Proximo Passo Recomendado

Depois deste estudo, o proximo documento ideal seria:

- definicao dos `use cases`
- contratos HTTP
- modelagem das entidades em Go
- desenho inicial das migrations

Esse seria o passo natural para transformar o estudo em implementacao.
