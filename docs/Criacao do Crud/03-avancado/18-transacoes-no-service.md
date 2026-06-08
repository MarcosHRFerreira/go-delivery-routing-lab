# Transacoes No Service

Este capitulo mostra quando um caso de uso precisa de transacao.

Transacao nao deve ser usada por reflexo.

Ela entra quando mais de uma operacao precisa acontecer como uma unidade atomica.

---

## Quando Criar

Use transacao quando um fluxo precisa:

- gravar mais de uma tabela
- atualizar estados relacionados
- manter consistencia mesmo em caso de erro intermediario

---

## Exemplos Reais No Dominio

No contexto de entregas, exemplos comuns seriam:

- criar uma entrega e atualizar o status do pedido
- confirmar conclusao da entrega e registrar historico
- reordenar entregas e salvar historico no mesmo fluxo

---

## Onde A Regra Deve Ficar

A decisao de abrir transacao costuma nascer no `service`.

Motivo:

- o `service` conhece o caso de uso completo
- o `repository` conhece apenas operacoes tecnicas

---

## Ordem Interna De Evolucao

1. identificar caso de uso atomico
2. definir como a transacao sera aberta
3. adaptar repositories para trabalhar com executor transacional
4. garantir rollback em falha
5. garantir commit em sucesso
6. testar caminho feliz e erro intermediario

---

## O Que Evitar

- abrir transacao para operacao simples de leitura
- deixar o handler controlar commit e rollback
- espalhar a mesma transacao em varios pontos sem coordenacao

---

## Beneficio Para A Apostila

Esse capitulo mostra com clareza a diferenca entre:

- CRUD simples
- fluxo de negocio mais robusto

---

## Proximo Capitulo

Depois de transacoes, o proximo tema natural e:

- `19-paginacao-em-listagens.md`
