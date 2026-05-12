/*
 * Este código é um crud simples usando Redis
 * Separei as etapas de CRUD por funções, assim o entendimento ficou seccionado
 * Pesquisei sobre pub/sub e vi que o uso não é difícil, então além do crud, eu programei atualizações no canal `general:updates` para CRUD de strings e `general:hash-updates` para CRUD de hashes
 */

package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {

	/*
	* OBJETIIVO:
	* - [x] Aprender CRUD simples com Go + Redis															11/05/26
	* - [x] Aprender usar String Hashes																		11/05/26
	* - [x] Aprender usar Publishers																		11/05/26
	* - [x] Aprender usar Subscribers - Usei o Redis Insight para acompanhar as atualziações 				11/05/26
	* - [x] Criar estrutura de Producers e Consumers														11/05/26
	* - [x] Simular a inserção de uma mensagem no Redis > Publisher anuncia alteração > Reação do Consumer	11/05/26
	*/

	fmt.Println(" -- Olá Redis! -- ")

	ctx := context.Background()
	
	// Conecta no Redis:
	rdb, err := conectarNoRedis()
	if err != nil {panic(err)}
	fmt.Println("[Golang] Conectado ao servidor Redis! (rodando em container docker)")
	defer rdb.Close()	// Programa desconexão ao fim do main()
	
	// Cria consumer:
	consumer := rdb.Subscribe(ctx, "general:hash-updates")
	defer consumer.Close()
	fmt.Println("[Golang] Consumer inscrito no canal \"general:hash-updates\"")
	ch := consumer.Channel()

	final := make(chan bool)

	// Consumer rodando em paralelo
	go func (){
		

		for msg := range ch {
			fmt.Printf("[CONSUMER] Alteração na chave %s detectada!\n", msg.Payload)
			
			dados, _ := rdb.HGetAll(ctx, msg.Payload).Result()
			fmt.Println(dados["Texto"])
		}
		
		
		}()
		
	// Realiza crud com String simples
	//fmt.Println("[Golang] Rodando exemplo crud:")
	//exemploCrudSimples(ctx, rdb)

	// Realiza crud com String Hash
	fmt.Println("[Golang] Rodando exemplo crud com hashes:")
	exemploCrudComHashes(ctx, rdb)
	
	final <- true

}

// Estabelece conexão com o servidor local do Redis
func conectarNoRedis() (*redis.Client, error) {
	
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if rdb == nil {
		panic(rdb)
	}

	return rdb, nil

}

// Adiciona registro ao Redis usando rdb.Set
func adicionarStringNoRedis(ctx context.Context, rdb *redis.Client, keyData string, stringData string) {

	err := rdb.Set(ctx, keyData, stringData, 0).Err()
	if err != nil {
		panic(err)
	}

	fmt.Printf("[Redis set] \"%s\" Adicionado com sucesso!\n", stringData)
		err = rdb.Publish(ctx, "general:updates", keyData).Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Redis Pub/Sub] Publish efetuado com sucesso!")



}

// Lê registro do Redis usando rdb.Get
func lerStringDoRedis(ctx context.Context, rdb *redis.Client, keyData string) {

	stringData, err := rdb.Get(ctx, keyData).Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("[Redis get] Chave: %s | Valor: %s\n", keyData, stringData)

}

// Deleta registro do Redis usando rdb.Del
func deletarStringDoRedis(ctx context.Context, rdb *redis.Client, keyData string){

	err := rdb.Del(ctx, keyData).Err()
	if err != nil {
		panic(err)
	}
	err = rdb.Publish(ctx, "general:updates", keyData).Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Redis Pub/Sub] Publish efetuado com sucesso!")
	fmt.Printf("[Redis del] Chave: %s | Deletada com sucesso!\n", keyData)

}

func exemploCrudSimples(ctx context.Context, rdb *redis.Client) {
	
	// Configura chave e valor do teste:
	chave := "saudação"
	valor := "Olá Redis!"

	// CREATE
	adicionarStringNoRedis(ctx, rdb, chave, valor)
	
	// READ
	lerStringDoRedis(ctx, rdb, chave)

	// UPDATE
	//atualizarStringDoRedis()	// no teste atual, repetir o adicionarStringNoRedis() realiza uma atualização no valor da chave livremente	
	
	// DELETE
	//deletarStringDoRedis(ctx, rdb, chave)

}


func adicionarHashNoRedis(ctx context.Context, rdb *redis.Client, keyData string, hashData map[string]interface{}) {
	// CREATE
	err := rdb.HSet(ctx, keyData, hashData).Err()
	if err != nil {
		panic(err)
	}
	fmt.Printf("[Redis HSet] %s adicionado com sucesso!\n", hashData)
	rdb.Publish(ctx, "general:hash-updates", keyData)
	//fmt.Println("[Redis Pub/Sub] Publish efetuado com sucesso!")
}


func lerHashDoRedis(ctx context.Context, rdb *redis.Client, key string){

	hashData, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("[Redis HGetAll] Chave: %s | Valor: %s\n", key, hashData)

}

func atualizarHashDoRedis(ctx context.Context, rdb *redis.Client, key string, hashDataUpdated map[string]interface{}){
	
	hashDataUpdated = map[string]interface{}{
		"Status":"Lida",
		"timestamp":"11-05-2026T00:10:00Z",
	}

	err := rdb.HSet(ctx, key, hashDataUpdated).Err()
	if err != nil {
		panic(err)
	}

	rdb.Publish(ctx, "general:hash-updates", key)
	//fmt.Println("[Redis Pub/Sub] Publish efetuado com sucesso!")
	fmt.Println("[Redis HSet] Hash atualizado!")

	//lerHashDoRedis(ctx, rdb, key)


}

func deletarHashDoRedis(ctx context.Context, rdb *redis.Client, keyData string){

	err := rdb.Del(ctx, keyData).Err()
	if err != nil {
		panic(err)
	}
	err = rdb.Publish(ctx, "general:hash-updates", keyData).Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("[Redis Pub/Sub] Publish efetuado com sucesso!")
	fmt.Printf("[Redis del] Chave: %s | Deletada com sucesso!\n", keyData)

}


func exemploCrudComHashes(ctx context.Context, rdb *redis.Client){
	
	key := "msg:1002"
	value := map[string]interface{}{
		"Remetente":"Desenvolvedor",
		"Texto":"Estou aprendendo usar hashes no Go...",
		"Status":"Não lida",
		"timestamp":"11-05-2026T00:00:00Z",
	}

	nkey := "msg:1003"
	nvalue := map[string]interface{}{
		"Remetente":"Recrutador",
		"Texto":"Você está sendo avaliado!",
		"Status":"Não lida",
		"timestamp":"11-05-2026T00:00:30Z",
	}

	// CREATE
	adicionarHashNoRedis(ctx, rdb, key, value)
	
	// READ
	lerHashDoRedis(ctx, rdb, key)
	
	// UPDATE
	// Assim como a atualização simples, a etapa de atualização com hashes sobescreveo valor no Redis sempre que executado
	// O diferencial é que, usando Hashes, é possível editar a string diretamente por partes ao invés de replicá-la completamente 
	atualizarHashDoRedis(ctx, rdb, key, value)
	
	// DELETE
	//	literalmente a mesma função de deletarStringDoRedis, por que o método Redis usado para Hashes é o mesmo que o de Strings
	//deletarHashDoRedis(ctx, rdb, key)	
	
	// Simula um segundo registro:
	fmt.Println("[Golang] Segundo registro:")
	adicionarHashNoRedis(ctx, rdb, nkey, nvalue)
	atualizarHashDoRedis(ctx, rdb, nkey, nvalue)

}