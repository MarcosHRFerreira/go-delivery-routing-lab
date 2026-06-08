# Observabilidade E Metricas

Este capitulo mostra quando a API precisa parar de ser apenas funcional e passar a ser observavel.

Observabilidade nao e a mesma coisa que logging.

Ela combina sinais diferentes para ajudar a responder:

- o que aconteceu
- quando aconteceu
- com quem aconteceu
- por que aconteceu

---

## Quando Criar

Observabilidade entra melhor quando:

- a API ja tem alguns modulos reais
- o projeto passa a rodar em mais de um ambiente
- erros intermitentes comecam a aparecer
- o tempo de resposta passa a importar mais

---

## Tres Pilares Classicos

Os pilares mais conhecidos sao:

- logs
- metricas
- traces

Para uma apostila inicial, o caminho mais simples e:

1. logger estruturado
2. metricas HTTP
3. correlacao de request
4. traces no futuro

---

## O Que Medir Primeiro

As metricas iniciais mais uteis costumam ser:

- quantidade de requests por rota
- quantidade de erros por rota
- latencia por endpoint
- quantidade de conexoes ou falhas de banco
- health check por ambiente

---

## Onde Isso Costuma Entrar

Arquivos e pastas que geralmente aparecem:

- `internal/middleware/metrics_middleware.go`
- `internal/observability/`
- `internal/logger/`

O `router.go` costuma registrar os middlewares e endpoints de metricas.

---

## Ordem Interna De Evolucao

1. padronizar `request id`
2. consolidar logger estruturado
3. medir tempo de resposta por request
4. expor endpoint de metricas
5. criar dashboards ou painéis no futuro

---

## O Que Nao Fazer

- medir tudo desde o primeiro dia
- expor dados sensiveis nas metricas
- misturar regra de negocio com coleta de metricas
- depender apenas de logs para tudo

---

## Valor Didatico

Esse capitulo ensina que API madura nao e so:

- CRUD
- autenticação
- testes

Ela tambem precisa ser monitoravel.

---

## Relacao Com Outros Capitulos

Esse tema conversa diretamente com:

- `17-logger-estruturado.md`
- `23-recovery-customizado-e-tratamento-global.md`
- `24-deploy-com-docker-e-ambientes.md`

---

## Proximo Capitulo

Depois de observabilidade, um passo muito coerente e:

- `23-recovery-customizado-e-tratamento-global.md`
