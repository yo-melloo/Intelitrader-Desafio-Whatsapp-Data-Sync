"use client";

import { useState } from "react";
import { UserPlus, Loader2, CheckCircle2, Phone } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { apiService } from "@/lib/api";

export function ContactForm() {
  const [name, setName] = useState("");
  const [number, setNumber] = useState("");
  const [status, setStatus] = useState<
    "idle" | "loading" | "success" | "error"
  >("idle");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !number) return;

    setStatus("loading");
    const success = await apiService.addContact(name, number);

    if (success) {
      setStatus("success");
      setName("");
      setNumber("");
      setTimeout(() => setStatus("idle"), 3000);
    } else {
      setStatus("error");
      setTimeout(() => setStatus("idle"), 3000);
    }
  };

  return (
    <div className="glass p-6 rounded-3xl border border-white/5 bg-white/[0.02] shadow-2xl overflow-hidden relative group">
      {/* Background Glow */}
      <div className="absolute -top-24 -right-24 w-48 h-48 bg-emerald-500/10 blur-[80px] rounded-full group-hover:bg-emerald-500/20 transition-colors duration-700" />

      <div className="relative z-10">
        <h3 className="text-sm font-bold uppercase tracking-widest text-zinc-400 mb-6 flex items-center gap-2">
          <UserPlus className="w-4 h-4 text-emerald-400" /> Adicionar Novo
          Contato
        </h3>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label className="text-[10px] uppercase font-bold text-zinc-500 ml-1">
              Nome
            </label>
            <input
              type="text"
              placeholder="Ex: João Silva Santos"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-black/40 border border-white/5 rounded-xl px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50 transition-all placeholder:text-zinc-700"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-[10px] uppercase font-bold text-zinc-500 ml-1">
              Número
            </label>
            <div className="relative">
              <Phone className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-600" />
              <input
                type="tel"
                placeholder="55119999999"
                value={number}
                onChange={(e) => setNumber(e.target.value)}
                className="w-full bg-black/40 border border-white/5 rounded-xl pl-12 pr-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50 transition-all placeholder:text-zinc-700"
                required
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={status === "loading"}
            className={`w-full py-3 rounded-xl font-bold text-xs uppercase tracking-widest flex items-center justify-center gap-2 transition-all active:scale-[0.98] ${
              status === "success"
                ? "bg-emerald-500 text-white shadow-[0_0_20px_rgba(16,185,129,0.3)]"
                : status === "error"
                  ? "bg-rose-500 text-white"
                  : "bg-emerald-600 hover:bg-emerald-500 text-white shadow-[0_0_20px_rgba(16,185,129,0.2)]"
            }`}
          >
            <AnimatePresence mode="wait">
              {status === "loading" ? (
                <motion.div
                  key="loading"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                >
                  <Loader2 className="w-4 h-4 animate-spin" />
                </motion.div>
              ) : status === "success" ? (
                <motion.div
                  key="success"
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  className="flex items-center gap-2"
                >
                  <CheckCircle2 className="w-4 h-4" /> Enviado para Android
                </motion.div>
              ) : (
                <span key="idle">Adicionar Contato</span>
              )}
            </AnimatePresence>
          </button>
        </form>

        <p className="text-[9px] text-zinc-600 mt-6 text-center leading-relaxed">
          * O comando será publicado no canal{" "}
          <code className="text-emerald-400/70">contacts:insert</code> do Redis.
        </p>
      </div>
    </div>
  );
}
