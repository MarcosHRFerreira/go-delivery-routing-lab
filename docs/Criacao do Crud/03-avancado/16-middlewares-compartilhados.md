# Middlewares Compartilhados

Este capitulo explica quando criar middlewares e como encaixa-los na ordem de evolucao da API.

Middleware nao deve ser o primeiro recurso do projeto.

Ele entra quando a aplicacao ja comeca a repetir comportamentos transversais.

---

## Quando Criar

A ordem recomendada e:

1. bootstrap
2. CRUDs basicos
3. autenticacao
4. middlewares compartilhados

---

## Sinais De Que Ja Vale Criar Middleware

- varios handlers repetem a mesma validacao transversal
- varias rotas precisam de autenticacao
- varios endpoints precisam de logging padrao
- varios fluxos precisam de correlacao de request

---

## Tipos Mais Comuns

Os middlewares mais naturais em uma API sao:

- autenticacao
- recovery customizado
- logging
- CORS
- request id
- medicao de tempo

---

## Onde Criar

Sugestao de pasta:

- `internal/middleware/`

Arquivos comuns:

- `auth_middleware.go`
- `request_id_middleware.go`
- `logging_middleware.go`
- `cors_middleware.go`
- `recovery_middleware.go`

---

## Ordem Interna De Construcao

1. criar middleware de recovery customizado
2. criar middleware de request id
3. criar middleware de logging
4. criar middleware de autenticacao
5. registrar todos no `router.go`

---

## O Que Nao Colocar Em Middleware

Evite colocar em middleware:

- regra de negocio do dominio
- SQL de casos de uso
- validacao especifica de um CRUD isolado
- logica de atualizacao de entidades

Middleware serve para preocupacoes transversais.

---

## Como Pensar A Ordem De Registro

No `router.go`, a ordem costuma importar.

Um fluxo comum e:

1. request id
2. logger
3. recovery
4. CORS
5. autenticacao em grupos protegidos

---

## Estrategia De Logging Via Middleware

Se voce quer adicionar log sem espalhar chamadas por todos os handlers, middleware e o melhor ponto de partida.

Um middleware de logging costuma registrar:

- metodo HTTP
- rota
- status code
- tempo de resposta
- request id
- erro quando existir

Essa abordagem ajuda porque:

- reduz repeticao nos handlers
- padroniza o formato do log
- permite medir latencia da request inteira

Boa divisao de responsabilidade:

- middleware registra fluxo HTTP transversal
- handler registra apenas erro relevante de entrada ou contexto especifico
- service registra regra de negocio importante ou conflito
- repository registra falha inesperada de banco

Evite no middleware:

- logar body completo com dados sensiveis
- logar token completo
- duplicar o mesmo erro em varias camadas sem necessidade

O middleware deve capturar a visao transversal da request.

---

## O Que O Middleware De Logging Deve Registrar

Quando voce for implementar esse middleware, os campos mais uteis no comeco costumam ser:

- metodo HTTP
- path
- status code
- tempo de resposta
- request id
- ip ou origem da request quando fizer sentido

Em caso de erro, vale incluir tambem:

- mensagem resumida do erro
- camada que falhou, quando isso estiver disponivel

O detalhamento de regra de negocio continua melhor dentro de `service` e `repository`, quando realmente necessario.

## Beneficio Didatico

Esse capitulo ensina que nem tudo pertence ao `handler`.

Ele ajuda a pessoa a entender separacao entre:

- infraestrutura HTTP
- autenticacao
- observabilidade
- dominio

---

## Proximo Capitulo

Depois de middlewares, vale estudar:

- `17-logger-estruturado.md`
