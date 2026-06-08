# Ordem Cronologica Da Construcao De Uma API

Este documento organiza a apostila em uma sequencia cronologica de construcao.

A ideia e responder uma pergunta simples:

qual arquivo deve ser criado primeiro, qual vem depois e quando cada conceito entra no projeto.

---

## Visao Geral

A ordem recomendada para construir a API do zero ate um nivel avancado e:

1. definir a estrutura do projeto
2. configurar ambiente e leitura de configuracao
3. padronizar erros da aplicacao
4. padronizar respostas HTTP
5. montar conexao com banco
6. criar health check
7. montar o router
8. criar o `main.go`
9. criar migrations
10. implementar o primeiro CRUD
11. expandir para outros CRUDs
12. adicionar testes automatizados
13. adicionar autenticacao e autorizacao
14. adicionar middlewares compartilhados
15. adicionar logger estruturado
16. adicionar transacoes
17. adicionar paginacao
18. documentar a API com OpenAPI
19. fortalecer testes de integracao com banco real
20. evoluir observabilidade e metricas
21. fortalecer recovery global
22. empacotar e preparar ambientes com Docker
23. automatizar qualidade e entrega com CI/CD

---

## Etapa 1: Estrutura Inicial

Objetivo: preparar o projeto para crescer sem virar um codigo desorganizado.

Arquivos e pastas que devem existir cedo:

- `cmd/`
- `internal/`
- `pkg/`
- `db/migrations/`
- `test/unit/`
- `test/integration/`
- `docs/Criacao do Crud/`

Dentro de `docs/Criacao do Crud/`, a organizacao recomendada agora fica em 3 blocos:

- `01-fundacao/`
- `02-cruds/`
- `03-avancado/`

Essa etapa define o terreno da aplicacao.

---

## Etapa 2: Configuracao

Arquivo principal:

- `internal/config/config.go`

Por que vem cedo:

- tudo depende de ambiente
- servidor depende de porta
- banco depende de credenciais
- JWT depende de segredo

Sem `config.go`, o bootstrap fica espalhado e fragil.

Documento relacionado:

- `08-criando-config-go.md`

---

## Etapa 3: Erros Padronizados

Arquivo principal:

- `internal/apperror/error.go`

Por que vem antes dos CRUDs:

- o `service` precisa devolver erros consistentes
- o `handler` precisa responder com semantica previsivel

Documento relacionado:

- `09-criando-error-go.md`

---

## Etapa 4: Respostas HTTP Compartilhadas

Arquivo principal:

- `internal/httpresponse/response.go`

Por que vem aqui:

- os handlers precisam de bind
- os handlers precisam de validacao
- os handlers precisam padronizar resposta e erro

Documento relacionado:

- `10-criando-response-go.md`

---

## Etapa 5: Banco E Infra De Conexao

Arquivos principais:

- `pkg/internalsql/mysql.go`
- `db/migrations/...`

Por que entra agora:

- sem banco, os repositories nao existem de verdade
- sem migrations, os CRUDs ficam sem schema

Aqui voce prepara:

- conexao com MySQL
- criacao de tabelas
- consistencia do schema

---

## Etapa 6: Health Check

Arquivo principal:

- `internal/handler/health/handler.go`

Por que vem antes do primeiro CRUD:

- ele valida o bootstrap
- ele prova que servidor e banco estao funcionando
- ele cria o primeiro endpoint da aplicacao

Documento relacionado:

- `11-criando-router-go-e-health-check.md`

---

## Etapa 7: Router

Arquivo principal:

- `internal/server/router.go`

Por que vem antes do `main.go` final:

- o `main.go` deve apenas orquestrar
- o `router.go` concentra o registro dos modulos HTTP

Documento relacionado:

- `11-criando-router-go-e-health-check.md`

---

## Etapa 8: Main

Arquivo principal:

- `cmd/main.go`

Por que ele nao deve ser o primeiro arquivo:

- sozinho ele nao resolve nada
- ele depende de `config`, `router` e banco

Documento relacionado:

- `12-criando-main-go.md`

---

## Etapa 9: Primeiro CRUD

Ordem recomendada:

1. `buyers`
2. `couriers`
3. `orders`
4. `deliveries`
5. `courier_locations`
6. `delivery_reorder`

Por que `buyers` vem primeiro:

- tem dominio mais simples
- ajuda a fixar `dto`, `model`, `repository`, `service` e `handler`
- reduz a complexidade de relacoes no inicio

Documentos relacionados:

- `01-crud-buyers.md`
- `02-crud-couriers.md`
- `03-crud-orders.md`
- `04-crud-deliveries.md`
- `05-crud-courier-locations.md`
- `06-delivery-reorder.md`

---

## Etapa 10: Testes

Quando entrar:

- junto do primeiro CRUD
- nunca so no final do projeto

Documento relacionado:

- `07-guia-unico-testes-fase1.md`

Regra pratica:

- terminou um modulo importante, escreva os testes dele

---

## Etapa 11: Autenticacao E Autorizacao

Quando entrar:

- depois que os endpoints publicos basicos estiverem prontos
- antes de abrir rotas sensiveis

Arquivos que costumam aparecer aqui:

- `internal/dto/auth/...`
- `internal/service/auth/...`
- `internal/handler/auth/...`
- `internal/middleware/auth_middleware.go`
- `internal/jwt/...`

---

## Etapa 12: Middlewares Compartilhados

Quando entrar:

- assim que voce perceber repeticao transversal

Exemplos:

- autenticacao
- correlacao de request
- logging
- recovery customizado
- CORS

---

## Etapa 13: Logger Estruturado

Quando entrar:

- quando o projeto parar de ser apenas local
- antes de crescer em modulos e ambientes

Valor pratico:

- facilita debug
- melhora rastreabilidade
- prepara observabilidade

Material complementar recomendado:

- `17-logger-estruturado.md`
- `26-padrao-oficial-de-logs-do-projeto.md`

---

## Etapa 14: Transacoes

Quando entrar:

- quando um caso de uso altera mais de uma tabela
- quando o fluxo precisa de consistencia atomica

Exemplo tipico:

- criar entrega e atualizar pedido no mesmo fluxo

---

## Etapa 15: Paginacao

Quando entrar:

- quando listagens deixarem de ser pequenas
- quando o front precisar de navegacao parcial

---

## Etapa 16: OpenAPI

Quando entrar:

- depois de uma base minima de endpoints estabilizar
- quando a API precisar ficar autoexplicativa para terceiros

---

## Etapa 17: Testes De Integracao Com Banco Real

Quando entrar:

- depois de unitarios basicos
- quando quiser validar repository e fluxos reais

---

## Etapa 18: Observabilidade E Metricas

Quando entrar:

- quando a API ja precisa ser acompanhada em tempo real
- quando logs sozinhos deixam de ser suficientes

---

## Etapa 19: Recovery Global

Quando entrar:

- quando a aplicacao ja possui varios handlers e middlewares
- quando voce quer evitar respostas caoticas em falhas inesperadas

---

## Etapa 20: Docker E Ambientes

Quando entrar:

- quando a API ja sobe localmente com consistencia
- quando voce precisa empacotar e repetir execucao fora da sua maquina

---

## Etapa 21: CI CD

Quando entrar:

- quando o projeto ja compila e testa de forma reproduzivel
- quando voce quer automatizar qualidade e entrega

---

## Linha Mestra Da Apostila

Se voce quiser seguir uma trilha didatica linear, a ordem ideal dos documentos fica assim:

1. `08-criando-config-go.md`
2. `09-criando-error-go.md`
3. `10-criando-response-go.md`
4. `11-criando-router-go-e-health-check.md`
5. `12-criando-main-go.md`
6. `13-fundacao-do-projeto-e-checklist.md`
7. `01-crud-buyers.md`
8. `02-crud-couriers.md`
9. `03-crud-orders.md`
10. `04-crud-deliveries.md`
11. `05-crud-courier-locations.md`
12. `06-delivery-reorder.md`
13. `07-guia-unico-testes-fase1.md`
14. `15-autenticacao-jwt.md`
15. `16-middlewares-compartilhados.md`
16. `17-logger-estruturado.md`
17. `26-padrao-oficial-de-logs-do-projeto.md`
18. `18-transacoes-no-service.md`
19. `19-paginacao-em-listagens.md`
20. `20-openapi-e-swagger.md`
21. `21-testes-de-integracao-com-banco-real.md`
22. `22-observabilidade-e-metricas.md`
23. `23-recovery-customizado-e-tratamento-global.md`
24. `24-deploy-com-docker-e-ambientes.md`
25. `25-ci-cd-e-pipeline-de-testes.md`

---

## Fechamento

A melhor sequencia nao e apenas tecnica.

Ela precisa respeitar a curva de aprendizagem.

Por isso a apostila deve sair:

- do bootstrap
- para o primeiro CRUD
- do CRUD para testes
- dos testes para seguranca
- da seguranca para robustez operacional
- da robustez para operacao e entrega continua
