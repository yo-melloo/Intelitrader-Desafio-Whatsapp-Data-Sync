/*
 * Agente NDK para Android (Golang)
 * Sincronizador de mensagens e contatos entre WhatsApp SQLite e Redis.
 * Protegido contra OOM Killer via prioridade de processo (-1000).
 */

package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
)

func getWhatsAppDbDir() string {
	// Variável de ambiente tem prioridade máxima (se informada)
	envDir := os.Getenv("WA_DB_DIR")
	if envDir != "" {
		return envDir
	}
	// Tenta primeiro o WhatsApp Business
	w4bDir := "/data/data/com.whatsapp.w4b/databases/"
	if _, err := os.Stat(w4bDir); err == nil {
		return w4bDir
	}
	// Fallback para o WhatsApp Normal
	return "/data/data/com.whatsapp/databases/"
}

// Setup global
var (
	dbDir = getWhatsAppDbDir()
	databasePath string = dbDir + "msgstore.db"
	databaseWalPath = databasePath + "-wal"
	lastProcessID int
	timer *time.Timer
	// Proteção contra concorrência
	mu         sync.Mutex 
	isFetching bool
)

type MessageRow struct {
	ID           int
	Contexto     sql.NullString
	NomeConversa sql.NullString
	Remetente    sql.NullString
	Conteudo     sql.NullString
	Horario      sql.NullString
}

type MessageDTO struct {
    ID           int    `json:"id"`
    Contexto     string `json:"contexto"`
    NomeConversa string `json:"nome_conversa"`
    Remetente    string `json:"remetente"`
    Conteudo     string `json:"conteudo"`
    Horario      string `json:"horario"`
}

// DTO para recebimento de comandos do Redis (Injeção de Contatos)
type ContactDto struct {
	Name   string `json:"name"`
	Number string `json:"number"`
}

//go:embed query-copy.sql
var sqlQueryMessages string

var ctx = context.Background()


func main() {

	// 0. Proteção contra o OOM Killer
    pid := os.Getpid()	// Obtém o PID do processo do agente
    scorePath := fmt.Sprintf("/proc/%d/oom_score_adj", pid)	// obtém a lista de prioridades do Android

	err := os.WriteFile(scorePath, []byte("-1000"), 0644) // adiciona a prioridade personalizada do agente
    if err != nil {
		log.Printf("[NATIVE AGENT] Aviso: Não foi possível elevar prioridade: %v", err)
        // Isso pode falhar se não houver root, mas o agente continuará rodando
    } else { log.Println("[NATIVE AGENT] Prioridade de processo definida para máxima (-1000)") }


	// 1. Conexão com Banco de dados
	dsn := databasePath + "?_pragma=journal_mode(WAL)&_pragma=synchronous=NORMAL&_pragma=query_only=1"
	database, err := sql.Open("sqlite", dsn)
	if err != nil { log.Fatal(err) }
	err = database.Ping()
	if err != nil { log.Fatal("Erro ao conectar com banco de dados:", err) }
	
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(0)
	defer database.Close()

	// Anexa BD secundário (wa.db) de forma dinâmica
	attachQuery := fmt.Sprintf("ATTACH DATABASE '%swa.db' AS wa_db", dbDir)
	_, err = database.Exec(attachQuery)
	if err != nil { log.Printf("[SQLITE3] Aviso: Erro ao anexar wa.db: %v", err) }
	log.Println("[SQLITE3 DRIVER] Conexão com os bancos de dados estabelecida com sucesso!")


	// 2. Estabelece conexão com Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "10.0.2.2" // Fallback para emulador local
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":6379",
		Password: "",
		DB:       0,
	})
	
	err = rdb.Ping(ctx).Err()
	if err != nil {
    	log.Fatalf("[Redis] Erro ao conectar: %v", err)
	}
	log.Println("[Redis] Conxão estabelecida com sucesso!")
	defer rdb.Close()

	// 3. Inserir Watcher no diretório
	watcher, err := fsnotify.NewWatcher()
	if err != nil { log.Fatal(err) }
	defer watcher.Close()
	log.Println("[WATCHER] Inicializado com sucesso!")

	if err := watcher.Add(dbDir); err != nil { log.Fatal(err) }
	log.Printf("[WATCHER] Watcher adicionado ao diretório: \"%s\"\n", dbDir)


	// 4. Calibra o lastProcessID
	var tempMsg sql.NullString
	err = database.QueryRow("SELECT _id, text_data FROM message ORDER BY _id DESC LIMIT 1").Scan(&lastProcessID, &tempMsg)
	if err != nil { log.Fatal("[SQLITE3 DRIVER] Erro ao calibrar ID inicial: ", err) }
	log.Printf("[SQLITE3 DRIVER] lastProcessID calibrado para: %d\n", lastProcessID)
	
	// Confirma calibração:
	fmt.Printf("[WATCHER] A útlima mensagem foi: %s\n", tempMsg.String)


	// 5. Goroutine que monitora o banco de dados do WhatsApp
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Fallback de polling para prevenir atrasos do inotify
		defer ticker.Stop()

		for {
			select {

				// Tratamento de eventos:
				case event, ok := <-watcher.Events:
					if !ok { return }

					// Log de debug:
					//log.Println("[DEBUG EVENT]:", event.Name, event.Op) // nome do evento, descrição do evento

					// Verificamos apenas eventos envolvem arquivo 'msgstore.db'
					// Usamos strings.Contains ou filepath.Base para evitar erros de path absoluto/relativo
					isDbEvent := strings.Contains(event.Name, "msgstore.db")
					
					if isDbEvent {
						// No Android, o SQLite WAL gera muitos eventos de atribuição de permissão (Chmod)
						// O agente deve reagir a quase tudo que signifique apenas "mudança"
						if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {	// verifica qual dos eventos aconteceu
							go pullMessage(database, rdb, sqlQueryMessages)
						}
					}
				
				case <-ticker.C:
					go pullMessage(database, rdb, sqlQueryMessages)
				
				case err, ok := <-watcher.Errors:
					if !ok { return }
					log.Println("[WATCHER ERROR]", err)
			}
		}
	}()

	// 8. Inicia o Motor de Escrita (Subscriber de Contatos) em paralelo
	go startContactSubscriber(rdb)

	<-make(chan struct{})
}

// Nova Goroutine que escuta ordens de injeção de contatos (Redis -> Android)
func startContactSubscriber(rdb *redis.Client) {
	channelSubscribe := "contacts:insert"
	pSub := rdb.Subscribe(ctx, channelSubscribe)
	log.Printf("[CONTACT SUBSCRIBER] Escutando canal: %s", channelSubscribe)
	defer pSub.Close()

	for msg := range pSub.Channel() {
		log.Println("[CONTACT SUBSCRIBER] Nova ordem recebida via Redis!")
		contactShoot(msg.Payload)
	}
}


// Lê o banco de dados e processa novas mensagens
func pullMessage(database *sql.DB, rdb *redis.Client, querysqlQueryMessages string) {
	
	// Proteção contra Race Condition:
    mu.Lock() // bloqueia o acesso às variáveis das outras Goroutines
    if isFetching { // verifica se o acesso já está interrompido
        mu.Unlock() // devolve o acesso
        return // reinicia
    }

    isFetching = true // interrompe o acesso de outras Go routines, se não estiver
    mu.Unlock() // começa a trabalhar com as variáveis

    defer func() { // quando terminar pullMessage() ou der erro...
        mu.Lock()
        isFetching = false // desocupa as variáveis
        mu.Unlock() // retoma acesso das outras goroutines
    }()

    // Filtro 2: Correção do N+1 Query Problem usando um Cursor (rows.Next())
    rows, err := database.Query(querysqlQueryMessages, lastProcessID)
    if err != nil {
        log.Printf("[SQLITE3] Erro ao executar query de mensagens: %v", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var m MessageRow
        
        err := rows.Scan(
            &m.ID, 
            &m.Contexto, 
            &m.NomeConversa, 
            &m.Remetente, 
            &m.Conteudo, 
            &m.Horario,
        )

        if err != nil || m.ID <= lastProcessID { continue }

        lastProcessID = m.ID

        // Helper para tratar os NullStrings
        format := func(ns sql.NullString) string {
            if ns.Valid { return ns.String }
            return "NULO"
        }

		dto := MessageDTO{
			ID:				m.ID,
			Contexto:		format(m.Contexto),
			NomeConversa:	format(m.NomeConversa),
			Remetente:		format(m.Remetente),
			Conteudo:		format(m.Conteudo),
			Horario:		format(m.Horario),
		}
		stringID := fmt.Sprintf("msg:%d", m.ID)

		// Se for necessário futuramente, o Agente já produz o JSON da mensagem:
		jsonData, err := json.Marshal(dto)
		if err != nil {
			log.Printf("Erro ao gerar JSON: %v", err)
		} else {
			fmt.Println("[Marshall JSON] JSON Gerado:", string(jsonData))
		}
		
		// 6. Log e formatação das mensagens capturadas
        log.Printf("\n[MENSAGEM CAPTURADA]\nContexto: %s | Remetente: %s\nConversa: %s | Horário: %s\nConteúdo: %s", 
		    format(m.Contexto), 
            format(m.Remetente), 
            format(m.NomeConversa), 
            format(m.Horario),
            format(m.Conteudo), 
        )

		// 7. Exporta mensagem para Redis:
		redisHash := createStringHash(dto)
		pushStringHash(ctx, rdb, stringID, redisHash)
    }
}

func pushStringHash(ctx context.Context, rdb *redis.Client, stringID string, redisHash map[string]interface{}){

	// Salva verificando conexão
	err := rdb.HSet(ctx, stringID, redisHash).Err()
	if err != nil {
		log.Printf("[REDIS ERRO] ❌ Falha ao salvar hash '%s': %v\n--------------------------------------------------\n", stringID, err)
		return
	}

	// Publica a mudança
	err = rdb.Publish(ctx, "general:hash-updates", stringID).Err()
	if err != nil {
		log.Printf("[REDIS ERRO] ❌ Falha ao publicar notificação de '%s': %v\n--------------------------------------------------\n", stringID, err)
	} else {
		log.Printf("[REDIS INFO] ✅ Mensagem '%s' enviada com sucesso!\n--------------------------------------------------\n", stringID)
	}
}

func createStringHash(dto MessageDTO) map[string]interface{} {

	redisHash := map[string]interface{}{
		"Contexto":     dto.Contexto,
		"NomeConversa": dto.NomeConversa,
		"Remetente":    dto.Remetente,
		"Conteudo":     dto.Conteudo,
		"Horario":      dto.Horario,
	}
	return redisHash

}

// Funções de suporte para o Motor de Escrita (Contatos)
func getActiveAccount() (name string, accType string) {
	cmd := exec.Command("dumpsys", "account")
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	re := regexp.MustCompile(`name=([^, ]+), type=(com\.google)`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", ""
}

func contactShoot(contactJSON string) {
	fmt.Printf("\n--------------------------------------------------\n")
	fmt.Println("[CONTACT BUILDER] Processando nova ordem de injeção...")

	var contactDto ContactDto
	err := json.Unmarshal([]byte(contactJSON), &contactDto)
	if err != nil || contactDto.Name == "" || contactDto.Number == "" {
		log.Println("[CONTACT BUILDER] Erro: Dados do contato inválidos no JSON")
		return
	}

	accName, accType := getActiveAccount()
	bindAccName := "account_name:n:"
	bindAccType := "account_type:n:"

	if accName != "" {
		fmt.Printf("[CONTACT BUILDER] Modo Conta Google: %s\n", accName)
		bindAccName = "account_name:s:" + accName
		bindAccType = "account_type:s:" + accType
	} else {
		fmt.Println("[CONTACT BUILDER] Modo Local (Sem conta Google)")
	}

	// 1. Criar Raw Contact
	cmdInsertRaw := exec.Command("content", "insert", "--uri", "content://com.android.contacts/raw_contacts", "--bind", bindAccName, "--bind", bindAccType)
	if err := cmdInsertRaw.Run(); err != nil {
		log.Printf("[CONTACT BUILDER] Erro ao criar raw_contact: %v", err)
		return
	}

	// 2. Recuperar ID
	cmdGetID := exec.Command("content", "query", "--uri", "content://com.android.contacts/raw_contacts", "--projection", "_id", "--sort", "_id DESC")
	output, err := cmdGetID.Output()
	if err != nil {
		log.Printf("[CONTACT BUILDER] Erro ao consultar ID: %v", err)
		return
	}

	re := regexp.MustCompile(`_id=(\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		log.Println("[CONTACT BUILDER] Erro: ID não localizado")
		return
	}
	rawID := matches[1]

	// 3. Vincular Nome
	errName := exec.Command("content", "insert", "--uri", "content://com.android.contacts/data", "--bind", "raw_contact_id:i:"+rawID, "--bind", "mimetype:s:vnd.android.cursor.item/name", "--bind", "data1:s:"+contactDto.Name).Run()
	if errName != nil {
		log.Printf("[CONTACT BUILDER] Erro ao vincular Nome: %v", errName)
	}

	// 4. Vincular Telefone
	errPhone := exec.Command("content", "insert", "--uri", "content://com.android.contacts/data", "--bind", "raw_contact_id:i:"+rawID, "--bind", "mimetype:s:vnd.android.cursor.item/phone_v2", "--bind", "data1:s:"+contactDto.Number, "--bind", "data2:i:2").Run()
	if errPhone != nil {
		log.Printf("[CONTACT BUILDER] Erro ao vincular Telefone: %v", errPhone)
	}

	log.Printf("[CONTACT BUILDER] Processamento do contato '%s' finalizado.\n--------------------------------------------------\n", contactDto.Name)
}