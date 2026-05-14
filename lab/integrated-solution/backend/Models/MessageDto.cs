namespace Backend.Models;

/**
 * MessageDto (Data Transfer Object) - Camada de Modelos
 * 
 * Em Java: Seria uma classe com getters/setters ou um Record (Java 14+).
 * Em Python: Seria uma Dataclass.
 * 
 * Usamos 'public' para que outros arquivos possam acessar.
 * 'string?' indica que o valor pode ser nulo (Nullable).
 */
public class MessageDto
{
    public string? Id { get; set; }
    public string? Contexto { get; set; }
    public string? NomeConversa { get; set; }
    public string? Nome_conversa { get; set; }  // -> Alias em snake_case para compatibilidade com o frontend
    public string? Remetente { get; set; }
    public string? Conteudo { get; set; }
    public string? Horario { get; set; }
    public string? ReceivedAt { get; set; }     // -> Timestamp de quando a mensagem foi lida pelo backend
}
