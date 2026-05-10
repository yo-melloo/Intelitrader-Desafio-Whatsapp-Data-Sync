/*
* Este código é um exemplo do uso de observadores para o monitoramento de mudanças em um banco de dados SQLite com uso de WAL (Write-Ahead Logging).
* Ele utiliza a biblioteca fsnotify para observar o arquivo de log WAL do SQLite e detectar quando novas entradas são escritas.
* Quando uma mudança é detectada, o código consulta o banco de dados para buscar os novos registros a partir do último ID processado,
* exibindo o conteúdo desses registros no console.
* O código também inclui tratamento de erros e logs para facilitar a depuração e monitoramento do processo.

 */

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	_ "modernc.org/sqlite" // usar "github.com/mattn/go-sqlite3" no projeto final por eficiência
)

var lastProcessID int
var databasePath string = ".//db/teste-wal.db"
var databaseWalPath string = databasePath + "-wal"
var timer *time.Timer

func main() {

	log.Println("[NATIVE AGENT] Iniciando o observador de banco de dados...")
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	
	done := make(chan bool)
	
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		log.Fatal(err)
		} else {
			log.Println("[SQLITE3 DRIVER] Conexão com o banco de dados estabelecida com sucesso.")
			watcher.Add(databaseWalPath)
			log.Println("[WATCHER] Watcher adicionado ao arquivo: \"" + databaseWalPath + "\"")
		}
		defer database.Close()
		
		err = database.QueryRow("SELECT COALESCE(MAX(id), 0) FROM example").Scan(&lastProcessID)
		if err != nil {
			log.Fatal("Erro ao calibrar ID inicial:", err)
		}
		
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					//log.Println("[WATCHER] Evento detectado")
					
					if event.Has(fsnotify.Write) {
						if timer != nil {
							timer.Stop() // para timers anteriores, se existirem, para evitar logs repetidos
						}
	
						timer = time.AfterFunc(200*time.Millisecond, func() { // inicia um novo timer para gerar o log após um curto atraso
						buscarUltimaMensagem(database) // o log que exibe o conteúdo do registro encontrado no banco de dados está nessa função
						})
					}
					
					if event.Has(fsnotify.Create) {
						log.Println("[WATCHER] O arquivo foi criado.\n", event.Op)
					}
					
					if event.Has(fsnotify.Remove) {
						log.Println("[WATCHER] O arquivo foi removido.\n", event.Op)
					}
					
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					fmt.Println("[WATCHER] Erro detectado:", err)
				}
			}
			}()
			
			
			<-done
		}
		
		
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