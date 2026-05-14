import { Message } from "@/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:5000/api";

/**
 * Serviço para interagir com o backend C#.
 * 
 * Para quem vem de Java: Pense nisso como um Cliente HTTP (Retrofit ou OpenFeign).
 * Para quem vem de Python: Como uma classe que usa a biblioteca 'requests'.
 */
export const apiService = {
  /**
   * Busca todas as mensagens iniciais.
   */
  async getMessages(): Promise<Message[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/messages`);
      if (!response.ok) throw new Error("Erro ao buscar mensagens");
      return await response.json();
    } catch (error) {
      console.error("Erro na API:", error);
      return [];
    }
  },

  /**
   * Subscreve para atualizações em tempo real usando Server-Sent Events (SSE).
   * SSE é mais leve que WebSockets para fluxos unidirecionais (servidor -> cliente).
   */
  subscribeToUpdates(onMessage: (msg: Message) => void) {
    const eventSource = new EventSource(`${API_BASE_URL}/stream`);

    eventSource.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        onMessage(message);
      } catch (error) {
        console.error("Erro ao processar mensagem SSE:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("Erro no EventSource:", error);
      eventSource.close();
    };

    return () => eventSource.close(); // Função para limpar a conexão
  },

  /**
   * Envia um novo contato para o Backend C#.
   */
  async addContact(name: string, number: string): Promise<boolean> {
    try {
      const response = await fetch(`${API_BASE_URL}/contacts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, number }),
      });
      return response.ok;
    } catch (error) {
      console.error("Erro ao adicionar contato:", error);
      return false;
    }
  }
};
