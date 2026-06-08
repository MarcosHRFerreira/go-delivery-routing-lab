# Roteiro Do CRUD Da Fase 1

## Objetivo

Este arquivo serve como guia para voce implementar manualmente o CRUD da fase 1 do projeto.

Escopo da fase 1:

- bootstrap da aplicacao
- configuracao
- banco de dados
- migrations
- CRUD basico de comprador
- CRUD basico de pedido
- CRUD basico de entregador

Importante:

- a ideia aqui e orientar onde cada parte deve ser criada
- eu nao vou colocar o codigo pronto da aplicacao neste arquivo
- voce pode seguir este roteiro e digitar tudo para fixar o aprendizado

---

## Pasta Gerada Para O Estudo

Este roteiro foi salvo em:

`docs/Criacao do Crud/02-cruds/roteiro-crud-fase1.md`

---

## Estrutura Recomendada Para A Fase 1

Use esta estrutura como base para comecar:

```text
cmd/
  main.go
internal/
  apperror/
  config/
  dto/
    buyer/
    order/
    courier/
  entity/
  handler/
    buyer/
    order/
    courier/
  httpresponse/
  repository/
    buyer/
    address/
    order/
    courier/
  service/
    buyer/
    order/
    courier/
pkg/
  internalsql/
db/
  migrations/
test/
  integration/
  unit/
docs/
  Criacao do Crud/
    01-fundacao/
    02-cruds/
    03-avancado/
```

---

## Antes De Comecar

Antes de implementar o CRUD, vale revisar as migrations que ja existem no projeto.

Pontos que merecem ajuste antes de seguir:

1. a tabela de compradores foi criada como `byers`, mas no estudo o nome esperado e `buyers`
2. a foreign key de `orders` referencia `byers(id)`, o que propaga o typo
3. a tabela `couriers` usa a coluna `vehicle`, mas no estudo original aparece `vehicle_type`
4. a tabela `addresses` atual exige `latitude` e `longitude` como `NOT NULL`, mas o DTO inicial do estudo nao exige isso
5. os blocos `-- migrate:down` ainda nao foram preenchidos

Recomendacao:

- alinhe primeiro as migrations com o documento principal
- depois implemente o CRUD em cima do schema final

---

## Ordem Recomendada De Implementacao

Siga esta ordem para aprender com menos friccao:

1. corrigir e validar as migrations
2. criar as entidades
3. criar os DTOs de entrada e saida
4. criar as interfaces de repository
5. criar os repositories MySQL
6. criar os services ou use cases
7. criar os handlers HTTP
8. registrar as rotas no `cmd/main.go`
9. criar os testes unitarios
10. criar os testes de integracao HTTP

---

## Entidades Da Fase 1

Voce deve criar estas entidades dentro de `internal/entity/`:

### Comprador

Arquivo sugerido:

- `internal/entity/buyer.go`

Campos esperados:

- `ID`
- `Name`
- `Document`
- `Phone`
- `Email`
- `CreatedAt`
- `UpdatedAt`

### Endereco

Arquivo sugerido:

- `internal/entity/address.go`

Campos esperados:

- `ID`
- `ZipCode`
- `Street`
- `Number`
- `Complement`
- `District`
- `City`
- `State`
- `Latitude`
- `Longitude`
- `CreatedAt`
- `UpdatedAt`

### Pedido

Arquivo sugerido:

- `internal/entity/order.go`

Campos esperados:

- `ID`
- `Code`
- `BuyerID`
- `DeliveryAddressID`
- `Status`
- `TotalAmount`
- `CreatedAt`
- `UpdatedAt`

### Entregador

Arquivo sugerido:

- `internal/entity/courier.go`

Campos esperados:

- `ID`
- `Name`
- `Phone`
- `VehicleType`
- `Status`
- `CreatedAt`
- `UpdatedAt`

---

## DTOs Da Fase 1

Crie os DTOs dentro de `internal/dto/` separados por modulo.

### Buyer

Pasta:

- `internal/dto/buyer/`

Arquivos sugeridos:

- `create_buyer_input.go`
- `update_buyer_input.go`
- `buyer_output.go`

### Order

Pasta:

- `internal/dto/order/`

Arquivos sugeridos:

- `create_order_input.go`
- `update_order_input.go`
- `order_output.go`

Observacao:

- como pedido depende de endereco, voce pode criar um DTO aninhado para endereco de entrega

Arquivos extras opcionais:

- `delivery_address_input.go`

### Courier

Pasta:

- `internal/dto/courier/`

Arquivos sugeridos:

- `create_courier_input.go`
- `update_courier_input.go`
- `update_courier_status_input.go`
- `courier_output.go`

---

## Repositories

Os repositories devem ser divididos por agregado.

### Buyer Repository

Pasta:

- `internal/repository/buyer/`

Arquivos sugeridos:

- `repository.go`
- `mysql_repository.go`

Responsabilidades:

- criar comprador
- buscar comprador por id
- listar compradores
- atualizar comprador
- remover comprador

### Address Repository

Pasta:

- `internal/repository/address/`

Arquivos sugeridos:

- `repository.go`
- `mysql_repository.go`

Responsabilidades:

- criar endereco
- buscar endereco por id
- atualizar endereco
- remover endereco quando fizer sentido

### Order Repository

Pasta:

- `internal/repository/order/`

Arquivos sugeridos:

- `repository.go`
- `mysql_repository.go`

Responsabilidades:

- criar pedido
- buscar pedido por id
- listar pedidos
- atualizar pedido
- cancelar pedido

### Courier Repository

Pasta:

- `internal/repository/courier/`

Arquivos sugeridos:

- `repository.go`
- `mysql_repository.go`

Responsabilidades:

- criar entregador
- buscar entregador por id
- listar entregadores
- atualizar entregador
- atualizar status
- remover entregador

---

## Services Ou Use Cases

Aqui voce coloca a regra de negocio da fase 1.

### Buyer Service

Pasta:

- `internal/service/buyer/`

Arquivos sugeridos:

- `create_buyer.go`
- `get_buyer_by_id.go`
- `list_buyers.go`
- `update_buyer.go`
- `delete_buyer.go`

Regras minimas:

- validar nome obrigatorio
- validar documento obrigatorio
- validar telefone obrigatorio

### Order Service

Pasta:

- `internal/service/order/`

Arquivos sugeridos:

- `create_order.go`
- `get_order_by_id.go`
- `list_orders.go`
- `update_order.go`
- `cancel_order.go`
- `mark_order_ready_for_delivery.go`

Regras minimas:

- validar se comprador existe
- validar se endereco foi informado
- gerar `code` do pedido
- iniciar pedido com status `created`
- permitir mudanca para `ready_for_delivery`
- permitir cancelamento conforme a regra que voce definir

### Courier Service

Pasta:

- `internal/service/courier/`

Arquivos sugeridos:

- `create_courier.go`
- `get_courier_by_id.go`
- `list_couriers.go`
- `update_courier.go`
- `update_courier_status.go`
- `delete_courier.go`

Regras minimas:

- validar nome obrigatorio
- validar telefone obrigatorio
- validar tipo de veiculo
- validar status inicial

---

## Handlers HTTP

Os handlers recebem a request, chamam o service e montam a response.

### Buyer Handler

Pasta:

- `internal/handler/buyer/`

Arquivos sugeridos:

- `create_buyer_handler.go`
- `get_buyer_by_id_handler.go`
- `list_buyers_handler.go`
- `update_buyer_handler.go`
- `delete_buyer_handler.go`
- `routes.go`

### Order Handler

Pasta:

- `internal/handler/order/`

Arquivos sugeridos:

- `create_order_handler.go`
- `get_order_by_id_handler.go`
- `list_orders_handler.go`
- `update_order_handler.go`
- `cancel_order_handler.go`
- `mark_order_ready_for_delivery_handler.go`
- `routes.go`

### Courier Handler

Pasta:

- `internal/handler/courier/`

Arquivos sugeridos:

- `create_courier_handler.go`
- `get_courier_by_id_handler.go`
- `list_couriers_handler.go`
- `update_courier_handler.go`
- `update_courier_status_handler.go`
- `delete_courier_handler.go`
- `routes.go`

---

## Rotas HTTP Da Fase 1

Estas sao as rotas que fazem mais sentido para o CRUD inicial.

### Buyers

- `POST /buyers`
- `GET /buyers/:buyer_id`
- `GET /buyers`
- `PUT /buyers/:buyer_id`
- `DELETE /buyers/:buyer_id`

### Orders

- `POST /orders`
- `GET /orders/:order_id`
- `GET /orders`
- `PUT /orders/:order_id`
- `PATCH /orders/:order_id/ready-for-delivery`
- `PATCH /orders/:order_id/cancel`

### Couriers

- `POST /couriers`
- `GET /couriers/:courier_id`
- `GET /couriers`
- `PUT /couriers/:courier_id`
- `PATCH /couriers/:courier_id/status`
- `DELETE /couriers/:courier_id`

Observacao:

- se preferir, voce pode trocar `PUT` por `PATCH` nos endpoints de atualizacao
- para pedidos, delete fisico normalmente nao e a melhor opcao; o `cancel` cobre melhor a realidade de negocio

---

## O Que Implementar Em Cada CRUD

## 1. CRUD De Comprador

### Objetivo

Permitir cadastrar, consultar, listar, atualizar e remover compradores.

### Pastas Envolvidas

- `internal/entity/`
- `internal/dto/buyer/`
- `internal/repository/buyer/`
- `internal/service/buyer/`
- `internal/handler/buyer/`
- `test/unit/`
- `test/integration/`

### Ordem De Trabalho

1. criar `internal/entity/buyer.go`
2. criar os DTOs em `internal/dto/buyer/`
3. criar a interface em `internal/repository/buyer/repository.go`
4. criar a implementacao MySQL em `internal/repository/buyer/mysql_repository.go`
5. criar os services em `internal/service/buyer/`
6. criar os handlers em `internal/handler/buyer/`
7. registrar as rotas no `routes.go` do modulo
8. plugar as rotas no `cmd/main.go`
9. criar testes unitarios dos services
10. criar testes de integracao dos endpoints

### Checklist

- criar comprador
- buscar por id
- listar todos
- atualizar dados
- remover comprador
- retornar `404` quando nao existir
- retornar `400` para payload invalido

---

## 2. CRUD De Pedido

### Objetivo

Permitir criar, consultar, listar e atualizar o estado do pedido.

### Pastas Envolvidas

- `internal/entity/`
- `internal/dto/order/`
- `internal/repository/order/`
- `internal/repository/address/`
- `internal/service/order/`
- `internal/handler/order/`
- `test/unit/`
- `test/integration/`

### Ordem De Trabalho

1. criar `internal/entity/order.go`
2. criar `internal/entity/address.go`
3. criar os DTOs em `internal/dto/order/`
4. criar interface e implementacao de endereco
5. criar interface e implementacao de pedido
6. criar service de criacao do pedido com persistencia de endereco
7. criar service para consulta por id
8. criar service para listagem
9. criar service para update
10. criar service para `ready_for_delivery`
11. criar service para cancelamento
12. criar handlers e rotas
13. registrar tudo no `cmd/main.go`
14. criar testes unitarios e de integracao

### Checklist

- validar comprador existente antes de criar pedido
- criar endereco de entrega
- criar pedido com status inicial
- buscar pedido por id
- listar pedidos
- atualizar pedido
- marcar como `ready_for_delivery`
- cancelar pedido
- retornar `404` para comprador ou pedido inexistente
- retornar `400` para payload invalido

### Dica

No CRUD de pedido, o endereco pode ser tratado como parte do processo de criacao do pedido, mesmo tendo repository separado.

---

## 3. CRUD De Entregador

### Objetivo

Permitir cadastrar, consultar, listar, atualizar e alterar o status do entregador.

### Pastas Envolvidas

- `internal/entity/`
- `internal/dto/courier/`
- `internal/repository/courier/`
- `internal/service/courier/`
- `internal/handler/courier/`
- `test/unit/`
- `test/integration/`

### Ordem De Trabalho

1. criar `internal/entity/courier.go`
2. criar os DTOs em `internal/dto/courier/`
3. criar a interface em `internal/repository/courier/repository.go`
4. criar a implementacao MySQL em `internal/repository/courier/mysql_repository.go`
5. criar os services em `internal/service/courier/`
6. criar handlers e rotas
7. registrar no `cmd/main.go`
8. criar testes unitarios
9. criar testes de integracao

### Checklist

- criar entregador
- buscar por id
- listar entregadores
- atualizar dados
- atualizar status
- remover entregador
- retornar `404` quando nao existir
- retornar `400` para payload invalido

---

## Arquivos Transversais

Existem alguns arquivos que nao pertencem a um unico CRUD, mas vao ser usados por todos.

### Configuracao

Ja existe:

- `internal/config/config.go`

Pode evoluir para:

- leitura de ambiente
- porta HTTP
- dados do banco
- segredo JWT

### Conexao Com Banco

Ja existe:

- `pkg/internalsql/mysql.go`

Use esse arquivo para:

- abrir conexao
- validar `ping`
- configurar pool

### Main

Arquivo:

- `cmd/main.go`

Nesse arquivo voce deve:

- carregar configuracao
- abrir conexao com MySQL
- instanciar repositories
- instanciar services
- instanciar handlers
- registrar rotas
- iniciar servidor HTTP

### Tratamento De Erros

Crie:

- `internal/apperror/`

Arquivos sugeridos:

- `app_error.go`
- `not_found_error.go`
- `validation_error.go`
- `conflict_error.go`

### Padronizacao De Resposta

Crie:

- `internal/httpresponse/`

Arquivos sugeridos:

- `success.go`
- `error.go`

Use para:

- retornar JSON padrao
- centralizar status code
- evitar duplicacao nos handlers

---

## Testes Recomendados

## Unitarios

Pasta:

- `test/unit/`

Arquivos sugeridos:

- `test/unit/buyer/create_buyer.test.go`
- `test/unit/buyer/update_buyer.test.go`
- `test/unit/order/create_order.test.go`
- `test/unit/order/mark_order_ready_for_delivery.test.go`
- `test/unit/courier/create_courier.test.go`
- `test/unit/courier/update_courier_status.test.go`

O que testar:

- regras de validacao
- transicoes de status
- cenarios felizes
- cenarios alternativos com erro

## Integracao

Pasta:

- `test/integration/`

Arquivos sugeridos:

- `test/integration/buyer_http.test.go`
- `test/integration/order_http.test.go`
- `test/integration/courier_http.test.go`

O que testar:

- `POST`
- `GET`
- `PUT` ou `PATCH`
- `DELETE` ou `cancel`
- status code
- mensagem de erro

---

## Sequencia De Estudo Recomendada

Se voce quiser fixar melhor, implemente nesta sequencia:

1. corrigir migrations
2. fazer CRUD completo de `buyers`
3. fazer CRUD completo de `couriers`
4. fazer CRUD de `orders`
5. adicionar `ready_for_delivery`
6. adicionar `cancel`
7. criar testes unitarios
8. criar testes de integracao

Motivo:

- `buyers` e mais simples
- `couriers` ainda e simples, mas ja introduz `status`
- `orders` depende de `buyers` e `addresses`, entao fica mais facil fazer depois

---

## Checklist Final Da Fase 1

Ao final da fase 1, voce deve ter:

- projeto subindo localmente
- banco conectado com sucesso
- migrations funcionando
- CRUD de compradores funcionando
- CRUD de entregadores funcionando
- CRUD de pedidos funcionando
- rotas registradas no servidor
- validacoes basicas implementadas
- erros HTTP padronizados
- testes unitarios minimos
- testes de integracao minimos

---

## Proximo Passo Depois Da Fase 1

Quando terminar este roteiro, a fase natural seguinte e:

- atribuicao de pedido ao entregador
- criacao da fila de entregas
- atualizacao do status da entrega
- registro da localizacao do entregador

Esses pontos pertencem mais a fase 2 e fase 3 do estudo.
