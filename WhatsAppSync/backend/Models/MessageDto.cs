namespace Backend.Models;

/**
 * MessageDto (Data Transfer Object) - Camada de Modelos
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
