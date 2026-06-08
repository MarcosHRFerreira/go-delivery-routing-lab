# OpenAPI E Swagger

Este capitulo mostra quando documentar a API formalmente.

---

## Quando Criar

OpenAPI costuma entrar depois que:

- os endpoints principais existem
- os contratos basicos estabilizaram
- os nomes de rotas e payloads pararam de mudar toda hora

---

## O Que A Documentacao Formal Entrega

- contratos claros
- exemplos de request e response
- apoio para front-end
- apoio para QA
- apoio para onboarding

---

## Ordem Interna De Evolucao

1. estabilizar endpoints principais
2. definir schemas de request e response
3. documentar status codes
4. documentar autenticacao
5. publicar Swagger UI ou arquivo OpenAPI

---

## Beneficio Didatico

Esse capitulo mostra que uma API madura nao termina no codigo.

Ela tambem precisa ser comunicavel.

---

## Ligacao Com Os CRUDs

Quando voce terminar a fase 1, os primeiros candidatos a documentacao sao:

- `buyers`
- `couriers`
- `orders`
- `deliveries`

---

## Proximo Capitulo

Depois de OpenAPI, vale fechar com:

- `21-testes-de-integracao-com-banco-real.md`
