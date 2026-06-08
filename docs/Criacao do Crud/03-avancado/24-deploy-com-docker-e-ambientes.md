# Deploy Com Docker E Ambientes

Este capitulo fecha a parte de empacotamento e execucao da API fora da maquina local.

Depois que a aplicacao cresce, nao basta funcionar no computador do desenvolvedor.

Ela precisa funcionar de forma previsivel em ambientes diferentes.

---

## Quando Criar

Esse capitulo entra melhor quando:

- a API ja sobe localmente com consistencia
- o banco ja esta separado por configuracao
- a aplicacao ja usa variaveis de ambiente

---

## O Que Esse Capitulo Ensina

- empacotar a aplicacao
- separar ambiente local, teste e producao
- evitar dependencia da maquina do desenvolvedor
- preparar terreno para CI/CD

---

## Conceitos Centrais

Os conceitos mais importantes aqui sao:

- `Dockerfile`
- `docker-compose.yml`
- variaveis de ambiente
- rede entre servicos
- volume quando necessario

---

## Ordem Interna De Evolucao

1. garantir que `config.go` usa ambiente corretamente
2. criar `Dockerfile` da API
3. definir imagem e porta exposta
4. conectar API e banco no `docker-compose`
5. validar health check em container
6. separar configuracao por ambiente

---

## Ambientes Minimos

Os ambientes minimos que vale considerar na apostila sao:

- desenvolvimento local
- testes automatizados
- producao

Cada ambiente pode mudar:

- credenciais
- host do banco
- nivel de log
- porta

---

## O Que Nao Fazer

- embutir segredo no `Dockerfile`
- depender de configuracao manual escondida
- usar o mesmo banco para tudo
- subir imagem sem health check e sem validacao basica

---

## Ligacao Com O Projeto

Esse capitulo conversa bem com:

- `08-criando-config-go.md`
- `11-criando-router-go-e-health-check.md`
- `12-criando-main-go.md`
- `22-observabilidade-e-metricas.md`

---

## Valor Didatico

Aqui a pessoa entende que construir API nao termina no `go run`.

Ela aprende a pensar em:

- portabilidade
- previsibilidade
- ambientes
- execucao empacotada

---

## Proximo Capitulo

Depois de Docker e ambientes, o proximo passo natural e:

- `25-ci-cd-e-pipeline-de-testes.md`
