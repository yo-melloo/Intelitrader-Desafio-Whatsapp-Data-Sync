import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Utilitário para mesclar classes do Tailwind CSS.
 * Familiar para quem usa bibliotecas como shadcn/ui.
 * Em Java, isso seria como um método estático utilitário para manipulação de Strings.
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
