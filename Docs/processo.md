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

> Commmit: 8d45888 feat(ambiente): Configura repositório, documentação e prepara Agente GO para execução no Android

---

### Etapa-002: Implementar leitura recursiva do banco de dados

#### DECISÃO-TÉCNICA-001: Polling vs. Observer

Após analisar, entendi que Polling realizaria uma consulta periódica no banco de dados, enquanto o Observer iria aguardar alguma mudança acontecer para disparar a função do agente. Para me decidir entre um e outro, levei em conta que o critério da aplicação é ser em tempo real, se reduzisse o tempo de polling isso poderia ser prejucidial tanto à bateria e performance quanto à experiencia. Já que existe um processo rodando ativamente chamado inotify no kernel Linux (base do Android) que pode ser apontado para o Write-Ahead Logging (WAL) do banco de dados do WhatsApp, programar um Observer seria como programar um gatilho que só dispara quando necessário, economizando recursos.

Decisão: Usar Observer para "vigiar" banco de dados com o próprio "vigia do sistema" (inotify)
Consequência: Arquitetura do projeto a partir desse ponto vira um Event-Driven.

Problema identificado: Alguns programas podem disparar mais de uma notificação para uma simples operação, o que pode gerar consultas desnecessárias ao banco de dados

```bash
$ go run .

# Usei .txt nesse exemplo
2026/05/09 18:50:37 [WATCHER] Watcher adicionado ao arquivo: ".//db/teste-db.txt"

# Um único save, duas ações de Write
[WATCHER] Evento detectado

2026/05/09 18:50:40 [WATCHER] O arquivo foi modificado.
WRITE

[WATCHER] Evento detectado
2026/05/09 18:50:40 [WATCHER] O arquivo foi modificado.
WRITE


exit status 0xc000013a
```

Solução: Criado fila e filtro de mudanças. Durante o processo me deparei com uma race condition, relatado em `Docs/dificuldades.md`

```go
case event, ok := <-watcher.Events:
  if !ok {
    return
  }
  log.Println("[WATCHER] Evento detectado")

  if event.Has(fsnotify.Write) {
    if timer != nil {
      timer.Stop() // para timers anteriores, se existirem, para evitar logs repetidos
    }

    timer = time.AfterFunc(200*time.Millisecond, func() { // inicia um novo timer para gerar o log após um curto atraso
      buscarUltimaMensagem(database) // o log que exibe o conteúdo do registro encontrado no banco de dados está nessa função
    })
  }

  // ...

func buscarUltimaMensagem(db *sql.DB) { // Lógica de fila para buscar os próximos registros do banco de dados a partir do lastProcessID (global)

	query := "SELECT id, content FROM example WHERE id > ? ORDER BY id ASC" // Seleciona o próximo registro com ID maior que lastProcessID

	rows, err := db.Query(query, lastProcessID)
	if err != nil {
		log.Println("[SQL] Falha na consulta. Verifique a conexão com o banco de dados e a estrutura da tabela.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var conteudo string

		if err := rows.Scan(&id, &conteudo); err != nil {
			fmt.Printf("[SQL] Erro ao ler os dados da linha: %v.\n", err)
			continue
		}

		fmt.Printf("[WATCHER] Conteúdo adicionado - ID: %d | Conteúdo: %s\n", id, conteudo)

		lastProcessID = id // Atualiza o lastProcessID (global) para o ID do registro encontrado
	}

}

```

Resultado:

```bash

$ go run .

2026/05/10 01:04:48 [NATIVE AGENT] Iniciando o observador de banco de dados...
2026/05/10 01:04:48 [SQLITE3 DRIVER] Conexão com o banco de dados estabelecida com sucesso.
2026/05/10 01:04:48 [WATCHER] Watcher adicionado ao arquivo: ".//db/teste-wal.db-wal"
[WATCHER] Conteúdo adicionado - ID: 47 | Conteúdo: w
[WATCHER] Conteúdo adicionado - ID: 48 | Conteúdo: x
[WATCHER] Conteúdo adicionado - ID: 49 | Conteúdo: y
[WATCHER] Conteúdo adicionado - ID: 50 | Conteúdo: z

```

- [x] Desenvolveu lab (teste) para implementar a leitura do banco de dados
  - [x] Aprendeu importar e usar Sqlite em Go
  - [x] Entendeu as Goroutines e Loops infinitos
  - [x] Fez o uso de Observer/Watcher de arquivos para ler o WAL de um banco de dados, para criar "gatilhos"
  - [x] Planejou a estrutura do código que será executado no Agente (antes de implementar o Redis)
  - [x] Resolveu uma Race Condition
  - [x] Testou manualmente conexão com banco de dados via ADB
  - [x] Resolveu Chatters (logs repetitivos) que aconteciam por serem executados no lugar errado dos loops

---

As etapas processo foram devidamente organizadas em um quadro Kanban usando Trello: https://trello.com/b/SuVJxaAJ/desafio-intelitrader
