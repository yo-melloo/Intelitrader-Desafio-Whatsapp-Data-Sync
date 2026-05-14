/**
 * Interface para representar uma mensagem vinda do Redis/Agent.
 * Em Java, isso seria uma Classe POJO (Plain Old Java Object) ou um Record.
 * Em Python, seria um Dataclass ou um Dicionário tipado.
 */
export interface Message {
  id: string;
  contexto: string;
  nomeConversa: string;
  remetente: string;
  conteudo: string;
  horario: string;
  receivedAt: string; // Timestamp local de quando chegou no dashboard
}

export interface DashboardStats {
  totalMessages: number;
  activeConversations: number;
  lastUpdate: string;
}
