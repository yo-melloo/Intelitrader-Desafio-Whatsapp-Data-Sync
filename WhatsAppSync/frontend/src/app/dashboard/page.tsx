"use client";

import { useEffect, useState, useRef } from "react";
import { Message, DashboardStats } from "@/types";
import { apiService } from "@/lib/api";
import {
  MessageCircle,
  MessageSquare,
  Terminal,
  Zap,
  UserIcon,
  Users,
  Bot,
  X,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { ContactForm } from "@/components/ContactForm";

/**
 * PÁGINA PRINCIPAL DO DASHBOARD
 *
 * Conceitos para quem vem de Java/Python:
 * - 'useState': Gerencia o estado (dados que mudam) da interface. Como variáveis de instância em uma classe.
 * - 'useEffect': Executa código baseado no ciclo de vida (ex: ao carregar a página ou mudar uma variável).
 * - 'fetch': Similar ao HttpClient do Java ou requests do Python.
 */
export default function DashboardPage() {
  const [isMounted, setIsMounted] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [selectedChat, setSelectedChat] = useState<string | null>(null);
  const [stats, setStats] = useState<DashboardStats>({
    totalMessages: 0,
    activeConversations: 0,
    lastUpdate: "Nunca",
  });
  const [isConnected, setIsConnected] = useState(false);
  const [isTerminalOpen, setIsTerminalOpen] = useState(false);
  const [isInfoOpen, setIsInfoOpen] = useState(false);
  const [githubUser, setGithubUser] = useState<{ avatar_url: string } | null>(
    null,
  );
  const [logs, setLogs] = useState<string[]>([]);

  // Referência para o scroll automático
  const scrollRef = useRef<HTMLDivElement>(null);

  // Auto-scroll para a última mensagem
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, selectedChat]);

  // Adiciona logs do sistema
  const addLog = (msg: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs((prev) => [`[${timestamp}] ${msg}`, ...prev].slice(0, 50));
  };

  // Integração Real com Backend C#
  useEffect(() => {
    setIsMounted(true);
    let isMountedLocal = true;

    // 1. Carrega mensagens iniciais
    apiService.getMessages().then((initialMsgs) => {
      if (isMountedLocal && initialMsgs.length > 0) {
        addLog(
          `Carregadas ${initialMsgs.length} mensagens históricas do Redis.`,
        );
        // Ordena por data de recebimento (mais antiga primeiro para o chat)
        const sortedMsgs = [...initialMsgs].sort((a, b) => {
          const dateA = a.receivedAt || a.horario || "";
          const dateB = b.receivedAt || b.horario || "";
          return dateA.localeCompare(dateB);
        });

        // Calcula conversas únicas
        const uniqueChats = Array.from(
          new Set(initialMsgs.map((m) => m.nomeConversa)),
        );
        if (uniqueChats.length > 0) setSelectedChat(uniqueChats[0]);

        setMessages(sortedMsgs);
        setStats((prev) => ({
          ...prev,
          totalMessages: initialMsgs.length,
          activeConversations: uniqueChats.length,
          lastUpdate: new Date().toLocaleTimeString(),
        }));
        setIsConnected(true);
      }
    });

    // 2. Subscreve para atualizações em tempo real (SSE)
    const unsubscribe = apiService.subscribeToUpdates((newMessage) => {
      if (!isMountedLocal) return;
      addLog(
        `Nova mensagem recebida (ID: ${newMessage.id}) de ${newMessage.remetente}`,
      );
      setMessages((prev) => {
        const updated = [...prev, newMessage].slice(-200); // Mantém as últimas 200 no estado
        // Atualiza estatísticas dinamicamente
        setStats((s) => ({
          ...s,
          totalMessages: s.totalMessages + 1,
          activeConversations: new Set(updated.map((m) => m.nomeConversa)).size,
          lastUpdate: new Date().toLocaleTimeString(),
        }));
        return updated;
      });
      setIsConnected(true);
    });

    // 3. Busca dados do desenvolvedor no GitHub
    fetch("https://api.github.com/users/yo-melloo")
      .then((res) => res.json())
      .then((data) => {
        if (data.avatar_url) setGithubUser(data);
      })
      .catch(() => console.log("Erro ao buscar avatar do GitHub"));

    return () => {
      isMountedLocal = false;
      unsubscribe();
    };
  }, []);

  // Filtra conversas únicas para a barra lateral
  const conversations = Array.from(
    new Set(messages.map((m) => m.nomeConversa)),
  ).map((name) => {
    const lastMsg = messages
      .filter((m) => m.nomeConversa === name)
      .slice(-1)[0];
    return { name, lastMsg };
  });

  // Mensagens da conversa selecionada
  const activeMessages = messages.filter(
    (m) => m.nomeConversa === selectedChat,
  );

  return (
    <div className="h-screen bg-[#0a0a0a] text-zinc-100 flex flex-col font-sans selection:bg-emerald-500/30 overflow-hidden">
      {/* Header Compacto */}
      <header className="bg-zinc-950/50 border-b border-white/5 p-4 flex justify-between items-center z-10 backdrop-blur-md">
        <div className="flex items-center gap-4">
          <button
            onClick={() => setIsInfoOpen(true)}
            className="flex items-center gap-2 hover:opacity-80 transition-opacity group"
          >
            <Zap className="text-emerald-400 w-5 h-5 group-hover:scale-110 transition-transform" />
            <h1 className="text-lg font-bold tracking-tight">
              WhatsApp <span className="text-emerald-500">Sync</span>
            </h1>
          </button>
          <div
            className={`px-3 py-1 rounded-full border flex items-center gap-2 ${
              isConnected
                ? "border-emerald-500/30 bg-emerald-500/5 text-emerald-400"
                : "border-amber-500/30 bg-amber-500/5 text-amber-400"
            }`}
          >
            <div
              className={`w-1.5 h-1.5 rounded-full ${isConnected ? "bg-emerald-500 animate-pulse" : "bg-amber-500"}`}
            />
            <span className="text-[10px] font-bold uppercase tracking-wider">
              {isConnected ? "Agente Online" : "Desconectado"}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsTerminalOpen(true)}
            className="p-2 rounded-lg hover:bg-white/5 text-zinc-400 transition-colors"
          >
            <Terminal className="w-4 h-4" />
          </button>
          <div className="text-[10px] text-zinc-500 uppercase font-bold tracking-widest hidden md:block">
            {stats.totalMessages} Mensagens Sincronizadas
          </div>
        </div>
      </header>

      {/* Layout Principal */}
      <main className="flex-1 flex overflow-hidden">
        {/* Sidebar: Lista de Conversas */}
        <aside className="w-80 border-right border-white/5 bg-zinc-950/20 flex flex-col border-r">
          <div className="p-4 border-b border-white/5 flex flex-col gap-4">
            <h2 className="text-xs font-black uppercase tracking-[0.2em] text-zinc-500 flex items-center gap-2">
              <MessageCircle className="w-3 h-3" /> Conversas Ativas
            </h2>
            <ContactForm />
          </div>

          <div className="flex-1 overflow-y-auto custom-scrollbar">
            {conversations.length === 0 ? (
              <div className="p-10 text-center text-zinc-700 italic text-sm">
                Nenhuma conversa encontrada...
              </div>
            ) : (
              conversations.map((chat) => (
                <button
                  key={chat.name}
                  onClick={() => setSelectedChat(chat.name)}
                  className={`w-full p-4 flex items-start gap-3 transition-all border-b border-white/5 hover:bg-white/[0.02] text-left group ${
                    selectedChat === chat.name
                      ? "bg-emerald-500/10 border-l-4 border-l-emerald-500"
                      : "border-l-4 border-l-transparent"
                  }`}
                >
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-zinc-800 to-zinc-900 flex items-center justify-center shrink-0 border border-white/10 group-hover:border-emerald-500/30">
                    {chat.lastMsg.contexto === "GRUPO" ? (
                      <Users
                        className={`w-5 h-5 ${selectedChat === chat.name ? "text-emerald-400" : "text-zinc-500"}`}
                      />
                    ) : (
                      <UserIcon
                        className={`w-5 h-5 ${selectedChat === chat.name ? "text-emerald-400" : "text-zinc-500"}`}
                      />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-baseline mb-1">
                      <div className="flex items-center gap-2 overflow-hidden">
                        <h3 className="font-bold text-sm truncate text-zinc-200">
                          {chat.name}
                        </h3>
                        <span
                          className={`text-[7px] px-1 py-0.2 rounded-full font-black uppercase tracking-tighter border shrink-0 ${
                            chat.lastMsg.contexto === "GRUPO"
                              ? "bg-blue-500/10 border-blue-500/30 text-blue-400"
                              : "bg-purple-500/10 border-purple-500/30 text-purple-400"
                          }`}
                        >
                          {chat.lastMsg.contexto}
                        </span>
                      </div>
                      <span className="text-[9px] text-zinc-600 font-mono">
                        {chat.lastMsg.horario.split(" ")[1].substring(0, 5)}
                      </span>
                    </div>
                    <p className="text-xs text-zinc-500 truncate italic">
                      {chat.lastMsg.conteudo}
                    </p>
                  </div>
                </button>
              ))
            )}
          </div>
        </aside>

        {/* Chat Area */}
        <section className="flex-1 flex flex-col bg-[#0d0d0d] relative">
          {selectedChat ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b border-white/5 bg-zinc-950/40 flex items-center justify-between backdrop-blur-sm">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
                    {activeMessages.length > 0 &&
                    activeMessages[0].contexto === "GRUPO" ? (
                      <Users className="w-4 h-4 text-emerald-400" />
                    ) : (
                      <UserIcon className="w-4 h-4 text-emerald-400" />
                    )}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h2 className="font-bold text-sm">{selectedChat}</h2>
                      {activeMessages.length > 0 && (
                        <span
                          className={`text-[8px] px-1.5 py-0.5 rounded-full font-black uppercase tracking-wider border ${
                            activeMessages[0].contexto === "GRUPO"
                              ? "bg-blue-500/10 border-blue-500/30 text-blue-400"
                              : "bg-purple-500/10 border-purple-500/30 text-purple-400"
                          }`}
                        >
                          {activeMessages[0].contexto}
                        </span>
                      )}
                    </div>
                    <p className="text-[10px] text-emerald-500/60 uppercase font-black tracking-widest">
                      Monitorando em Tempo Real
                    </p>
                  </div>
                </div>
                <div className="flex gap-2">
                  <div className="px-2 py-1 rounded bg-zinc-900 text-[10px] text-zinc-500 border border-white/5">
                    {activeMessages.length} mensagens
                  </div>
                </div>
              </div>

              {/* Message Bubbles Area */}
              <div
                ref={scrollRef}
                className="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] bg-fixed"
              >
                <AnimatePresence initial={false}>
                  {activeMessages.map((msg, index) => {
                    const isMe = msg.remetente === "Você";
                    return (
                      <motion.div
                        key={msg.id}
                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                        animate={{ opacity: 1, y: 0, scale: 1 }}
                        className={`flex ${isMe ? "justify-end" : "justify-start"}`}
                      >
                        <div
                          className={`max-w-[70%] group relative ${isMe ? "items-end" : "items-start"}`}
                        >
                          {!isMe && (
                            <span className="text-[9px] font-bold text-emerald-500/70 mb-1 ml-1 uppercase tracking-tighter">
                              {msg.remetente}
                            </span>
                          )}
                          <div
                            className={`p-3 rounded-2xl shadow-xl text-sm relative ${
                              isMe
                                ? "bg-emerald-600 text-white rounded-tr-none border border-emerald-400/20"
                                : "bg-zinc-800 text-zinc-200 rounded-tl-none border border-white/5"
                            }`}
                          >
                            {msg.conteudo}
                            <div
                              className={`text-[9px] mt-1 text-right opacity-50 font-mono`}
                            >
                              {msg.horario.split(" ")[1].substring(0, 5)}
                            </div>
                          </div>
                        </div>
                      </motion.div>
                    );
                  })}
                </AnimatePresence>
              </div>

              {/* Fake Input Area for aesthetics */}
              <div className="p-4 bg-zinc-950/80 border-t border-white/5 backdrop-blur-md">
                <div className="max-w-3xl mx-auto bg-zinc-900/50 border border-white/5 rounded-full px-6 py-3 flex items-center gap-4 text-zinc-600 text-sm italic">
                  <Bot className="w-4 h-4 text-emerald-500/50" />O agente está
                  monitorando esta conversa...
                </div>
              </div>
            </>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-zinc-700 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-zinc-900/20 via-transparent to-transparent">
              <div className="p-6 rounded-full bg-zinc-900/50 mb-6 border border-white/5">
                <MessageSquare className="w-12 h-12 opacity-20" />
              </div>
              <h2 className="text-lg font-bold text-zinc-500">
                Selecione uma conversa
              </h2>
              <p className="text-sm mt-2">
                Para visualizar o histórico de sincronização
              </p>
            </div>
          )}
        </section>
      </main>

      {/* Terminal / Log Overlay (Mantido igual) */}
      <AnimatePresence>
        {isTerminalOpen && (
          <motion.div
            initial={{ opacity: 0, y: 100 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 100 }}
            className="fixed bottom-6 right-6 left-6 md:left-auto md:w-[500px] z-50"
          >
            <div className="bg-zinc-950 border border-emerald-500/30 rounded-2xl shadow-2xl overflow-hidden flex flex-col h-[400px]">
              <div className="bg-zinc-900 px-4 py-3 border-b border-white/5 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Terminal className="w-4 h-4 text-emerald-400" />
                  <span className="text-[10px] font-bold uppercase tracking-widest text-zinc-400">
                    Console de Logs (Redis Stream)
                  </span>
                </div>
                <button
                  onClick={() => setIsTerminalOpen(false)}
                  className="p-1 hover:bg-white/5 rounded-md transition-colors"
                >
                  <X className="w-4 h-4 text-zinc-500" />
                </button>
              </div>
              <div className="flex-1 overflow-y-auto p-4 font-mono text-[10px] space-y-1 custom-scrollbar">
                {logs.length === 0 ? (
                  <p className="text-zinc-700 italic">
                    Nenhum evento registrado ainda...
                  </p>
                ) : (
                  logs.map((log, i) => (
                    <div key={i} className="flex gap-3">
                      <span className="text-emerald-500/40 shrink-0">
                        [{logs.length - i}]
                      </span>
                      <span className="text-zinc-400 break-all">{log}</span>
                    </div>
                  ))
                )}
              </div>
              <div className="px-4 py-2 bg-emerald-500/5 border-t border-white/5">
                <p className="text-[9px] text-emerald-500/50 flex items-center gap-2 uppercase font-bold tracking-tighter">
                  <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse" />
                  Escutando Porta 6379 via Pub/Sub
                </p>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      {/* Info Modal */}
      <AnimatePresence>
        {isInfoOpen && (
          <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setIsInfoOpen(false)}
              className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.9, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.9, y: 20 }}
              className="bg-zinc-950 border border-emerald-500/30 rounded-3xl shadow-2xl overflow-hidden max-w-lg w-full relative z-10"
            >
              <div className="h-32 bg-gradient-to-br from-emerald-600/20 to-zinc-900 relative">
                <div className="absolute -bottom-10 left-8">
                  <div className="w-20 h-20 rounded-2xl bg-zinc-900 border-2 border-emerald-500 flex items-center justify-center shadow-xl overflow-hidden">
                    {githubUser?.avatar_url ? (
                      <img
                        src={githubUser.avatar_url}
                        alt="GitHub Avatar"
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <UserIcon className="w-10 h-10 text-emerald-400" />
                    )}
                  </div>
                </div>
                <button
                  onClick={() => setIsInfoOpen(false)}
                  className="absolute top-4 right-4 p-2 hover:bg-white/10 rounded-full transition-colors"
                >
                  <X className="w-5 h-5 text-zinc-400" />
                </button>
              </div>

              <div className="p-8 pt-14">
                <div className="flex justify-between items-start mb-6">
                  <div>
                    <h2 className="text-2xl font-bold">Gustavo Mello</h2>
                    <p className="text-emerald-500 font-mono text-xs uppercase tracking-widest mt-1">
                      Desenvolvedor Back-End
                    </p>
                  </div>
                  <a
                    href="https://github.com/yo-melloo"
                    target="_blank"
                    className="px-4 py-2 bg-zinc-900 border border-white/5 rounded-xl text-xs font-bold hover:bg-emerald-500/10 hover:border-emerald-500/30 transition-all flex items-center gap-2"
                  >
                    GitHub
                  </a>
                </div>

                <div className="space-y-6">
                  <div>
                    <h3 className="text-[10px] font-black uppercase tracking-[0.2em] text-zinc-500 mb-2">
                      Sobre o Projeto
                    </h3>
                    <p className="text-zinc-400 text-sm leading-relaxed">
                      Desenvolvido para o{" "}
                      <span className="text-zinc-100 font-semibold">
                        Desafio Técnico Intelitrader
                      </span>
                      , este projeto sincroniza mensagens do WhatsApp em tempo
                      real utilizando um Agente NDK nativo (Go) e uma
                      arquitetura resiliente baseada em Redis e .NET.
                    </p>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-zinc-900/50 p-4 rounded-2xl border border-white/5">
                      <p className="text-[9px] font-bold text-zinc-500 uppercase mb-1">
                        Tecnologia Base
                      </p>
                      <p className="text-xs font-mono text-emerald-400">
                        Go / .NET / Next.js
                      </p>
                    </div>
                    <div className="bg-zinc-900/50 p-4 rounded-2xl border border-white/5">
                      <p className="text-[9px] font-bold text-zinc-500 uppercase mb-1">
                        Status
                      </p>
                      <p className="text-xs font-mono text-emerald-400">
                        Desafio Concluído
                      </p>
                    </div>
                  </div>

                  <div className="pt-4 border-t border-white/5">
                    <p className="text-center text-[10px] text-zinc-600 italic">
                      "A engenharia não é sobre o que você sabe, mas sobre como
                      você resolve o que ainda não conhece."
                    </p>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
}
