# 📝 Teste de Triangulação Técnica: WhatsApp SQL Forensics

---

## O "Mapa" (Estrutura e Tabelas)

### 1. No início, tentamos ligar a tabela message diretamente à tabela jid. Por que isso resultou em "mensagens fantasmas" ou no rótulo errado de status@broadcast?

- Resultou no rótulo `status@broadcast` por que precisavamos de um mediador de IDs para identificar quem corretamente era o remetente da mensagem.

### 2. Qual é a função da tabela chat no "meio de campo" entre a mensagem e o endereço real (JID)? Por que o WhatsApp não liga a mensagem direto ao contato?

- Ela armazena o estado das conversas, pois serve à tela de chats/conversas a informação necessária para estruturar a UI.

---

## A "Ponte" (LID vs. PN)

### 3. Explique a diferença entre um LID (@lid) e um PN (@s.whatsapp.net). Por que o Mello aparecia com nome em alguns lugares e com um ID numérico longo em outros?

- Lid são identificadores únicos gerados pelo WhatsApp para mascarar os PN (números de telefone), Mello aparecia como lid em alguns lugares que referenciavam o identificador, provávelmente por alguma transição imcompleta do WhatsApp para métodos mais privados e seguros.

### 4. Tivemos um momento "Eureka" ao encontrar a tabela status_ranking. Por que ela foi chamada de "Pedra de Roseta" para a nossa query?

- Ela foi quem revelou a ligação entre Lids e PNs

---

## SQL Avançado (Joins e Lógica)

### 5. Usamos JOIN (Inner Join) para a tabela chat, mas usamos LEFT JOIN para a tabela wa_contacts. Qual é a diferença prática de comportamento se trocasse tudo para JOIN simples?

- Nós precisávamos apenas dos nomes de contatos que têm na tabela chat, enquanto precisávamos de mais informações em wa_contacts. Um Inner Join devolve a intersecção entre tabela A e B, onde apenas as correspondências de ambas aparecem, enquanto a Left Join deve as intersecção entre tabela A e B, onde, apenas as correspondências da B aparecem junto à tabela A no produto final. se tivéssemos Joins simples, teriamos perca de informações das tabelas principais.

### 6. O que o comando ATTACH DATABASE faz e por que ele foi essencial para resolver os nomes dos contatos?

- Ele anexa um segundo banco de dadados (conjunto de tabelas), e facilita o trabalho de uma única query.

### 7. Na query final, usamos a função COALESCE. Explique o que ela faz e por que usamos uma sequência de 4 ou 5 opções dentro dela para definir o Nome_Conversa.

- Ela trata os dados nulos, ou vazios para serem humanamente identificáveis.

---

## Investigação e Debug

### 8. O que significa o sufixo -wal (ex: msgstore.db-wal) que mencionei, e por que ele é importante para quem faz sincronização de dados em tempo real?

- Write Ahead (escrita frente a logs), significa um banco de dados voltado apenas para logs do principal. Ele permite que cada registro tenha uma camada de segurança contra falhas, o mantendo íntegro

### 9. Por que a query precisou fazer o JOIN com a tabela wa_contacts duas vezes (uma para o Chat e outra para o Remetente)?

- Cada tabela tem sua particularidade, mas em ambas têm correspondências para encontrar o nome salvo do contato.

### 10. Se amanhã o WhatsApp criar um novo tipo de identificador (ex: @private), o que você precisaria fazer na sua query para ela continuar funcionando?

- Acredito que primeiro tenho que identificar em qual banco, e qual tabela eles vão passar a salvar esse identificador, e quais as particularidades dele. Com certeza isso vai pedir mais joins, por esperarmos que o whatsapp continue usando sua infraestrutura de cascata, então para meu código continuar funcionando, eu diria que isso vai depender do quão radical será a mudança do aplicativo.

---

# Correção do Mentor (IA)

| Questão                  | Nota | Feedback                                                                                                                                                                                         |
| :----------------------- | :--: | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **1. Erro Inicial**      |  7   | Você entendeu que faltava um mediador. O ponto exato é a **Colisão de IDs**: o ID `1` na tabela `message` aponta para a linha `1` da tabela `chat`, mas na tabela `jid`, a linha `1` é o Status. |
| **2. Função da Chat**    |  8   | Perfeito. Ela é o **Controller** da interface. Sem ela, o WhatsApp não saberia o que mostrar na tela inicial.                                                                                    |
| **3. LID vs PN**         |  9   | Resposta precisa. LID é a máscara, PN é a identidade real. Você capturou bem o motivo da "esquizofrenia" do banco.                                                                               |
| **4. Pedra de Roseta**   |  10  | **Excepcional.** Curto e direto. Sem a `status_ranking`, teríamos o dado (LID), mas não a informação (Quem é o dono do LID).                                                                     |
| **5. Join vs Left Join** |  8   | Correto. Se usássemos `JOIN` simples, mensagens de pessoas não salvas na agenda sumiriam do resultado. O `LEFT JOIN` é a nossa rede de segurança.                                                |
| **6. Attach Database**   |  8   | Exato. É o que permite o SQL "enxergar" além das paredes de um único arquivo.                                                                                                                    |
| **7. Coalesce**          |  9   | Ótima definição. É o nosso "Plano B, C e D" para garantir que o usuário nunca veja um campo vazio.                                                                                               |
| **8. Sufixo -WAL**       |  7   | Você entendeu a parte da integridade. O pulo do gato para o Sync em Tempo Real é que o WAL contém as mensagens que acabaram de chegar e ainda não foram "consolidadas".                          |
| **9. Join Duplo**        |  6   | Razão técnica: uma mensagem tem duas identidades distintas: o "Lugar" (Chat) e a "Pessoa" (Remetente). Precisamos de dois Joins para buscar nomes diferentes ao mesmo tempo.                     |
| **10. Novo ID**          |  8   | Mentalidade de desenvolvedor sênior: primeiro investigar, depois adaptar os Joins.                                                                                                               |

**Média Final: 8.0/10**