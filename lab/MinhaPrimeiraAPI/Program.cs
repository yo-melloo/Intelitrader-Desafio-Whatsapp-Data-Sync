/*
* @autor: Gustavo Melo
* @date: 2026-05-13
* @description: Primeira API em C# usando Minimal API.
* Objetivo: Criar uma API que retorne uma lista de chats.
*/

// NÃO É PRECISO USAR O USING PARA TRAZER O MODELO DO CHAT, PORQUE ELE ESTÁ NO MESMO NAMESPACE
//using Chat;

using StackExchange.Redis;
using Microsoft.Extensions.Caching.Distributed;

namespace MinhaPrimeiraAPI
{

    public class Program
    {
        public static void Main(string[] args)
        {

            /*
             * Processo:
             * 1. Criar um builder. Esse builder é responsável por montar a aplicação. --> WebApplication.CreateBuilder(args);
             * 2. Adicionar serviços ao builder. São componentes que auxiliam a aplicação, como o CORS e o Redis. --> builder.Services.AddCors(options => { ... }) e builder.Services.AddStackExchangeRedisCache(...);
             * 3. Montar a aplicação. --> var app = builder.Build();
             * 4. Configurar os serviços na aplicação. --> app.UseCors();
             * 5. Criar rotas na aplicação. --> app.MapGet("/", () => "Olá Mundo Dotnet Minimal API!");
             * 6. Adicionar subscriber do Redis e imprimir mensagens no terminal --> builder.Services.AddHostedService<RedisSubscriberWorker>();
             * 7. Simular endpoint POST /contacts do desafio --> app.MapPost("/contacts", async (Microsoft.Extensions.Caching.Distributed.IDistributedCache cache, Contact contact) => { ... });
             * 8. Iniciar a aplicação --> app.Run();
             * 
             * 
             * Para ler mais tarde: https://learn.microsoft.com/pt-br/aspnet/core/fundamentals/minimal-apis/?view=aspnetcore-9.0
             */

            // 1. Criar um builder.
            var builder = WebApplication.CreateBuilder(args);

            // 2. Adicionar serviços ao builder.
            // Serviço do CORS para permitir requisições externas
            builder.Services.AddCors(options => {
                options.AddDefaultPolicy(policy => {
                    policy.AllowAnyOrigin().AllowAnyHeader().AllowAnyMethod();
                });
            });

            // Serviço do Redis no localhost, com prefixo "Teste-API:" em todas as chaves.
            var redisConnectionString = "localhost:52713";
            builder.Services.AddStackExchangeRedisCache(options => {
                options.Configuration = redisConnectionString;
                options.InstanceName = "Teste-API:"; // o "Teste-API:" é prefixo das chaves do Redis
            });

            builder.Services.AddSingleton<IConnectionMultiplexer>(ConnectionMultiplexer.Connect(redisConnectionString));

            // --> 6. Adicionar subscriber do Redis e imprimir mensagens no terminal
            builder.Services.AddHostedService<RedisSubscriberWorker>();

            // 3. Montar a aplicação.
            var app = builder.Build();

            // 4. Configurar os serviços na aplicação.
            app.UseCors();

            // 5. Criar rotas na aplicação.
            // rota padrão, só para verificar se a API está funcionando.
            app.MapGet("/", () => "Olá Mundo Dotnet Minimal API!");

            //rota para retornar uma lista de chats.
            // chama uma função anônima lambda -> `() => ...`
            app.MapGet("/chats", () => new List<Chat> 
            {
                new Chat 
                {
                    Nome = "Carlos",
                    Telefone = "1234567890",
                    Foto = "carlos.jpg",
                    UltimaMensagem = DateTime.Now,
                    Mensagem = "Oi",
                    EstaEscrito = false,
                    EstaOnline = true
                },
                new Chat 
                {
                    Nome = "Antonia",
                    Telefone = "1234567890",
                    Foto = "antonia.jpg",
                    UltimaMensagem = DateTime.Now,
                    Mensagem = "Oi",
                    EstaEscrito = false,
                    EstaOnline = true
                },
                new Chat 
                {
                    Nome = "Pedro",
                    Telefone = "1234567890",
                    Foto = "pedro.jpg",
                    UltimaMensagem = DateTime.Now,
                    Mensagem = "Oi",
                    EstaEscrito = false,
                    EstaOnline = true
                }   
            });

            app.MapGet("/chat/{chave}", async (Microsoft.Extensions.Caching.Distributed.IDistributedCache cache, string chave) => {
                
                // 1. Busca o valor no cache
                var valor = await cache.GetStringAsync(chave);

                // 2. Caso o valor não existir, retorna 404
                if (valor == null) {
                    return Results.NotFound($"Não foi possível encontrar o chat com a chave: {chave}");
                }

                // 3. Caso o valor existir, retorna o valor deserializado para Chat
                return Results.Ok(System.Text.Json.JsonSerializer.Deserialize<Chat>(valor));

            });

            // O parâmetro 'IDistributedCache cache' serve para injetar o serviço do Redis na rota (injeção de dependência)
            // as rotas de Minimal API usam Injeção de Dependência automática.
            // Se tiver um parâmetro na rota, o builder vai buscar esse serviço na memória e entregar aqui.
            app.MapPost("/chats", async (Microsoft.Extensions.Caching.Distributed.IDistributedCache cache, IConnectionMultiplexer redis) => {
                
                // 1. Gera uma chave simples usando string comum (o IDistributedCache não usa RedisKey)
                var chave = DateTime.Now.Minute.ToString() + DateTime.Now.Second.ToString();
                
                var chat = new Chat {
                    Nome = "Antonia",
                    Telefone = "9999999999",
                    Foto = "antonia.jpg",
                    UltimaMensagem = DateTime.Now,
                    Mensagem = "Oi",
                    EstaEscrito = false,
                    EstaOnline = true
                };

                // 2. Transforma o objeto em string JSON (usando a lib System.Text.Json)
                var valor = System.Text.Json.JsonSerializer.Serialize(chat);

                // 3. Grava no Redis de forma assíncrona usando o cache injetado
                await cache.SetStringAsync(chave, valor);

                var publisher = redis.GetSubscriber();

                // 4. Publica no Redis no canal 'chat-mensagens'
                await publisher.PublishAsync(RedisChannel.Literal("chat-mensagens"), valor);

                // Retorna o status HTTP 201 Created corretamente
                return Results.Created($"/chats/{chave}", chat); 
            });

            // 7. Simular endpoint POST /contacts do desafio 
            // O parâmetro 'Contact contact' já aceita JSONs enviados pelo Body da requisição, 
            // pois o 'contact' está como parâmetro da rota.
            // O builder já faz a desserialização automática.
            app.MapPost("/contacts", async (Microsoft.Extensions.Caching.Distributed.IDistributedCache cache, Contact contact) => {
                
                // 1. Serializa para String
                var valor = System.Text.Json.JsonSerializer.Serialize(contact);

                // 2. Cria um id aleatório:
                // "contact:" é o namespace em que o contato vai ser salvo no banco.
                // DateTime.Now.Minute.ToString() + DateTime.Now.Second.ToString() -> gera um id aleatório usando minuto e segundo atual.
                var id = "contact:" + DateTime.Now.Minute.ToString() + DateTime.Now.Second.ToString();

                // 3. Grava no Redis de forma assíncrona usando o cache injetado    
                await cache.SetStringAsync(id, valor);
                return Results.Created($"Criado: {id}: {contact}", contact);
           
            });

            // Testa GET no contato criado:
            app.MapGet("/contact/{id}", async (Microsoft.Extensions.Caching.Distributed.IDistributedCache cache, string id) => {
                
                // 1. Busca o valor no cache
                var valor = await cache.GetStringAsync("contact:" + id);

                // 2. Caso o valor não existir, retorna 404
                if (valor == null) {
                    return Results.NotFound($"Contato {id} não encontrado");
                }

                // 3. Transforma o objeto em string JSON (usando a lib System.Text.Json)
                var contact = System.Text.Json.JsonSerializer.Deserialize<Contact>(valor);

                // Retorna o status HTTP 200 OK corretamente
                return Results.Ok(contact);
            });
            
            // 8. Iniciar a aplicação.
            app.Run();
        }
    }
}