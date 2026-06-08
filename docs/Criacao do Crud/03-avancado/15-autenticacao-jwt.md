# Autenticacao Com JWT

Este capitulo mostra quando e como adicionar autenticacao JWT na evolucao da API.

A autenticacao nao deve ser o primeiro passo do projeto.

Ela entra melhor depois que o bootstrap e os primeiros CRUDs ja estiverem claros.

---

## Quando Criar

A ordem recomendada e:

1. bootstrap da API
2. primeiro CRUD
3. testes basicos
4. autenticacao JWT

Motivo:

- autenticar antes de entender handlers e services costuma aumentar a confusao
- com a base pronta, JWT vira uma evolucao e nao um obstaculo

---

## Objetivos Do JWT

Com JWT, a API passa a conseguir:

- autenticar usuarios
- emitir token
- proteger rotas privadas
- transportar informacoes basicas do usuario autenticado

---

## Pastas Que Costumam Surgir

Uma organizacao simples e:

- `internal/dto/auth/`
- `internal/service/auth/`
- `internal/handler/auth/`
- `internal/middleware/auth_middleware.go`
- `internal/jwt/token.go`

---

## Ordem Interna De Criacao

Se for implementar JWT do zero, siga esta ordem:

1. adicionar segredo no `config.go`
2. criar pacote `internal/jwt`
3. criar DTOs de login
4. criar service de autenticacao
5. criar handler de login
6. criar middleware de autenticacao
7. proteger rotas no `router.go`
8. criar testes

---

## O Que Vai No `config.go`

O `config.go` ja deve carregar algo como:

- `SECRET_JWT`

Essa variavel nao deve ficar hardcoded no codigo.

---

## O Que Vai No Pacote `internal/jwt`

Esse pacote costuma centralizar:

- gerar token
- validar token
- extrair claims
- calcular expiracao

Ele nao deve conhecer `gin.Context`.

Isso mantem o codigo reaproveitavel.

---

## O Que Vai No `service/auth`

O service de autenticacao costuma:

- validar credenciais
- buscar usuario no banco
- comparar senha
- pedir ao pacote JWT a geracao do token

Ele nao deve escrever resposta HTTP.

---

## O Que Vai No `handler/auth`

O handler de auth costuma:

- receber login e senha
- validar o body
- chamar o service
- devolver token em JSON

Exemplo de endpoints:

- `POST /auth/login`
- `GET /auth/me`

---

## O Que Vai No Middleware

O middleware deve:

- ler header `Authorization`
- validar formato `Bearer <token>`
- validar token
- colocar contexto autenticado na request
- bloquear acesso invalido com `401`

Esse middleware nao deve conhecer regra de negocio de `buyer`, `order` ou `delivery`.

---

## Ordem De Protecao Das Rotas

Proteja primeiro:

- rotas de alteracao de dados
- rotas administrativas
- fluxos operacionais sensiveis

Voce pode deixar publicos no inicio:

- `GET /check-health`
- `POST /auth/login`

---

## Cuidados Importantes

- nao guardar senha em texto puro
- nao colocar segredo JWT fixo no codigo
- nao colocar regra de autenticacao dentro do handler de dominio
- nao misturar validacao do token com autorizacao de negocio

---

## Relacao Com A Apostila

Esse capitulo deve ser estudado depois de:

- `12-criando-main-go.md`
- `01-crud-buyers.md`
- `07-guia-unico-testes-fase1.md`

---

## Proximo Capitulo

Depois de JWT, o proximo passo natural e:

- `16-middlewares-compartilhados.md`
