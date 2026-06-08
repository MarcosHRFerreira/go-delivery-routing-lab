# Testes De Integracao Com Banco Real

Este capitulo mostra quando sair apenas de stubs e mocks e validar fluxo real com banco.

---

## Quando Criar

Esse tipo de teste deve entrar depois que:

- os testes unitarios basicos existem
- os repositories reais comecaram a acumular logica SQL relevante
- os fluxos principais merecem validacao de ponta a ponta parcial

---

## O Que Testar Aqui

Priorize:

- repositories
- integracao HTTP principal com banco real
- transacoes
- leituras e gravacoes importantes

Evite transformar tudo em teste de integracao pesado.

---

## Ordem Interna De Evolucao

1. preparar banco isolado para teste
2. subir schema de teste
3. popular dados minimos
4. executar caso de uso
5. validar estado final do banco
6. limpar ambiente ao final

---

## Pasta Sugerida

- `test/integration/`

Separacoes uteis:

- `test/integration/repository/`
- `test/integration/http/`

---

## Cuidados Importantes

- nao depender do banco de desenvolvimento manual
- nao compartilhar estado entre testes
- limpar dados ao final de cada execucao
- fechar conexoes

---

## Valor Didatico

Esse capitulo fecha muito bem a apostila porque mostra o salto de:

- testar unidade
- para validar comportamento real da aplicacao

---

## Fechamento

Com esse capitulo, a trilha passa a cobrir:

- bootstrap
- CRUD
- testes
- seguranca
- robustez
- documentacao
