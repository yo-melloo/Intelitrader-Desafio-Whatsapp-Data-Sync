import { motion } from "framer-motion";
import { LucideIcon } from "lucide-react";

interface StatsCardProps {
  label: string;
  value: string | number;
  icon: LucideIcon;
  color: string;
}

/**
 * Componente para exibir métricas rápidas.
 * Recebe um ícone e cores dinâmicas.
 */
export function StatsCard({ label, value, icon: Icon, color }: StatsCardProps) {
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="glass p-6 rounded-2xl flex items-center gap-5 border-l-4"
      style={{ borderLeftColor: color }}
    >
      <div className="p-3 rounded-xl" style={{ backgroundColor: `${color}15` }}>
        <Icon className="w-6 h-6" style={{ color }} />
      </div>
      <div>
        <p className="text-zinc-500 text-xs font-medium uppercase tracking-widest">
          {label}
        </p>
        <p className="text-2xl font-bold text-white mt-1">{value}</p>
      </div>
    </motion.div>
  );
}
