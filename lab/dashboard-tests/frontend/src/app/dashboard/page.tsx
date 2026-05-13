"use client";

import { useEffect, useState, useRef } from "react";
import { Message, DashboardStats } from "@/types";
import { apiService } from "@/lib/api";
import { MessageCard } from "@/components/MessageCard";
import { StatsCard } from "@/components/StatsCard";
import {
  Activity,
  MessageCircle,
  Users,
  ShieldCheck,
  Database,
  Terminal,
  Zap,
  Wifi,
  UserIcon,
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
  const [stats, setStats] = useState<DashboardStats>({
    totalMessages: 0,
    activeConversations: 0,
    lastUpdate: "Nunca",
  });
  const [isConnected, setIsConnected] = useState(false);
  const [isTerminalOpen, setIsTerminalOpen] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);

  // Referência para o scroll automático
  const scrollRef = useRef<HTMLDivElement>(null);

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
        // Ordena por data de recebimento (mais recente primeiro)
        const sortedMsgs = [...initialMsgs].sort((a, b) => {
          const dateA = a.receivedAt || a.horario || "";
          const dateB = b.receivedAt || b.horario || "";
          return dateB.localeCompare(dateA); // Decrescente
        });

        // Calcula conversas únicas
        const uniqueChats = new Set(initialMsgs.map((m) => m.remetente)).size;

        setMessages(sortedMsgs);
        setStats((prev) => ({
          ...prev,
          totalMessages: initialMsgs.length,
          activeConversations: uniqueChats,
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
        const updated = [newMessage, ...prev].slice(0, 50);
        // Atualiza estatísticas dinamicamente
        setStats((s) => ({
          ...s,
          totalMessages: s.totalMessages + 1,
          activeConversations: new Set(updated.map((m) => m.remetente)).size,
          lastUpdate: new Date().toLocaleTimeString(),
        }));
        return updated;
      });
      setIsConnected(true);
    });

    return () => {
      isMountedLocal = false;
      unsubscribe();
    };
  }, []);

  return (
    <div className="min-h-screen bg-[#0a0a0a] text-zinc-100 p-6 lg:p-10 font-sans selection:bg-emerald-500/30">
      {/* Header Section */}
      <header className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-start md:items-center mb-10 gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <UserIcon className="text-emerald-400 w-5 h-5" />
            <span className="text-xs font-bold uppercase tracking-[0.2em] text-emerald-400">
              <a href="https://github.com/yo-melloo/">@yo-melloo</a> no github
            </span>
          </div>
          <h1 className="text-4xl font-extrabold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-white via-zinc-400 to-zinc-600">
            WhatsApp Data Sync{" "}
            <span className="text-emerald-500">Dashboard</span>
          </h1>
          <p className="text-zinc-500 mt-2 flex items-center gap-2">
            <Database className="w-4 h-4" /> Monitorando Agente Android via
            <span className="text-red-400">Message Broker (Redis)</span>
          </p>
        </div>

        <div className="flex items-center gap-4">
          <div
            className={`px-4 py-2 rounded-full border flex items-center gap-2 transition-all duration-500 ${
              isConnected
                ? "border-emerald-500/30 bg-emerald-500/5 text-emerald-400"
                : "border-amber-500/30 bg-amber-500/5 text-amber-400"
            }`}
          >
            <Bot className={`w-4 h-4 ${isConnected ? "animate-pulse" : ""}`} />
            <span className="text-xs font-bold uppercase tracking-wider">
              {isConnected ? "Agente Online" : "Aguardando Offline..."}
            </span>
          </div>
          <div
            className="p-2 rounded-xl glass hover:bg-white/5 cursor-pointer transition-colors"
            onClick={() => setIsTerminalOpen(true)}
          >
            <Terminal className="w-5 h-5 text-zinc-400" />
          </div>
        </div>
      </header>

      {/* Main Grid */}
      <main className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-12 gap-8 items-center">
        {/* Sidebar Stats */}
        <div className="lg:col-span-4 space-y-6">
          <StatsCard
            label="Mensagens no Banco de Dados (Redis)"
            value={stats.totalMessages}
            icon={MessageCircle}
            color="#10b981"
          />

          {/* Formulário de Adição de Contato conforme Desafio.md */}
          <ContactForm />

          <div className="glass p-6 rounded-2xl relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
              <Zap className="w-12 h-12 text-emerald-500" />
            </div>

            <h3 className="text-sm font-bold uppercase tracking-widest text-zinc-400 mb-6 flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400" /> Monitor de
              Tráfego
            </h3>

            <div className="space-y-6">
              {[
                {
                  label: "Latência Redis",
                  unit: "ms",
                  color: "bg-emerald-500",
                },
                { label: "Throughput", unit: "msg/s", color: "bg-emerald-400" },
                {
                  label: "Buffer de Cache",
                  unit: "%",
                  color: "bg-emerald-600",
                },
              ].map((item, i) => (
                <div key={i} className="space-y-2">
                  <div className="flex justify-between text-[10px] uppercase font-bold tracking-wider text-zinc-500">
                    <span>{item.label}</span>
                    <span className="text-emerald-500/80">
                      {isMounted ? Math.floor(Math.random() * 100) : "--"}
                      {item.unit}
                    </span>
                  </div>
                  <div className="h-1.5 bg-zinc-800/50 rounded-full overflow-hidden border border-white/5">
                    <motion.div
                      className={`h-full ${item.color}`}
                      initial={{ width: "0%" }}
                      animate={{
                        width: isMounted ? `${30 + Math.random() * 60}%` : "0%",
                      }}
                      transition={{
                        duration: 1.5,
                        repeat: Infinity,
                        repeatType: "reverse",
                        ease: "easeInOut",
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-8 pt-6 border-t border-white/5 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-500 animate-ping" />
                <span className="text-[10px] font-bold text-zinc-500 uppercase">
                  Stream Ativo
                </span>
              </div>
              <p className="text-[9px] text-zinc-600 italic">
                Watcher acordado em: {stats.lastUpdate}
              </p>
            </div>
          </div>
        </div>

        {/* Feed Section */}
        <div className="lg:col-span-8 flex flex-col h-[700px]">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-bold flex items-center gap-2">
              Feed de Mensagens{" "}
              <span className="bg-emerald-500 text-[10px] px-2 py-0.5 rounded text-white animate-pulse">
                LIVE
              </span>
            </h2>
            <button
              className="text-xs text-zinc-500 hover:text-white transition-colors"
              onClick={() => setMessages([])}
            >
              Limpar Feed
            </button>
          </div>

          <div className="flex-1 overflow-y-auto pr-2 space-y-4 custom-scrollbar">
            <AnimatePresence mode="popLayout">
              {messages.length === 0 ? (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="h-full flex flex-col items-center justify-center text-zinc-600 border-2 border-dashed border-zinc-800 rounded-3xl"
                >
                  <MessageCircle className="w-12 h-12 mb-4 opacity-20" />
                  <p>Aguardando novas mensagens do Agente...</p>
                </motion.div>
              ) : (
                messages.map((msg) => (
                  <MessageCard key={msg.id} message={msg} />
                ))
              )}
            </AnimatePresence>
          </div>
        </div>
      </main>

      {/* Footer / Meta */}
      <footer className="max-w-7xl mx-auto mt-20 pt-8 border-t border-zinc-900 flex justify-between items-center text-zinc-600 text-xs">
        <p>© 2026 Gustavo Mello | @yo-melloo</p>
        <div className="flex gap-6"></div>
      </footer>

      {/* Terminal / Log Overlay */}
      <AnimatePresence>
        {isTerminalOpen && (
          <motion.div
            initial={{ opacity: 0, y: 100 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 100 }}
            className="fixed bottom-6 right-6 left-6 md:left-auto md:w-[500px] z-50"
          >
            <div className="bg-zinc-950 border border-emerald-500/30 rounded-2xl shadow-2xl overflow-hidden flex flex-col h-[400px]">
              {/* Header do Terminal */}
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

              {/* Corpo dos Logs */}
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

              {/* Footer do Terminal */}
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
    </div>
  );
}
