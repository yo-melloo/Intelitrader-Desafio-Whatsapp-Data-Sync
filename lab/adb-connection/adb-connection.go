/*
* Esse código é um agente semelhante ao WhatsAppSync/main.go,
* aqui a conexão com o SQLite vai tentar mirar diretamente no banco de dados do WhatsApp do android e exibir as mensgens (brutas) no console,
* a itenção é identificar e mitigar os desafios de acesso ao ambiente do Android e o acesso ao banco de dados do WhatsApp,
* Condição de sucesso: o console exibir as mensagens que começarão a chegar no whatsapp após a inicialização, sem perdas.
* Definição de Sucesso -> Conclui Etapa de Implementação de leitura do banco de dados -> Integrar em WhatsAppSync/main.go -> Iniciar integração com Redis (Proxima etapa) + modelagem de dados
 */

package main

/* Problemas identificados: <<<<<<<<<<<<<<<<<<<<<
*
* 1. Permissão de root - O agente sofre de restrições em pastas do sistema, sendo visto como processo não prioritário
* 2. Hierarquia de permissões - Para o banco de dados em um Android, dificilmente duas atividades podem acessar o banco de dados com permissão integral (edição),
* 	 a menos que uma delas esteja apenas em readonly (apenas leitura),
* 3. Dooze Mode - o sistema precisa reconhecer o processo como um serviço prioritário e evitar "matar" ele
* 4. Arquitetura de Abstrações de baixo consumo - O Kernel do Android é otimizado para economizar bateria e desempenho, então alguns eventos não são processados em tempo real como no computador
*
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>*/

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "modernc.org/sqlite"
)

// Setup global
var (
	dbDir = "/data/data/com.whatsapp/databases/"
	databasePath string = dbDir + "msgstore.db"
	databaseWalPath = databasePath + "-wal"
	lastProcessID int
	timer *time.Timer
	// Proteção contra concorrência
	mu         sync.Mutex 
	isFetching bool
)

type MessageRow struct {	
	// é como uma classe, só que sem polimorfismo
	// é útil para criar "objetos"
	ID           int
	Contexto     sql.NullString
	NomeConversa sql.NullString
	Remetente    sql.NullString
	Conteudo     sql.NullString
	Horario      sql.NullString
}

//go:embed query-copy.sql
var sqlQueryMessages string

func main() {

	/* Processo: <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	* 0. Estabelece proteção contra OOM Killer
	* 1. Estabelece e configura conexão com banco de dados
	* 2. Insere Watcher para monitorar o diretório de dados do WhatsApp, com flags para trabalhar conforme Write-Back Cache 
	* 3. Calibra o lastProcessID para o ID mais recente no banco de dados, garantindo que apenas novas mensagens sejam processadas
	* 4. Goroutine que monitora o banco de dados do WhatsApp
	* 5. Loga as mensagens novas que chegarem no WhatsApp após a inicialização do agente
	*
	>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>*/ 

	// 0. Proteção contra o OOM Killer
	// Certifique-se de que o binário seja executado via su ou por um script de boot do Magisk
    pid := os.Getpid()	// Obtém o PID do processo do agente
    scorePath := fmt.Sprintf("/proc/%d/oom_score_adj", pid)	// obtém a lista de prioridades do Android

	err := os.WriteFile(scorePath, []byte("-1000"), 0644) // adiciona a prioridade personalizada do agente
    if err != nil {
		log.Printf("[NATIVE AGENT] Aviso: Não foi possível elevar prioridade: %v", err)
        // Isso pode falhar se não houver root, mas o agente continuará rodando
    } else { log.Println("[NATIVE AGENT] Prioridade de processo definida para máxima (-1000)") }


	// 1. Conexão com Banco de dados:
	dsn := databasePath + "?_pragma=journal_mode(WAL)&_pragma=synchronous=NORMAL&_pragma=query_only=1" 	// o uso de ? permite adicionar "flags" que configuram o acesso ao banco de dados
	database, err := sql.Open("sqlite", dsn)
	if err != nil { log.Fatal(err) }
	err = database.Ping()
	if err != nil { log.Fatal("Erro ao conectar com banco de dados:", err) }
	
	// Parâmetros de conexão com bancos de dados:
	database.SetMaxOpenConns(1)	// Conexões a serem abertas pelo Watcher
	database.SetMaxIdleConns(1)	// Conexões que podem serem mantidas em idle
	database.SetConnMaxLifetime(0)	// tempo de expiração da conexão (0 = Não Definido/Até o processo ser cancelado)
	defer database.Close()

	// Anexa BD secundário:
	_, err = database.Exec("ATTACH DATABASE '/data/data/com.whatsapp/databases/wa.db' AS wa_db")
	if err != nil { log.Printf("[SQLITE3] Aviso: Erro ao anexar wa.db: %v", err) }
	log.Println("[SQLITE3 DRIVER] Conexão com os bancos de dados estabelecida com sucesso!")


	// 2. Inserir Watcher no diretório:
	watcher, err := fsnotify.NewWatcher()	// Instancia novo Watcher
	if err != nil { log.Fatal(err) }
	defer watcher.Close()
	log.Println("[WATCHER] Inicializado com sucesso!")

	if err := watcher.Add(dbDir); err != nil { log.Fatal(err) }
	log.Printf("[WATCHER] Watcher adicionado ao diretório: \"%s\"\n", dbDir)


	// 3. Calibra o lastProcessID para o ID:
	var tempMsg sql.NullString	// Agente agora lida com Null Strings do SQL = Quando a query não funciona ou o resultado é nulo por algum padrão
	err = database.QueryRow("SELECT _id, text_data FROM message ORDER BY _id DESC LIMIT 1").Scan(&lastProcessID, &tempMsg)
	if err != nil { log.Fatal("[SQLITE3 DRIVER] Erro ao calibrar ID inicial: ", err) }
	log.Printf("[SQLITE3 DRIVER] lastProcessID calibrado para: %d\n", lastProcessID)
	
	// Confirma calibração:
	fmt.Printf("[WATCHER] A útlima mensagem foi: %s\n", tempMsg.String)


	// 4. Goroutine que monitora o banco de dados do WhatsApp:
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Fallback para Polling para evitar que o Watcher "pisque"
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
							go pullMessage(database, sqlQueryMessages)
						}
					}
				
				case <-ticker.C:
					// O Ticker acordar o Watcher se o cache do Android demorar a processar a próxima alteração (Heartbeat)
					go pullMessage(database, sqlQueryMessages)
				
				case err, ok := <-watcher.Errors:
					if !ok { return }
					log.Println("[WATCHER ERROR]", err)
			}
		}
	}()

	<-make(chan struct{})
}


// Lê o banco de dados e retorna a útlima mensagem inserida
// Retorna registro com pouca modelagem - TODO [x]: MODELAR DADOS PARA EXIBIÇÃO ÍNTEGRA DE MENSAGENS
func pullMessage(database *sql.DB, querysqlQueryMessages string) {
	
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

    for {
        time.Sleep(50 * time.Millisecond)
        
        var m MessageRow // Instância da struct - Funciona quase da mesma forma que criar um Objeto em Java

        // A query .sql deve terminar com "WHERE m._id > ?" para que o Scan funcione corretamente no loop de novas mensagens.
        err := database.QueryRow(querysqlQueryMessages, lastProcessID).Scan(
            &m.ID, 
            &m.Contexto, 
            &m.NomeConversa, 
            &m.Remetente, 
            &m.Conteudo, 
            &m.Horario,
        )

        if err != nil || m.ID <= lastProcessID { break }

        lastProcessID = m.ID

        // Helper para tratar os NullStrings
        format := func(ns sql.NullString) string {
            if ns.Valid { return ns.String }
            return "NULO"
        }

        log.Printf("[%s] %s: %s (via %s) em %s", 
            format(m.Contexto), 
            format(m.Remetente), 
            format(m.Conteudo), 
            format(m.NomeConversa), 
            format(m.Horario),
        )
    }
}