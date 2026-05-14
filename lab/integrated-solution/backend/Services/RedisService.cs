using StackExchange.Redis; // Usando a biblioteca StackExchange.Redis para conexão com o Redis
using Backend.Models; // Importando os modelos de Mensagem e Contato
using System.Text.Json; // Usando a biblioteca System.Text.Json para serialização e desserialização de JSON

namespace Backend.Services;

/**
 * RedisService - Camada de Serviços
 * 
 * Este serviço gerencia a conexão com o Redis.
 * 
 * Em Java: Seria um @Service do Spring.
 * Em Python: Seria um módulo ou uma classe Singleton.
 */
public class RedisService
{
    private readonly ConnectionMultiplexer _redis; // Faz a conexão com o Redis
    private readonly IDatabase _db; // Banco de dados
    private readonly ISubscriber _subscriber; // Subscriber para mensagens em tempo real

    public RedisService(IConfiguration config)
    {
        // Conecta ao Redis usando o IP do host (10.0.2.2 é o host visto do emulador, mas aqui usamos localhost)
        // config é a variável que contém as configurações da aplicação, como a string de conexão com o Redis.
        // O operador ?? é um operador de coalescência (verifica valores nulos), que significa "se o valor da esquerda for nulo, use o valor da direita". -> usado em Fallbaks
        // Aqui, caso não seja encontrado o host no arquivo de configuração, use localhost:6379
        string connectionString = config.GetSection("Redis:ConnectionString").Value ?? "localhost:6379";
        
        // Multiplexer é o objeto central que gerencia as conexões de forma eficiente.
        // o "_" antes do nome da variável indica que ela é privada e somente acessível dentro desta classe.
        _redis = ConnectionMultiplexer.Connect(connectionString); // Conecta ao Redis
        _db = _redis.GetDatabase(); // Pega o banco de dados
        _subscriber = _redis.GetSubscriber(); // Pega o subscriber

        if(_redis.IsConnected)
        {
            Console.WriteLine("\n--------------------------------------------------\n[REDIS INFO] ✅ Conexão estabelecida com sucesso!\n--------------------------------------------------\n");
        }
        else
        {
            Console.WriteLine("\n--------------------------------------------------\n[REDIS ERRO] ❌ Falha crítica ao conectar ao servidor Redis.\n--------------------------------------------------\n");
        }

    }

    /**
     * Busca uma mensagem específica pelo ID no Redis (Hash).
     */
    public async Task<MessageDto?> GetMessageAsync(string id) // o '?' indica que pode retornar nulo se não retornar o DTO
    {
        // HGetAllAsync busca todos os campos do Hash no Redis. -> como no HGetAll do Go
        var fields = await _db.HashGetAllAsync(id);
        
        if (fields.Length == 0) return null;

        // Mapeamos os campos do Redis para o nosso objeto C#.
        // Dictionary<string, string> em C# é como Map<String, String> em Java.
        // x => x.Name.ToString() -> lambda que recebe um parâmetro x e retorna o nome do campo.
        // ToDictionary -> Transforma um array de objetos em um dicionário.
        // Dicionário é uma das alternativas para serializar campos dinâmicos.
        // No lugar do dicionário, poderíamos criar um objeto Message.
        var dict = fields.ToDictionary(x => x.Name.ToString(), x => x.Value.ToString());

        return new MessageDto
        {
            Id = id,
            Contexto = dict.GetValueOrDefault("Contexto"),
            NomeConversa = dict.GetValueOrDefault("NomeConversa"),
            Nome_conversa = dict.GetValueOrDefault("NomeConversa"),
            Remetente = dict.GetValueOrDefault("Remetente"),
            Conteudo = dict.GetValueOrDefault("Conteudo"),
            Horario = dict.GetValueOrDefault("Horario"),
            ReceivedAt = DateTime.UtcNow.ToString("O")
        };
    }

    /**
     * Inscreve-se no canal de atualizações do Redis.
     * Quando o Agente Go publica no canal, esta função é disparada.
     */
    public Action<RedisChannel, RedisValue> Subscribe(Action<string> onMessageReceived)
    {
        Action<RedisChannel, RedisValue> handler = (channel, message) => {
            onMessageReceived(message.ToString());
        };
        _subscriber.Subscribe("general:hash-updates", handler);
        return handler;
    }

    /**
     * Remove a subscrição de um canal.
     */
    public void Unsubscribe(string channel, Action<RedisChannel, RedisValue> handler)
    {
        _subscriber.Unsubscribe(channel, handler);
    }

    /**
     * Publica um comando para o Agente Android inserir um contato.
     * Envia para o canal 'contacts:insert' conforme especificado.
     */
    public async Task PublishContactAsync(object contact)
    {
        try
        {
            var json = JsonSerializer.Serialize(contact);
            await _subscriber.PublishAsync("contacts:insert", json);
            Console.WriteLine("\n--------------------------------------------------\n[REDIS INFO] ✅ Ordem de injeção de contato enviada à fila!\n--------------------------------------------------\n");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"\n--------------------------------------------------\n[REDIS ERRO] ❌ Falha ao publicar ordem de injeção: {ex.Message}\n--------------------------------------------------\n");
            throw; // Re-lança a exceção para que o endpoint HTTP falhe (Erro 500)
        }
    }



    /**
     * Busca as últimas mensagens usando SCAN (para não travar o Redis com KEYS).
     */
    public async Task<List<MessageDto>> GetRecentMessagesAsync()
    {
        var messages = new List<MessageDto>();
        
        try 
        {
            var endpoints = _redis.GetEndPoints();
            Console.WriteLine($"\n--------------------------------------------------\n[REDIS INFO] Varredura de endpoints iniciada: {string.Join(", ", endpoints.Select(e => e.ToString()))}");
            
            var server = _redis.GetServer(endpoints[0]);
            var keys = server.Keys(pattern: "msg:*").Take(20).ToList();
            Console.WriteLine($"[REDIS INFO] Histórico resgatado: {keys.Count} mensagens\n--------------------------------------------------\n");

            foreach (var key in keys)
            {
                var msg = await GetMessageAsync(key!);
                if (msg != null) messages.Add(msg);
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"\n--------------------------------------------------\n[REDIS ERRO] ❌ Falha ao buscar histórico de mensagens: {ex.Message}\n--------------------------------------------------\n");
        }

        return messages;
    }
}
