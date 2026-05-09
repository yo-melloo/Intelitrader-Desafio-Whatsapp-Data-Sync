# 📱 Intelitrader Desafio Técnico: WhatsApp Data Sync

[![Status: Em Desenvolvimento](https://img.shields.io/badge/Status-Em_Desenvolvimento-yellow.svg)](Docs/processo.md)
[![Candidato: Gustavo Melo](https://img.shields.io/badge/Candidato-Gustavo_Melo-blue.svg)](Docs/Sobre-Gustavo.md)

Este repositório contém a solução para o **Desafio Técnico de Integração e Monitoramento Android Real-Time**, proposto pela **Intelitrader**.

O objetivo principal é construir um ecossistema que integre a extração de mensagens em tempo real do banco de dados do WhatsApp em um dispositivo Android e a inserção remota de contatos, utilizando um Agente Nativo (Golang) e uma Interface Externa (C# .NET) com comunicação intermediada pelo Redis.

---

## 🎯 Objetivos do Desafio

### Agente Nativo (Android/Golang)

- [ ] Monitorar o banco SQLite (`msgstore.db`) de mensagens do WhatsApp e, a cada nova inserção, enviar o conteúdo para uma instância do Redis.
- [ ] Escutar uma fila/tópico no Redis.
- [ ] Ao receber um novo comando, inserir um contato na agenda telefônica do Android (via content insert ou chamadas de API do sistema).

### Interface Externa (C# .NET)

- [ ] Ler as mensagens publicadas no Redis e exibi-las no console (ou em um log/dashboard simples) em **tempo real**.
- [ ] Expor um endpoint `POST /contacts`, que deve publicar os dados do novo contato no Redis para que o Agente Nativo processe a inserção no dispositivo Android.

---

## 🛠️ Tecnologias e Arquitetura

- **Agente Nativo:** Golang (cross-compilado nativamente via NDK para arquitetura do Android).
- **Interface Externa:** C# (.NET).
- **Mensageria e Cache:** Redis (rodando localmente via Docker).
- **Ambiente de Teste:** Emulador Android Studio (Pixel 6 Pro) com acesso Root liberado.

### Arquitetura da Solução

1. O aplicativo **WhatsApp** (no Emulador Root) escreve suas mensagens no banco `msgstore.db`.
2. O **Agente Nativo (Go)** roda em background no Android, lê as novas entradas e faz o push para o **Redis**.
3. A **Interface Externa (C#)** consome o Redis para exibir as mensagens e envia comandos de volta ao Redis quando o endpoint `/contacts` é acionado.
4. O **Agente Nativo (Go)** consome o comando do Redis e interage com os provedores de conteúdo do Android para salvar o contato.

---

## 🚀 Progresso Atual: Etapa 001 - Configuração do Ambiente

O projeto está sendo construído passo a passo. Até o momento, a base ambiental e de pesquisa foi estabelecida:

- [x] Configuração inicial do ambiente Go e Android Studio.
- [x] Criação de um emulador Android com acesso **Root** (Pixel 6 Pro) para contornar as limitações de acesso aos dados de outros apps.
- [x] Instalação do WhatsApp no emulador e mapeamento manual do banco de dados SQLite (`msgstore.db` -> tabela `message`).
- [x] Download e configuração do NDK para compilação C/C++/Go para Android.
- [x] Configuração do contêiner Docker para o Redis.
- [x] Desenvolvimento de um código base em `main.go` para compilação cruzada (cross-compile) e envio ao emulador via `adb push`, realizando os primeiros testes de conexão entre o ambiente nativo e o Redis.

---

## 📚 Documentação Auxiliar

Ao longo do desenvolvimento, documentações detalhadas sobre o processo, decisões técnicas e aprendizados estão sendo registradas (este readme foi revisado textualmente com ajuda de IA, mas a documentação NÃO):

- 👤 **[Sobre o Candidato (Gustavo Melo)](Docs/Sobre-Gustavo.md)**: Um pouco sobre a minha trajetória, o pragmatismo no aprendizado de Go e C# para esse desafio, e como abordo arquitetura e engenharia.
- 📝 **[Decisões Técnicas e Processo](Docs/processo.md)**: Registro do fluxo de trabalho e das decisões arquiteturais tomadas.
- 🚧 **[Dificuldades Encontradas](Docs/dificuldades.md)**: Um log honesto sobre os desafios técnicos enfrentados (como a curva de aprendizado de novas linguagens) e as soluções adotadas.
- 📱 **[Limitações do Android](Docs/limitacoes.md)**: _Em breve_. Explicações problemas enfrentados com a arquitetura do Android.

---

_Construído com dedicação para a equipe técnica da Intelitrader._
