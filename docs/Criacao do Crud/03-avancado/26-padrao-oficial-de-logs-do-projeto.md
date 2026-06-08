# Padrao Oficial De Logs Do Projeto

Este capitulo consolida uma convencao unica para logging no projeto.

A ideia nao e apenas "ter logs".

O objetivo e garantir que os logs:

- tenham formato consistente
- usem campos previsiveis
- facilitem suporte e debug
- nao exponham dados sensiveis sem necessidade

---

## Objetivo Deste Capitulo

Depois de estudar `17-logger-estruturado.md`, este documento serve para responder perguntas praticas:

- o que cada camada deve logar
- como nomear mensagens
- quais campos devem aparecer sempre
- o que nunca deve ir para log
- como manter rastreabilidade sem poluicao

---

## Principios Do Padrao

O padrao recomendado para o projeto e:

- log com mensagem curta e objetiva
- campos nomeados para contexto
- mesma linguagem em todo o projeto
- prioridade para identificadores tecnicos e de negocio
- cuidado com dados pessoais e segredos

Boa regra pratica:

- mensagem explica o evento
- campos explicam o contexto

---

## Formato Recomendado

Mesmo que a implementacao comece simples, pense no log neste formato mental:

- `level`
- `message`
- `request_id`
- `module`
- `operation`
- campos extras de contexto

Exemplo conceitual:

```text
level=INFO message="buyer created successfully" module=buyer operation=create_buyer buyer_id=10 request_id=abc123
```

Outro exemplo:

```text
level=ERROR message="failed to create buyer" module=buyer operation=create_buyer request_id=abc123 error="duplicate buyer document"
```

---

## Niveis De Log

Use uma convencao simples no inicio:

- `DEBUG`: detalhes tecnicos que ajudam desenvolvimento local
- `INFO`: eventos importantes de fluxo normal
- `WARN`: situacoes anormais que nao derrubam o fluxo
- `ERROR`: falhas que impedem a operacao esperada

Uma distribuicao saudavel costuma ser:

- sucesso importante de negocio em `INFO`
- conflito ou tentativa invalida em `WARN`
- falha de banco, dependencia externa ou erro inesperado em `ERROR`

Se quiser manter o projeto enxuto no inicio, voce pode até comecar apenas com:

- `INFO`
- `ERROR`

---

## Padrao De Mensagens

As mensagens devem ser:

- curtas
- objetivas
- orientadas a evento
- escritas sempre no mesmo estilo

Prefira mensagens como:

- `buyer created successfully`
- `buyer already exists by document`
- `failed to fetch buyer by id`
- `http request completed`
- `database connection closed`

Evite mensagens como:

- `deu erro`
- `erro aqui`
- `entrou no if`
- `chegou ate aqui`

O log deve explicar o que aconteceu sem depender de leitura do codigo.

---

## Campos Padrao Recomendados

Quando disponiveis, estes campos ajudam bastante:

- `request_id`
- `module`
- `operation`
- `status_code`
- `duration_ms`
- `buyer_id`
- `courier_id`
- `order_id`
- `delivery_id`

Campos de infraestrutura tambem costumam ser uteis:

- `host`
- `port`
- `environment`
- `error`

Boa pratica:

- mantenha nomes de campos estaveis
- nao troque `buyer_id` por `id_buyer` em outro modulo

---

## Convencao Por Camada

### `main.go`

Deve logar:

- inicio do bootstrap
- falha ao carregar configuracao
- falha ao conectar no banco
- inicio do servidor
- recebimento de sinal de desligamento
- sucesso ou falha no shutdown

### `router.go`

Deve logar:

- ativacao de middlewares relevantes
- montagem de dependencias compartilhadas quando fizer sentido

Evite excesso aqui.

O `router.go` e mais composicao do que execucao de regra.

### `middleware`

Deve logar:

- inicio e fim de request
- metodo
- path
- status code
- latencia
- request id

Esse e o melhor lugar para log transversal.

### `handler`

Deve logar:

- erro de bind
- erro de validacao quando precisar de contexto adicional
- falha inesperada na borda HTTP

Nao precisa logar toda resposta de sucesso se o middleware ja fizer isso.

### `service`

Deve logar:

- conflito de regra de negocio
- tentativa invalida de transicao
- ausencia importante de dependencia de dominio
- sucesso de operacao relevante quando isso ajuda rastreabilidade

### `repository`

Deve logar:

- falha inesperada de banco
- erro de query
- falha de persistencia

Evite logar cada `SELECT` bem-sucedido sem necessidade.

---

## Convencao De `module` E `operation`

Para deixar os logs pesquisaveis, uma convencao simples funciona bem:

- `module`: nome do dominio ou da infraestrutura
- `operation`: nome do caso de uso ou da acao tecnica

Exemplos:

- `module=buyer operation=create_buyer`
- `module=buyer operation=get_buyer_by_id`
- `module=order operation=mark_ready_for_delivery`
- `module=delivery operation=complete_delivery`
- `module=http operation=request_logging`
- `module=bootstrap operation=start_server`

---

## O Que Nunca Deve Ir Para Log

Evite registrar:

- senha
- token completo
- segredo JWT
- payload completo com dados sensiveis
- documento pessoal completo sem necessidade
- email completo em cenarios que nao exigem isso
- telefone completo sem necessidade

Quando algum dado pessoal for importante para correlacao, prefira:

- mascarar parte do valor
- registrar apenas o identificador interno

---

## Exemplos De Mensagens Por Modulo

### Buyers

```text
level=INFO message="buyer created successfully" module=buyer operation=create_buyer buyer_id=10 request_id=abc123
level=WARN message="buyer already exists by document" module=buyer operation=create_buyer request_id=abc123
level=ERROR message="failed to persist buyer" module=buyer operation=create_buyer request_id=abc123 error="insert failed"
```

### Couriers

```text
level=INFO message="courier created successfully" module=courier operation=create_courier courier_id=3 request_id=def456
level=WARN message="courier already exists by phone" module=courier operation=create_courier request_id=def456
level=WARN message="courier status transition is invalid" module=courier operation=update_courier_status courier_id=3 request_id=def456
level=ERROR message="failed to update courier status" module=courier operation=update_courier_status courier_id=3 request_id=def456 error="update failed"
```

### Orders

```text
level=INFO message="order created successfully" module=order operation=create_order order_id=21 request_id=ghi789
level=WARN message="buyer was not found for order creation" module=order operation=create_order request_id=ghi789
level=WARN message="order status transition is invalid" module=order operation=mark_ready_for_delivery order_id=21 request_id=ghi789
level=ERROR message="failed to update order status" module=order operation=mark_ready_for_delivery order_id=21 request_id=ghi789 error="update failed"
```

### Deliveries

```text
level=INFO message="delivery assigned successfully" module=delivery operation=assign_delivery delivery_id=9 order_id=21 courier_id=3 request_id=jkl012
level=INFO message="delivery completed successfully" module=delivery operation=complete_delivery delivery_id=9 request_id=jkl012
level=WARN message="delivery cannot be started from current status" module=delivery operation=start_delivery delivery_id=9 request_id=jkl012
level=ERROR message="failed to persist delivery status update" module=delivery operation=complete_delivery delivery_id=9 request_id=jkl012 error="update failed"
```

### Courier Locations

```text
level=INFO message="courier location stored successfully" module=location operation=register_courier_location courier_id=3 request_id=mno345
level=WARN message="courier location coordinates are invalid" module=location operation=register_courier_location courier_id=3 request_id=mno345
level=WARN message="courier latest location was not found" module=location operation=get_latest_courier_location courier_id=3 request_id=mno345
level=ERROR message="failed to store courier location" module=location operation=register_courier_location courier_id=3 request_id=mno345 error="insert failed"
```

### Delivery Reorder

```text
level=INFO message="delivery reorder completed successfully" module=routing operation=reorder_deliveries courier_id=3 reordered_count=4 request_id=pqr678
level=INFO message="delivery sequence fetched successfully" module=routing operation=get_delivery_sequence courier_id=3 request_id=pqr678
level=WARN message="courier current location is required for reorder" module=routing operation=reorder_deliveries courier_id=3 request_id=pqr678
level=WARN message="courier has no pending deliveries to reorder" module=routing operation=reorder_deliveries courier_id=3 request_id=pqr678
level=ERROR message="failed to persist delivery reorder history" module=routing operation=reorder_deliveries courier_id=3 request_id=pqr678 error="insert failed"
```

---

## Erros Comuns Ao Criar Logs

Os erros mais comuns costumam ser:

- logar a mesma falha em todas as camadas
- escrever mensagens vagas
- misturar portugues e ingles sem criterio
- mudar o nome dos campos entre modulos
- registrar sucesso irrelevante o tempo inteiro
- expor dado sensivel em texto puro

Boa regra:

- cada camada loga o que e responsabilidade dela

---

## Sugestao De Politica Inicial

Se voce quiser um padrao pratico para comecar, use:

1. middleware loga toda request
2. `main.go` loga ciclo de vida da aplicacao
3. `service` loga conflitos e transicoes importantes
4. `repository` loga apenas falhas inesperadas
5. dados sensiveis entram mascarados ou nao entram

Esse padrao ja entrega:

- rastreabilidade
- clareza
- baixo ruido
- base pronta para observabilidade futura

---

## Relacao Com Outros Capitulos

Este capitulo conversa diretamente com:

- `16-middlewares-compartilhados.md`
- `17-logger-estruturado.md`
- `22-observabilidade-e-metricas.md`
- `23-recovery-customizado-e-tratamento-global.md`

---

## Fechamento

Um projeto com logs consistentes fica mais facil de:

- operar
- depurar
- auditar
- evoluir

Por isso, o melhor caminho nao e sair logando tudo.

O melhor caminho e definir um padrao e aplica-lo com disciplina.
