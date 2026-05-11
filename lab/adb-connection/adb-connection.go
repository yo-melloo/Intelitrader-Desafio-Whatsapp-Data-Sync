/*
* Esse código é um agente semelhante ao WhatsAppSync/main.go,
* aqui a conexão com o SQLite vai tentar mirar diretamente no banco de dados do WhatsApp do android e exibir as mensgens (brutas) no console,
* a itenção é identificar e mitigar os desafios de acesso ao ambiente do Android e o acesso ao banco de dados do WhatsApp,
* Condição de sucesso: o console exibir as mensagens que começarão a chegar no whatsapp após a inicialização, sem perdas.
* Definição de Sucesso -> Conclui Etapa de Implementação de leitura recursiva do banco de dados -> Integrar em WhatsAppSync/main.go -> Iniciar integração com Redis (Proxima etapa) + modelagem de dados
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
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "modernc.org/sqlite"
)

// Setup global

var dbDir = "/data/data/com.whatsapp/databases/"
var databasePath string = dbDir + "msgstore.db"
var databaseWalPath = databasePath + "-wal"
var lastProcessID int
var message string
var timer *time.Timer


func main() {

	/* Processo: <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	* 0. Estabelece proteção contra OOM Killer
	* 1. Estabelece conexão com banco de dados
	* 2. Insere Watcher para monitorar o .db-wal do WhatsApp, com flags para trabalhar conforme Write-Back Cache 
	* 3. Calibra o lastProcessID para o ID mais recente no banco de dados, garantindo que apenas novas mensagens sejam processadas
	* 4. Goroutine que monitora o banco de dados do WhatsApp
	* 5. Loga as mensagens novas que chegarem no WhatsApp após a inicialização do agente
	*
	>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>*/ 

	// 0. Proteção contra o OOM Killer
	// Certifique-se de que o binário seja executado via su ou por um script de boot do Magisk
    pid := os.Getpid()
    scorePath := fmt.Sprintf("/proc/%d/oom_score_adj", pid)

	err := os.WriteFile(scorePath, []byte("-1000"), 0644)
    if err != nil {
        log.Printf("[NATIVE AGENT] Aviso: Não foi possível elevar prioridade: %v", err)
        // Isso pode falhar se não houver root, mas o agente continuará rodando
    } else {
        log.Println("[NATIVE AGENT] Prioridade de processo definida para máxima (-1000)")
    }

	// 1. Conexão com Banco de dados:
	dsn := databasePath + "?_pragma=journal_mode(WAL)&_pragma=synchronous=NORMAL&_pragma=query_only=1" 	// o uso de ? permite adicionar "flags" que configuram o acesso ao banco de dados
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	log.Println("[SQLITE3 DRIVER] Conexão com o banco de dados estabelecida com sucesso!")

	// 2. Inserir Watcher no .db-wal:
	watcher, err := fsnotify.NewWatcher()	// Instancia novo Watcher
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Println("[WATCHER] Inicializado com sucesso!")

	if err := watcher.Add(dbDir); err != nil {	// Adiciona Watcher ao diretório
		log.Fatal(err)
	}
	log.Printf("[WATCHER] Watcher adicionado ao diretório: \"%s\"\n", dbDir)

	// 3. Calibra o lastProcessID para o ID:
	err = database.QueryRow("SELECT _id, text_data FROM message ORDER BY _id DESC LIMIT 1").Scan(&lastProcessID, &message)
	if err != nil {
		log.Fatal("[SQLITE3 DRIVER] Erro ao calibrar ID inicial: ", err)
	}
	log.Printf("[SQLITE3 DRIVER] lastProcessID calibrado para: %d\n", lastProcessID)
	fmt.Printf("[WATCHER] A útlima mensagem foi: %s\n", message)

	// 4. Goroutine que monitora o banco de dados do WhatsApp:
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//log.Printf("[DEBUG EVENTO] %s: %s", event.Op.String(), event.Name)
				
				// 5. Log de mensagens do WhatsApp:
				if event.Name == databaseWalPath || event.Name == databasePath {
					pullMessage(database)
				}

			case <-ticker.C:
				pullMessage(database)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("[WATCHER ERROR]", err)
			}
		}
	}()

	<-make(chan struct{})
}

// Lê o banco de dados e retorna a útlima mensagem inserida
// Retorna registro com pouca modelagem - TODO: MODELAR DADOS PARA EXIBIÇÃO ÍNTEGRA DE MENSAGENS
func pullMessage(database *sql.DB) {
	if timer != nil {
		timer.Stop()
	}

	timer = time.AfterFunc(100*time.Millisecond, func() {
		var id int
		var msg sql.NullString

		err := database.QueryRow("SELECT _id, text_data FROM message WHERE _id > ? ORDER BY _id ASC LIMIT 1", lastProcessID).Scan(&id, &msg)
		if err == nil {
			lastProcessID = id
			texto := "[Mensagem sem texto/Mídia]"
			if msg.Valid {
				texto = msg.String
			}
			log.Printf("[WATCHER] NOVA MENSAGEM: %s [ID: %d]", texto, id)
			pullMessage(database)
		}
	})
}