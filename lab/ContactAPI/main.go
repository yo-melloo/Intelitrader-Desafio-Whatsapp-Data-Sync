/*
* Este é um laboratório para testar a API de contatos do Android.
* Ele simula parte do Agente que recebe comandos do Redis para adicionar contatos na agenda do Android.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

func main() {

	
	/*
	* Processo:
	* Frontend -> Usuário preenche nome e número de telefone e envia ordem para API do Backend
	* Backend -> Valida os dados, e salva no Redis (Publica mensagem no canal "contacts:add")
	* Redis -> Publica mensagem no canal "contacts:add"
	* Agente -> Recebe mensagem e dispara contato na agenda
	* 
	* 1 - Conecta com o Redis
	* 2 - Se inscreve no canal "contacts:add"
	* 3 - Aguarda receber mensagem
	* 4 - Função insere contato na agenda dispara quando recebe mensagem
	*/

	fmt.Println("[API] Agente Online!")

	// 1. Conecta com o Redis
	redisClient = redisConnect()

	// 2. Se inscreve no canal "contacts:add"
	redisSubscriber()

}

func redisConnect() *redis.Client{
	
	client := redis.NewClient(&redis.Options{
		Addr: "10.0.2.2:6379",
		Password: "",
		DB: 0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Println("[Redis] Erro ao conectar com o Redis", err)
		return nil
	}

	fmt.Println("[Redis] Conectado com o Redis")
	return client

}

func redisSubscriber(){

	// Define o canal de inscrição
	channelSubscribe := "contacts:insert"

	// Inscreve o cliente no canal definido acima
	pSub := redisClient.Subscribe(ctx, channelSubscribe)
	fmt.Printf("[Redis Subscriber] Subscriber inscrito no canal %s\n", channelSubscribe)
	defer pSub.Close()

	for msg := range pSub.Channel() {	// <- Contato em String
		fmt.Println("[Redis Subscriber] Mensagem recebida:", msg.Payload)
		
		contactShoot(msg.Payload)
	}

}

func getActiveAccount() (name string, accType string) {
	// Executa dumpsys para listar contas no dispositivo
	cmd := exec.Command("dumpsys", "account")
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	// Regex para capturar: name=email@gmail.com, type=com.google
	re := regexp.MustCompile(`name=([^, ]+), type=(com\.google)`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", "" // Retorna vazio se não encontrar conta Google
}

func contactShoot(contactJSON string) {

	/*
	* 
	* Processo de injeção do contato
	* O Android armazena os dados dos contatos em banco de dados SQLite, 
	* por isso para adicionar, editar ou remover contatos, é necessário interagir com o banco de dados.
	* 
	* 1 - Desserializa a String JSON recebida -> melhor forma de fazer isso é com dicionários
	* 1 - Localiza conta do usuário (Whatsapp, SIM Card ou Conta Google)
	* 2 - Define se o contato será salvo na Conta Google, SIM ou Armazenamento do Telefone
	* 3 - Insere uma entrada na tabela raw_contacts -> Inicia registro do contato para gerar ID do objeto
	* 4 - Obtém o ID do registro do contato
	* 5 - Preenche e salva os dados do contato
	* 
	*/

	fmt.Println("[CONTACT BUILDER] Iniciando injeção do contato")

	// Cria dicionário para receber os dados do contato
	var contactDto ContactDto 

	err := json.Unmarshal([]byte(contactJSON), &contactDto)

	if err != nil {
		fmt.Println("[CONTACT BUILDER] Erro ao desserializar JSON", err)
		return
	}

	if contactDto.Name == "" || contactDto.Number == "" {
		fmt.Println("[CONTACT BUILDER] Nome ou número inválido ou vazio")
		return
	}

	fmt.Println("[CONTACT BUILDER] Contato recebido:")
	fmt.Println("Nome: ", contactDto.Name)
	fmt.Println("Número: ", contactDto.Number)

	/*
	* Estrutura de um os.exec.Command:
	* ---------------------------------
	* 
	*  cmd := exec.Command("echo", "Hello World") -> exec.Command("comando", "argumentos")
	*
	*  err = cmd.Run() ->  Garante que o processo seja executado e aguarda até que o comando termine.
	*  if err != nil {
	*	 fmt.Println("Erro ao executar comando", err)
	*	 return
	*  }
	*/

	// 1. Descobre Conta Ativa - Evita delay ou dessincronização
	accName, accType := getActiveAccount()
	bindAccName := "account_name:n:"
	bindAccType := "account_type:n:"

	if accName != "" {
		fmt.Printf("[CONTACT BUILDER] Operando no modo Conta Google: %s\n", accName)
		bindAccName = "account_name:s:" + accName
		bindAccType = "account_type:s:" + accType
	} else {
		fmt.Println("[CONTACT BUILDER] Operando no modo Local (Nenhuma conta Google ativa)")
	}

	// 2. PRIMEIRO: Criar o Raw Contact (Plantar a semente)
	fmt.Println("[CONTACT BUILDER] Criando registro bruto...")
	cmdInsertRaw := exec.Command("content", "insert", "--uri", "content://com.android.contacts/raw_contacts", 
		"--bind", bindAccName, 
		"--bind", bindAccType)
	
	err = cmdInsertRaw.Run()
	if err != nil {
		fmt.Println("[CONTACT BUILDER] Erro ao criar raw_contact:", err)
		return
	}
	
	// 3. SEGUNDO: Recuperar o ID do registro que ACABAMOS de criar
	// Removemos o --limit. O --sort _id DESC garante que o mais novo venha primeiro.
	fmt.Println("[CONTACT BUILDER] Recuperando ID gerado...")
	cmdGetID := exec.Command("content", "query", "--uri", "content://com.android.contacts/raw_contacts", "--projection", "_id", "--sort", "_id DESC")
	output, err := cmdGetID.Output()
	if err != nil {
		fmt.Println("[CONTACT BUILDER] Erro ao consultar banco:", err)
		return
	}
	
	// Regex captura a primeira ocorrência de _id=X (que será o maior ID devido ao SORT DESC)
	re := regexp.MustCompile(`_id=(\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		fmt.Println("[CONTACT BUILDER] Erro: ID não encontrado no output")
		return
	}
	rawID := matches[1]
	fmt.Printf("[CONTACT BUILDER] ID Vinculado: %s para %s\n", rawID, contactDto.Name)
	// Consideração: Eu não sei trabalhar com Regex, então gerei isso com IA (Filtrar ID do Raw Contact).
	
	// 4. TERCEIRO: Vincular Nome e Telefone ao ID encontrado
	fmt.Println("[CONTACT BUILDER] Vinculando metadados...")

	exec.Command("content", "insert", "--uri", "content://com.android.contacts/data", 
		"--bind", "raw_contact_id:i:"+rawID, 
		"--bind", "mimetype:s:vnd.android.cursor.item/name", 
		"--bind", "data1:s:"+contactDto.Name).Run()

	exec.Command("content", "insert", "--uri", "content://com.android.contacts/data", 
		"--bind", "raw_contact_id:i:"+rawID, 
		"--bind", "mimetype:s:vnd.android.cursor.item/phone_v2", 
		"--bind", "data1:s:"+contactDto.Number, 
		"--bind", "data2:i:2").Run()

	fmt.Println("[CONTACT BUILDER] Contato Construído com sucesso!")
}

type ContactDto struct {
	Name string `json:"name"`
	Number string `json:"number"`
}
