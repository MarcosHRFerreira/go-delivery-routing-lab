# Logger Estruturado

Este capitulo explica quando sair do `log.Printf(...)` simples para um logger mais maduro.

---

## Quando Criar

A ordem recomendada e:

1. bootstrap
2. CRUDs
3. testes basicos
4. middlewares
5. logger estruturado

---

## Por Que Isso Importa

Com o crescimento da API, o logger passa a ser util para:

- rastrear falhas
- entender fluxo de requests
- cruzar logs por request id
- diferenciar ambiente local e producao

---

## O Que Um Logger Estruturado Deve Entregar

- nivel de log
- mensagem clara
- campos nomeados
- integracao com request id
- possibilidade de trocar implementacao no futuro

---

## Onde Normalmente Fica

Sugestao:

- `internal/logger/`
- `internal/middleware/logging_middleware.go`

---

## Ordem Interna De Evolucao

1. definir interface simples de logger
2. escolher implementacao
3. injetar logger no bootstrap
4. criar middleware HTTP de logging
5. passar logger para services importantes quando fizer sentido

---

## O Que Logar

Priorize:

- inicio e fim de request
- status code
- tempo de resposta
- erros internos
- falhas de infraestrutura

Evite:

- senha
- token completo
- dados sensiveis em texto puro

---

## Distribuicao De Logs Por Camada

Quando voce for incrementar logs no projeto, uma divisao saudavel costuma ser:

- `main.go`: eventos de ciclo de vida da aplicacao
- `router.go`: registro de middlewares e composicao da infraestrutura
- `middleware`: inicio e fim de request, status e latencia
- `handler`: erro de contexto ou falha relevante na borda HTTP
- `service`: regra de negocio importante, conflito e decisao critica
- `repository`: erro inesperado de banco ou falha de infraestrutura

Essa separacao evita dois extremos ruins:

- log demais em todos os pontos
- falta de contexto quando um erro acontecer

Boa regra pratica:

- sucesso tecnico repetitivo pode ficar so no middleware
- erro de negocio importante pode aparecer no `service`
- erro de banco merece log no `repository`

---

## Campos Uteis Nos Logs

No comeco, os campos mais valiosos costumam ser:

- `request_id`
- `method`
- `path`
- `status_code`
- `duration_ms`
- `buyer_id`, `order_id` ou outro identificador de negocio quando existir

Evite colocar:

- email completo quando nao for necessario
- documento pessoal em texto puro
- telefone completo sem necessidade

---

## Relacao Com Outros Capitulos

Esse capitulo conversa bem com:

- `16-middlewares-compartilhados.md`
- `21-testes-de-integracao-com-banco-real.md`
- `26-padrao-oficial-de-logs-do-projeto.md`

---

## Proximo Capitulo

Depois de logger, um aprofundamento natural e:

- `26-padrao-oficial-de-logs-do-projeto.md`

Depois disso, um proximo tema muito forte e:

- `18-transacoes-no-service.md`
