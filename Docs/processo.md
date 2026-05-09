### Etapa-001: Configuração do Ambiente

- [x] Instalou GO
- [x] Executou Olá Mundo em GO: Entendeu que a linguagem funciona usando `packages`
- [x] Instalou Android Studio
- [x] Instalou, executou e configurou um Emulador de Android
  - Primeira tentativa deu erro: Usei uma imagem Android que não dava acesso root (Pixel 14 Pro)
  - Segunda tentativa deu certo: Usei uma imagem Android que dava acesso root (Pixel 6 Pro)
  - Tentei instalar APK do WhatsApp Business, mas o emulador não instalou. Tentei do WhatsApp comum, e funcionou.
  - Personalizei o Android para entender o quanto ele é limitado na emulação, e na minha máquina (não é - o que é bom).
- [x] Baixou NDK (que não vem instalado por padrão)
- [x] Realizou teste de acesso ao root do sistema (manual), copiando o banco de dados do WhatsApp e analisando na máquina local
  - Usei SQLite para abrir o arquivo `msgstore.db` copiado para o armazenamento do meu computador
  - Identifiquei dezenas de tabelas, e pesquisei qual delas armazena as mensagens (literalmente uma tabela chamada `message`) [Informação vai servir mais tarde para construir a consulta SQL]
- [x] Criou container Docker para usar Redis na máquina local
- [x] Usou IA para criar o primeiro código `main.go`:
  - Código revisado (vide comentários adicionados)
  - Aprendeu buildar binários do Go
  - Aprendeu a dar push em arquivos via Adb
  - Ao ser buildado e pushado para o Android, o binário tenta conexão com o banco de dados do WhatsApp, e com o Redis (validando conexão)
- [x] Usou IA para reescrever o `.gitignore`, e manter boas práticas para evitar commitar segredos e dependências pesadas.
- [x] Usou IA para reescrever o `README.md`, e melhorar a descrição pré-escrita do projeto.
- [x] Gerou arquivo LICENSE com as condições de uso do projeto (restrito apenas para o teste).

---

As etapas processo foram devidamente organizadas em um quadro Kanban usando Trello: https://trello.com/b/SuVJxaAJ/desafio-intelitrader
