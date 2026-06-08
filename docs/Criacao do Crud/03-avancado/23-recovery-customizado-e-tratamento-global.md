# Recovery Customizado E Tratamento Global

Este capitulo mostra como tratar falhas inesperadas sem deixar a API responder de forma caotica.

Mesmo com `apperror`, ainda existem erros nao previstos:

- `panic`
- nil pointer
- falhas inesperadas de infraestrutura
- erros nao mapeados

---

## Quando Criar

Esse capitulo entra melhor quando:

- a API ja tem varios handlers
- middlewares compartilhados ja existem
- voce quer evitar respostas quebradas em producao

---

## O Papel Do Recovery

Um middleware de recovery customizado serve para:

- capturar `panic`
- registrar log detalhado
- devolver resposta HTTP consistente
- evitar que uma request derrube o processo inteiro

---

## O Papel Do Tratamento Global

O tratamento global organiza a forma como erros sao devolvidos.

Idealmente:

- `service` devolve `apperror`
- `httpresponse` traduz erro conhecido
- `middleware` protege contra erro inesperado

---

## Onde Isso Costuma Entrar

Arquivos comuns:

- `internal/middleware/recovery_middleware.go`
- `internal/httpresponse/response.go`
- `internal/logger/`

O `router.go` costuma registrar esse middleware cedo na cadeia HTTP.

---

## Ordem Interna De Evolucao

1. manter `apperror` para erros esperados
2. manter `httpresponse` para respostas padronizadas
3. criar middleware de recovery customizado
4. registrar logs tecnicos em `panic`
5. devolver `500` padronizado ao cliente

---

## Resposta Recomendada Ao Cliente

Quando acontecer falha inesperada, o cliente nao deve receber stack trace.

O ideal e uma resposta enxuta, por exemplo:

- `message: internal server error`

Os detalhes tecnicos ficam apenas no log.

---

## O Que Nao Fazer

- devolver stack trace para o cliente
- usar `panic` como regra de controle de fluxo
- misturar logica de recovery com regra de negocio
- deixar cada handler tratar `panic` manualmente

---

## Valor Didatico

Esse capitulo mostra a diferenca entre:

- erro previsto de negocio
- falha inesperada de execucao

Essa separacao amadurece bastante a arquitetura da API.

---

## Relacao Com Outros Capitulos

Esse tema conversa com:

- `09-criando-error-go.md`
- `10-criando-response-go.md`
- `16-middlewares-compartilhados.md`
- `22-observabilidade-e-metricas.md`

---

## Proximo Capitulo

Depois de recovery global, um proximo passo natural e:

- `24-deploy-com-docker-e-ambientes.md`
