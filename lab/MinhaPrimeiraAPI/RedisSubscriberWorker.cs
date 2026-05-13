using Microsoft.Extensions.Hosting;
using StackExchange.Redis;
using System.Threading;
using System.Threading.Tasks;

namespace MinhaPrimeiraAPI
{
    // Processo 6: Adicionar subscriber do Redis e imprimir mensagens no terminal.
    // --> builder.Services.AddHostedService<RedisSubscriberWorker>();
    public class RedisSubscriberWorker : BackgroundService  // Quase como uma Goroutine
    {
        private readonly IConnectionMultiplexer _redis;  // Conexão com o Redis
        private readonly ILogger<RedisSubscriberWorker> _logger;  // Logger para reportar as mensagens

        // Construtor do Worker
        public RedisSubscriberWorker(IConnectionMultiplexer redis, ILogger<RedisSubscriberWorker> logger) 
        {
            _redis = redis;
            _logger = logger;
        }

        // Método que define o que o Worker vai fazer
        // protected override -> método está sobrescrevendo um método de uma classe pai -> public abstract class BackgroundService
        // async Task -> método assíncrono que retorna uma Task <- uma promessa de retorno de valor, neste caso, não retorna valor.
        // CancellationToken stoppingToken -> serve para cancelar a execução do método se o usuário encerrar/cancelar a aplicação.
        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            // Obtém o subscriber do Redis
            var subscriber = _redis.GetSubscriber();
            
            _logger.LogInformation("Iniciando Subscriber do Redis. Escutando canal 'chat-mensagens'...");

            // Se inscreve no canal do Redis
            await subscriber.SubscribeAsync(RedisChannel.Literal("chat-mensagens"), (channel, message) =>
            {
                // Código executado toda vez que chegar uma mensagem nova no canal
                _logger.LogInformation($"[Mensagem Recebida do Redis]: {message}");
                
                // Aqui você pode deserializar e processar o JSON:
                // var chat = System.Text.Json.JsonSerializer.Deserialize<Chat>(message);
            });

            // Mantém o Worker vivo enquanto a aplicação estiver rodando
            while (!stoppingToken.IsCancellationRequested)
            {
                await Task.Delay(1000, stoppingToken);
            }
        }
    }
}
