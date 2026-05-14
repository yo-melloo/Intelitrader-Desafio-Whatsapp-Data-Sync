/*
 * API REST de Sincronização
 * Gerencia a comunicação entre o Dashboard (Frontend) e o Redis (Backend/Agente).
 */

using Backend.Services;
using Backend.Models;
using System.Text.Json;

var builder = WebApplication.CreateBuilder(args); // Cria o Builder da aplicação

// Configuração do CORS para comunicação com o Frontend
builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy.AllowAnyOrigin().AllowAnyHeader().AllowAnyMethod();
    });
});

// Registra o RedisService no sistema de Injeção de Dependência
builder.Services.AddSingleton<RedisService>();

// Monta a API (app):
var app = builder.Build();

app.UseCors();

/**
 * Endpoint: GET /api/messages
 * Retorna o histórico de mensagens recentes do Redis.
 */
app.MapGet("/api/messages", async (RedisService redis) =>
{
    var messages = await redis.GetRecentMessagesAsync();
    return Results.Ok(messages);
});

app.MapGet("/api/health", () => Results.Ok(new { status = "ok" }));

/**
 * Endpoint: GET /api/stream
 * Server-Sent Events (SSE) para streaming de mensagens em tempo real.
 */
app.MapGet("/api/stream", async (HttpContext context, RedisService redis) =>
{
    context.Response.Headers.Add("Content-Type", "text/event-stream");
    context.Response.Headers.Add("Cache-Control", "no-cache");
    context.Response.Headers.Add("Connection", "keep-alive");

    var responseStream = context.Response.Body;
    var cts = CancellationTokenSource.CreateLinkedTokenSource(context.RequestAborted);

    Action<string> handler = async (messageId) => { 
        try 
        {
            if (cts.IsCancellationRequested) return;

            var message = await redis.GetMessageAsync(messageId);
            if (message != null)
            {
                // Serialização em camelCase para compatibilidade com Frontend
                var options = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.CamelCase };
                var jsonMessage = JsonSerializer.Serialize(message, options); 
                
                await responseStream.WriteAsync(System.Text.Encoding.UTF8.GetBytes($"data: {jsonMessage}\n\n"), cts.Token);
                await responseStream.FlushAsync(cts.Token);
            }
        }

        catch(OperationCanceledException ex)
        {
            // Cliente desconectou, sai da função sem erro
            Console.WriteLine($"\n--------------------------------------------------\n[SSE INFO] Conexão com o painel encerrada: {ex.Message}\n--------------------------------------------------\n");
            return;
        }

        catch (Exception ex)
        {
            Console.WriteLine($"\n--------------------------------------------------\n[SSE ERRO] ❌ Falha crítica no envio do evento (SSE): {ex.Message}\n--------------------------------------------------\n");
        }
    };

    redis.Subscribe(handler);

    // Keep-Alive (Heartbeat) - Evita timeout do Nginx (Erro 499/504)
    _ = Task.Run(async () => {
        while (!cts.Token.IsCancellationRequested)
        {
            try {
                await Task.Delay(15000, cts.Token);
                await responseStream.WriteAsync(System.Text.Encoding.UTF8.GetBytes(":\n\n"), cts.Token);
                await responseStream.FlushAsync(cts.Token);
            } catch { break; }
        }
    });

    await Task.Delay(Timeout.Infinite, context.RequestAborted);
});

/**
 * Endpoint: POST /api/contacts
 * Recebe um contato do Dashboard e publica no Redis para o Agente Android.
 */
app.MapPost("/api/contacts", async (ContactDto contact, RedisService redis) =>
{
    if (string.IsNullOrEmpty(contact.Name) || string.IsNullOrEmpty(contact.Number))
    {
        return Results.BadRequest("Nome e número são obrigatórios.");
    }

    // Validação: Aceita apenas números no campo 'Number'
    if (!System.Text.RegularExpressions.Regex.IsMatch(contact.Number, @"^\d+$"))
    {
        return Results.StatusCode(403); // Status 403 solicitado se não for número
    }

    await redis.PublishContactAsync(contact);
    return Results.Accepted(); // Status 202
});

app.Run();

public record ContactDto(string Name, string Number);
