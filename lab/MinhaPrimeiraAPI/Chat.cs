public class Chat
{
    public string Nome { get; set; } = string.Empty;
    public string Telefone { get; set; } = string.Empty;
    public string Foto { get; set; } = string.Empty;
    public DateTime UltimaMensagem { get; set; }
    public string Mensagem { get; set; } = string.Empty;
    public bool EstaEscrito { get; set; } = false;
    public bool EstaOnline { get; set; } = false;
}