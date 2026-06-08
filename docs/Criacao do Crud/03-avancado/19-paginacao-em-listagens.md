# Paginacao Em Listagens

Este capitulo mostra quando listagens simples deixam de ser suficientes.

---

## Quando Criar

Paginacao entra quando:

- a tabela cresce
- o front precisa navegar em blocos
- carregar tudo de uma vez deixa de ser saudavel

---

## Onde Isso Costuma Tocar

A paginacao normalmente afeta:

- `dto`
- `handler`
- `service`
- `repository`
- documentacao da API

---

## Ordem Interna De Evolucao

1. definir parametros de consulta
2. validar `page` e `limit`
3. traduzir isso para `LIMIT` e `OFFSET`
4. devolver metadados de navegacao
5. testar pagina vazia, pagina invalida e pagina parcial

---

## Estrutura Mental

Uma listagem paginada costuma devolver:

- itens
- pagina atual
- limite
- total
- total de paginas

---

## O Que Ensinar Na Apostila

Esse capitulo e bom para mostrar que a API evolui em camadas:

- primeiro a listagem simples funciona
- depois ela fica preparada para crescer

---

## Proximo Capitulo

Depois de paginacao, o proximo tema natural e:

- `20-openapi-e-swagger.md`
