package main

import (
	"context" // Context é uma ferramenta que controla o tempo de vida de operações interligadas, como conexões de banco de dados ou chamadas de API. Ele é usado para cancelar operações ou definir prazos. Será aliado nesse projeto para reduzir consumo de recursos do Android (memória, processamento e bateria), principalmente ao finalizar ou falhar operações.
	"database/sql"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite" // O Golang não executa o código se algum import não for usado, por isso, esse import foi "silenciado" (usando "_ " no começo) para forçar o compilador ignorar a regar nessa linha.
)

var ctx = context.Background() 

func main() {
	fmt.Println("Iniciando Agente Nativo no Android...")

	// 1. Testando conexão com o SQLite
	// O caminho para o banco de dados é /data/data/com.whatsapp/databases/msgstore.db
	// Go consegue atribuir o retorno de uma função a mais de uma variável, por isso, "db" recebe a conexão e "err" recebe o erro (se houver)
	db, err := sql.Open("sqlite", "/data/data/com.whatsapp/databases/msgstore.db")
	if err != nil {
		log.Fatal("Erro ao abrir SQLite:", err)
	} else {
		fmt.Println("Conexão com SQLite estabelecida!")
	}
	defer db.Close()

	// 2. Testando conexão com o Redis na máquina local
	// 10.0.2.2 é o IP apontado para o Host dentro do emulador
	rdb := redis.NewClient(&redis.Options{ // Cria um cliente Redis com as opções fornecidas
		Addr: "10.0.2.2:6379", // o Container docker está servindo Redis na porta 6379 do IP 127.0.0.1 (localhost), mas dentro do emulador, o IP para acessar o Host é 10.0.2.2
	})

	err = rdb.Set(ctx, "agente_status", "online", 0).Err() // Tenta criar a chave "agente_status" com o valor "online" e sem expiração. Se houver um erro, ele será atribuído à variável "err".
	if err != nil {
		fmt.Println("Aviso: Redis não alcançado, mas o código compilou!")
	} else {
		fmt.Println("Conexão com Redis estabelecida!")
	}
}