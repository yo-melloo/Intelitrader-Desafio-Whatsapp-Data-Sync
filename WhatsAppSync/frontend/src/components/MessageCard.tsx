import { Message } from "@/types";
import { MessageSquare, User, Clock, Hash } from "lucide-react";
import { motion } from "framer-motion";

interface MessageCardProps {
  message: Message;
}

/**
 * Componente funcional para exibir uma mensagem individual.
 * Pense nisso como uma "View" reutilizável.
 * O 'motion.div' permite animações suaves de entrada.
 */
export function MessageCard({ message }: MessageCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      className="glass p-4 rounded-xl mb-4 hover:border-emerald-500/50 transition-all duration-300 glow"
    >
      <div className="flex justify-between items-start mb-2">
        <div className="flex items-center gap-2">
          <div className="bg-emerald-500/20 p-1.5 rounded-lg">
            <User className="w-4 h-4 text-emerald-400" />
          </div>
          <span className="font-semibold text-emerald-100">
            {message.remetente}
          </span>
        </div>
        <div className="flex items-center gap-1 text-xs text-zinc-500 bg-zinc-900/50 px-2 py-1 rounded-full border border-zinc-800">
          <Clock className="w-3 h-3" />
          {message.horario}
        </div>
      </div>

      <p className="text-zinc-300 text-sm leading-relaxed mb-3 mt-1 ml-1 border-l-2 border-emerald-500/30 pl-3 italic">
        {message.conteudo}
      </p>

      <div className="flex flex-wrap gap-2 text-[10px] uppercase tracking-wider font-bold">
        <div className="flex items-center gap-1 bg-zinc-900/80 text-zinc-400 px-2 py-1 rounded-md border border-zinc-800">
          <MessageSquare className="w-3 h-3" />
          {message.nomeConversa}
        </div>
        <div className="flex items-center gap-1 bg-zinc-900/80 text-zinc-400 px-2 py-1 rounded-md border border-zinc-800">
          <Hash className="w-3 h-3" />
          {message.contexto}
        </div>
      </div>
    </motion.div>
  );
}
