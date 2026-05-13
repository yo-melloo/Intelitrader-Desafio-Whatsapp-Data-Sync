/*
*   Este arquivo é a API REST.
*   Ele é responsável por receber comandos do Dashboard e enviar para o Redis.
*   Também é responsável por receber mensagens do Redis e enviar para o Dashboard.
*   Foi gerado com Inteligência Artificial, mas devidamente revisado, documentado e testado.
*/

using Backend.Services; // Importa camada de Serviços
using Backend.Models; // Importa camada de Modelos
using System.Text.Json; // Importa biblioteca de JSON

var builder = WebApplication.CreateBuilder(args); // Cria o Builder da aplicação

// - Configuração do CORS (Cross-Origin Resource Sharing)
// - Permite que o Next.js (rodando em outra porta) acesse esta API.
builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy.AllowAnyOrigin().AllowAnyHeader().AllowAnyMethod();
    });
});

// - Registra o RedisService no sistema de Injeção de Dependência.
// - Em Java Spring: Seria como anotar com @Service e deixar o Spring instanciar.
builder.Services.AddSingleton<RedisService>();

// Monta a API (app):
var app = builder.Build();

app.UseCors(); // -> Cors já vem como parâmetro de app.Build(), aqui explicitamos ele

/**
 * Endpoint: GET /api/messages
 * Retorna as mensagens recentes.
 */
app.MapGet("/api/messages", async (RedisService redis) =>
{
    var messages = await redis.GetRecentMessagesAsync();     // await -> Retorna um array de mensagens
    return Results.Ok(messages);                             // Results.Ok(messages) -> Retorna um objeto com o array de mensagens | status 200 OK
});

app.MapGet("/api/health", () => Results.Ok(new { status = "ok" }));

/**
 * Endpoint: GET /api/stream
 * Implementa Server-Sent Events (SSE) para enviar dados em tempo real.
 * SSEs são mensagens que o servidor envia para o cliente em tempo real.
 * Em vez de o cliente perguntar "tem coisa nova?" a cada segundo (polling), 
 * o servidor "empurra" a mensagem assim que ela chega.
 * 
 * Seria como usar WebSocket, mas apenas em "uma direção", sem configurar o servidor do websocket na API REST
 * 
 */
app.MapGet("/api/stream", async (HttpContext context, RedisService redis) =>
{
    // Define o cabeçalho para SSE
    context.Response.Headers.Add("Content-Type", "text/event-stream");  // -> Define que o conteúdo é um stream de eventos em texto
    context.Response.Headers.Add("Cache-Control", "no-cache");           // -> Cache desabilitado para sempre receber atualizações
    context.Response.Headers.Add("Connection", "keep-alive");           // -> Mantém a conexão aberta

    var responseStream = context.Response.Body; // -> Pega o corpo da resposta

    // O minimal API (ASP.NET Core) consegue saber quando o cliente se desconectou para economizar recursos do servidor
    // Token de cancelamento para saber quando o browser desconectou
    var cts = CancellationTokenSource.CreateLinkedTokenSource(context.RequestAborted);

    // Função que será chamada quando o Redis receber uma nova mensagem
    // Actions são funções anônimas, que recebem um parâmetro (string) e não retornam valor
    // string [tipo] e messageId [nome da variável] -> nessa linha, recebendo um valor string ao ser chamada:
    Action<string> handler = async (messageId) => { 
        try 
        {
            // Se o cliente desconectou, não envia mais mensagens
            if (cts.IsCancellationRequested) return;

            // Pega a mensagem pelo ID que vem do Redis
            var message = await redis.GetMessageAsync(messageId);
            if (message != null) // Verifica sempre se a mensagem existe
            {
                // Configura a serialização para usar camelCase (mesmo padrão da API REST)
                // Sem isso, o .NET enviaria PascalCase (Id, Conteudo) e o Frontend não encontraria os campos.
                var options = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.CamelCase };
                var jsonMessage = JsonSerializer.Serialize(message, options); 
                
                // Envia a mensagem para o cliente (painel front-end)
                await responseStream.WriteAsync(System.Text.Encoding.UTF8.GetBytes($"data: {jsonMessage}\n\n"), cts.Token); // Codifica para UTF-8 e envia a mensagem para o cliente
                await responseStream.FlushAsync(cts.Token); // Força o envio da mensagem se estiver em buffer
            }
        }

        catch(OperationCanceledException ex)
        {
            // Cliente desconectou, sai da função sem erro
            Console.WriteLine($"Operação cancelada: {ex.Message}");
            return;
        }

        catch (Exception ex)
        {
            Console.WriteLine($"Erro ao enviar SSE: {ex.Message}");
        }
    };

    redis.Subscribe(handler); // Inicia a subscrição no canal de atualizações do Redis

    // - Keep-Alive (Heartbeat): envia um "ping" silencioso a cada 15 segundos
    // - Sem isso, o Windows/Nginx encerra conexões inativas com timeout (erro 499/504)
    // - O protocolo SSE ignora linhas que começam com ":", são tratadas como comentários
    _ = Task.Run(async () => {
        while (!cts.Token.IsCancellationRequested)
        {
            try {
                await Task.Delay(15000, cts.Token);
                await responseStream.WriteAsync(System.Text.Encoding.UTF8.GetBytes(":\n\n"), cts.Token); // -> Comentário SSE (ignorado pelo browser, mas mantém a conexão)
                await responseStream.FlushAsync(cts.Token);
            } catch { break; } // -> Sai silenciosamente quando o processo encerrar
        }
    });

    // Mantém a conexão ativa até o processo ser encerrado ou o browser desconectar
    // Timeout.Infinite significa que nunca vai expirar por conta própria
    await Task.Delay(Timeout.Infinite, context.RequestAborted);
});

/**
 * Endpoint: POST /api/contacts
 * Recebe um contato e publica no Redis para o Agente Android.
 */
app.MapPost("/api/contacts", async (ContactDto contact, RedisService redis) =>
{
    if (string.IsNullOrEmpty(contact.Name) || string.IsNullOrEmpty(contact.Number))
    {
        return Results.BadRequest("Nome e número são obrigatórios.");
    }

    await redis.PublishContactAsync(contact);
    return Results.Accepted(); // 202 Accepted: O comando foi enviado para processamento.
});

app.Run();

// DTO para receber o contato
public record ContactDto(string Name, string Number); // Bem mais simples que criar uma classe separada!
