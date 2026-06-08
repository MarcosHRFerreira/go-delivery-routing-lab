# CI CD E Pipeline De Testes

Este capitulo fecha a trilha mostrando como automatizar a qualidade e a entrega da API.

Depois que a aplicacao ja tem:

- bootstrap
- CRUDs
- testes
- documentacao
- empacotamento

o proximo passo natural e automatizar validacoes.

---

## Quando Criar

Esse capitulo entra melhor quando:

- o projeto ja compila com consistencia
- os testes ja conseguem rodar sem acao manual pesada
- o time quer reduzir regressao

---

## O Que Uma Pipeline Deve Fazer

Uma pipeline minima costuma:

- baixar dependencias
- validar formatacao
- rodar testes unitarios
- rodar testes de integracao quando fizer sentido
- montar artefato ou imagem

---

## Ordem Interna De Evolucao

1. garantir comandos reproduziveis localmente
2. automatizar `go test ./...`
3. adicionar validacao de build
4. adicionar validacao de imagem Docker
5. separar etapas de `build`, `test` e `deploy`

---

## Tipos De Pipeline

Os fluxos mais comuns sao:

- pipeline de pull request
- pipeline de merge para branch principal
- pipeline de release

---

## O Que Validar Primeiro

Na apostila, a sequencia mais simples e:

1. compilacao
2. testes unitarios
3. testes de integracao
4. build da imagem

---

## O Que Nao Fazer

- criar pipeline antes de conseguir rodar o projeto localmente
- esconder falhas importantes com etapas opcionais demais
- misturar deploy em producao com testes instaveis
- depender de configuracoes secretas nao documentadas

---

## Ligacao Com Outros Capitulos

Esse capitulo fecha bem a trilha junto com:

- `21-testes-de-integracao-com-banco-real.md`
- `24-deploy-com-docker-e-ambientes.md`
- `22-observabilidade-e-metricas.md`

---

## Valor Didatico

Esse tema mostra que maturidade de API tambem envolve:

- repetibilidade
- confiabilidade
- seguranca de entrega

Nao basta o codigo funcionar apenas na mao do desenvolvedor.

---

## Fechamento

Com esse capitulo, a trilha passa a cobrir:

- construcao da base
- evolucao dos CRUDs
- testes
- seguranca
- observabilidade
- empacotamento
- automacao de qualidade e entrega
