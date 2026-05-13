# 🏗️ Guia Técnico: WhatsApp Data Sync

Guia completo de arquitetura, decisões técnicas e aprendizado top-down para o projeto.  
**Leia de cima para baixo**: cada seção aprofunda a anterior.

---

## Sumário

1. [Visão Geral da Arquitetura](#1--visão-geral-da-arquitetura)
2. [Decisões Técnicas](#2--decisões-técnicas)
3. [Roadmap de Aprendizado: Frontend](#3--roadmap-de-aprendizado-frontend)
4. [Roadmap de Aprendizado: Backend](#4--roadmap-de-aprendizado-backend)
5. [Fluxo de Dados Completo](#5--fluxo-de-dados-completo)
6. [Mapa de Arquivos do Projeto](#6--mapa-de-arquivos-do-projeto)
7. [Troubleshooting](#7--troubleshooting)

---

## 1. 🔭 Visão Geral da Arquitetura

O sistema possui **3 camadas independentes** conectadas pelo Redis:

```
┌──────────────────┐       ┌─────────┐       ┌──────────────────┐       ┌──────────────┐
│  Android (Go)    │──────►│  Redis  │◄──────│  Backend (.NET)  │◄──────│  Frontend    │
│  Agente Nativo   │ PUSH  │  Broker │ READ  │  Minimal API     │ HTTP  │  Next.js     │
│                  │◄──────│         │──────►│                  │──────►│  Dashboard   │
│  - Lê msgstore   │ SUB   │         │ PUB   │  - SSE Stream    │ SSE   │  - Live Feed │
│  - Insere contato│       │         │       │  - POST /contacts│       │  - Form      │
└──────────────────┘       └─────────┘       └──────────────────┘       └──────────────┘
```

**Princípio central**: Nenhuma camada conhece a outra diretamente. O Redis é o único ponto de contato. Isso significa que você pode trocar o Frontend inteiro sem tocar no Backend, ou substituir o Agente Go por Rust sem afetar nada.

---

## 2. 🧠 Decisões Técnicas

### 2.1 Por que Go no Android?

| Alternativa | Motivo da rejeição |
|---|---|
| **Java/Kotlin** | Precisa do framework Android completo. Não roda como binário nativo via ADB. |
| **Rust** | Curva de aprendizado do borrow checker seria um bloqueio para o prazo. |
| **C/C++** | Gerenciamento manual de memória para SQLite + Redis seria arriscado. |
| **Go ✅** | Cross-compila para ARM com `GOARCH=arm64`. Goroutines resolvem concorrência. Bibliotecas maduras para Redis (`go-redis`) e SQLite (`go-sqlite3`). |

### 2.2 Por que Redis como Broker (e não Kafka, RabbitMQ, etc)?

- **Latência**: Redis opera em memória. Pub/Sub tem latência sub-milissegundo.
- **Simplicidade**: Não precisa de schemas, partições ou consumer groups para este caso de uso.
- **Dual-purpose**: Funciona como **cache de dados** (Hashes para mensagens) e **message broker** (Pub/Sub para eventos) ao mesmo tempo.

### 2.3 Por que .NET 8 Minimal API (e não Spring Boot, FastAPI, Express)?

- **Requisito do desafio**: A camada de consumo deve ser em .NET/C#.
- **Minimal API**: Elimina a cerimônia de Controllers, Startup.cs e configuração excessiva. Todo o backend cabe em ~100 linhas.
- **Async nativo**: `Task<T>` e `async/await` são cidadãos de primeira classe no C#, ideal para IO-bound (Redis, HTTP).

### 2.4 Por que Next.js 16 (e não Angular, Vue, Svelte)?

- **React Server Components**: Permite decidir o que roda no servidor vs. cliente.
- **App Router**: Sistema de rotas baseado em pastas — intuitivo e zero configuração.
- **Ecossistema**: TailwindCSS + Framer Motion + Lucide Icons = UI premium com esforço mínimo.

### 2.5 Por que SSE (e não WebSockets)?

| SSE | WebSocket |
|---|---|
| Unidirecional (servidor → cliente) | Bidirecional |
| HTTP nativo, funciona com proxies | Protocolo separado (ws://) |
| Reconexão automática pelo browser | Precisa implementar reconexão |
| **Ideal para este caso** ✅ | Overkill — não enviamos dados do browser para o stream |

---

## 3. 🎓 Roadmap de Aprendizado: Frontend

### 3.1 O Modelo Mental do React

Se você vem de Java/Python, o React inverte a lógica:

```
┌─────────────────────────────────────────────────────────┐
│ IMPERATIVO (Java Swing / Python Tkinter)                │
│                                                         │
│   botao.onClick(() -> {                                 │
│       label.setText("Novo valor");   // MUDA o DOM      │
│   });                                                   │
│                                                         │
├─────────────────────────────────────────────────────────┤
│ DECLARATIVO (React)                                     │
│                                                         │
│   const [texto, setTexto] = useState("Valor inicial");  │
│   return <p>{texto}</p>;   // DESCREVE o que mostrar    │
│                                                         │
│   // Quando setTexto("Novo") é chamado,                 │
│   // o React re-renderiza SOZINHO.                      │
└─────────────────────────────────────────────────────────┘
```

**Regra de ouro**: Você nunca toca no HTML diretamente. Você muda o **estado**, e o React atualiza a tela.

### 3.2 Os 3 Hooks que você precisa dominar

#### `useState` — A variável reativa

```tsx
// Equivalente Java:  private int counter = 0;
// Equivalente Python: self.counter = 0
const [counter, setCounter] = useState(0);

// Para mudar:
setCounter(counter + 1);  // Isso REDESENHA o componente inteiro
```

#### `useEffect` — O listener de ciclo de vida

```tsx
// Equivalente Java:  @PostConstruct / onDestroy()
// Equivalente Python: __init__() / __del__()
useEffect(() => {
  // EXECUTA quando o componente aparece na tela
  const connection = openSSE();

  return () => {
    // EXECUTA quando o componente é removido (cleanup)
    connection.close();
  };
}, []);  // [] = executa UMA VEZ. Sem [] = executa a CADA render.
```

#### `useRef` — A variável que NÃO re-renderiza

```tsx
// Equivalente a um campo privado que não dispara atualização visual
const scrollRef = useRef<HTMLDivElement>(null);
// Acessa o elemento DOM real: scrollRef.current
```

### 3.3 Anatomia de um Componente (MessageCard.tsx)

```tsx
// 1. IMPORTS — o que este componente precisa
import { Message } from "@/types";          // Tipagem
import { motion } from "framer-motion";     // Animação

// 2. INTERFACE — contrato de entrada (como um construtor tipado)
interface MessageCardProps {
  message: Message;   // Este componente EXIGE receber uma Message
}

// 3. FUNÇÃO DO COMPONENTE — retorna JSX (HTML turbinado)
export function MessageCard({ message }: MessageCardProps) {
  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
      <span>{message.remetente}</span>
      <p>{message.conteudo}</p>
    </motion.div>
  );
}

// 4. USO — em qualquer lugar do projeto:
// <MessageCard message={minhaMsg} />
```

### 3.4 Estrutura de Pastas do Next.js (App Router)

```
src/
├── app/                    # ROTAS (cada pasta = uma URL)
│   ├── layout.tsx          # Template que envolve TODAS as páginas
│   ├── page.tsx            # Página da rota "/" (raiz)
│   └── globals.css         # Estilos globais
├── components/             # Componentes reutilizáveis
│   ├── MessageCard.tsx     # Card de mensagem individual
│   ├── StatsCard.tsx       # Card de métrica
│   └── ContactForm.tsx     # Formulário de adição de contato
├── lib/                    # Lógica de negócio / utilitários
│   ├── api.ts              # Cliente HTTP (como um Retrofit/requests)
│   └── utils.ts            # Funções auxiliares
└── types/                  # Tipagem TypeScript
    └── index.ts            # Interfaces (Message, DashboardStats)
```

### 3.5 O Fluxo de Dados no Frontend

```
   api.ts                    page.tsx                    MessageCard.tsx
┌──────────┐          ┌─────────────────┐           ┌──────────────────┐
│ fetch()  │─ JSON ──►│ useState()      │── props ──►│ Renderiza card   │
│ SSE      │─ event ─►│ setMessages()   │           │ com animação     │
│ POST     │◄─ form ──│ setStats()      │           │                  │
└──────────┘          └─────────────────┘           └──────────────────┘
```

1. `api.ts` faz a chamada HTTP ou abre o SSE.
2. `page.tsx` recebe os dados e armazena no estado (`useState`).
3. Quando o estado muda, o React re-renderiza os componentes filhos (`MessageCard`, `StatsCard`).
4. `ContactForm` envia dados de volta via `api.ts` → Backend → Redis → Agente Android.

---

## 4. 🎓 Roadmap de Aprendizado: Backend

### 4.1 O Modelo Mental do .NET Minimal API

Se você vem de Spring Boot ou FastAPI:

| Conceito | Spring Boot (Java) | FastAPI (Python) | .NET Minimal API (C#) |
|---|---|---|---|
| Ponto de entrada | `@SpringBootApplication` | `app = FastAPI()` | `var app = builder.Build()` |
| Rota GET | `@GetMapping("/api/x")` | `@app.get("/api/x")` | `app.MapGet("/api/x", ...)` |
| Rota POST | `@PostMapping("/api/x")` | `@app.post("/api/x")` | `app.MapPost("/api/x", ...)` |
| Injeção de Dep. | `@Autowired` | `Depends()` | `builder.Services.AddSingleton<T>()` |
| DTO | `record Dto(...)` | `class Dto(BaseModel)` | `record Dto(string X, string Y)` |

### 4.2 Program.cs — Linha por Linha

```csharp
// 1. SETUP — Configura o container de DI e middlewares
var builder = WebApplication.CreateBuilder(args);

// 2. CORS — Permite que o Next.js (porta 3000) acesse a API (porta 5000)
builder.Services.AddCors(options => {
    options.AddDefaultPolicy(policy => {
        policy.AllowAnyOrigin().AllowAnyHeader().AllowAnyMethod();
    });
});

// 3. DI — Registra serviços. AddSingleton = UMA instância para toda a aplicação.
//    Em Java: @Scope("singleton") @Bean
//    Em Python: módulo-level variable
builder.Services.AddSingleton<RedisService>();

var app = builder.Build();
app.UseCors();

// 4. ENDPOINTS — Cada MapGet/MapPost é uma rota independente
app.MapGet("/api/messages", async (RedisService redis) => { ... });
app.MapGet("/api/stream",   async (HttpContext ctx, RedisService redis) => { ... });
app.MapPost("/api/contacts", async (ContactDto contact, RedisService redis) => { ... });

app.Run();
```

### 4.3 RedisService.cs — O Serviço Central

```csharp
public class RedisService
{
    // CAMPOS PRIVADOS — equivalente a atributos de instância
    private readonly ConnectionMultiplexer _redis;  // Gerenciador de conexões
    private readonly IDatabase _db;                 // Acesso a comandos Redis (GET, HSET, etc)
    private readonly ISubscriber _subscriber;       // Acesso a Pub/Sub

    // CONSTRUTOR — recebe IConfiguration via DI (como @Value no Spring)
    public RedisService(IConfiguration config) { ... }

    // MÉTODOS ASSÍNCRONOS — todos retornam Task<T> (equivalente a Future<T> em Java)
    public async Task<MessageDto?> GetMessageAsync(string id) { ... }
    public async Task<List<MessageDto>> GetRecentMessagesAsync() { ... }
    public async Task PublishContactAsync(object contact) { ... }
}
```

### 4.4 Async/Await — O Conceito Mais Importante

```csharp
// SEM async — a thread TRAVA esperando o Redis responder (BLOQUEANTE)
var result = redis.HashGetAll("msg:1");  // Thread parada por 5ms

// COM async — a thread é LIBERADA enquanto espera (NÃO-BLOQUEANTE)
var result = await redis.HashGetAllAsync("msg:1");  // Thread livre para atender outros requests
```

Em Java, isso seria `CompletableFuture`. Em Python, `asyncio`. A diferença é que no C# o `async/await` é integrado à linguagem, não precisa de runtime especial.

### 4.5 Records vs Classes

```csharp
// CLASSE — mutável, verbosa
public class ContactDto {
    public string Name { get; set; }
    public string Number { get; set; }
}

// RECORD — imutável, 1 linha (Java 14+ Record / Python @dataclass(frozen=True))
public record ContactDto(string Name, string Number);
// O compilador gera automaticamente: construtor, Equals, GetHashCode, ToString
```

---

## 5. 🔄 Fluxo de Dados Completo

### 5.1 Leitura de Mensagens (Android → Dashboard)

```
1. Agente Go detecta nova mensagem no msgstore.db (SQLite polling)
2. Agente Go executa HSET msg:1234 { Remetente, Conteudo, Horario... } no Redis
3. Agente Go executa PUBLISH general:hash-updates "msg:1234"
4. Backend C# (subscrito no canal) recebe o evento
5. Backend C# executa HGETALL msg:1234 para buscar os dados completos
6. Backend C# serializa para JSON e envia via SSE: "data: {...}\n\n"
7. Frontend recebe o evento no EventSource.onmessage
8. Frontend chama setMessages() → React re-renderiza → MessageCard aparece
```

### 5.2 Inserção de Contato (Dashboard → Android)

```
1. Usuário preenche o formulário no Frontend (ContactForm.tsx)
2. Frontend envia POST /api/contacts { name: "João", number: "5511..." }
3. Backend C# valida os dados
4. Backend C# executa PUBLISH contacts:insert '{"name":"João","number":"5511..."}'
5. Agente Go (subscrito no canal contacts:insert) recebe o comando
6. Agente Go executa `content insert` no Android para adicionar o contato
```

---

## 6. 🗂️ Mapa de Arquivos do Projeto

```
Intelitrader-Desafio-Whatsapp-Data-Sync/
│
├── TECH_GUIDE.md              ← VOCÊ ESTÁ AQUI
├── README.md                  ← Visão geral do repositório
├── desafio.md                 ← Enunciado original do desafio
│
├── lab/
│   ├── hello-redis/           ← Laboratório Go + Redis (CRUD, Pub/Sub)
│   │   └── main.go            ← Exemplos de HSET, HGETALL, Subscribe, Publish
│   │
│   ├── adb-connection/        ← Agente nativo Android (Go)
│   │
│   └── dashboard-tests/       ← Stack de monitoramento
│       ├── docker-compose.yml ← Orquestra Backend + Redis via Docker
│       │
│       ├── backend/           ← .NET 8 Minimal API
│       │   ├── Dockerfile     ← Build multi-stage (SDK → Runtime)
│       │   ├── Program.cs     ← Endpoints: GET /messages, GET /stream, POST /contacts
│       │   ├── Services/
│       │   │   └── RedisService.cs  ← Conexão, Subscribe, Publish, SCAN
│       │   └── Models/
│       │       └── MessageDto.cs    ← DTO de mensagem
│       │
│       └── frontend/          ← Next.js 16 + React 19
│           ├── src/app/
│           │   ├── page.tsx         ← Dashboard principal (estado, SSE, layout)
│           │   ├── layout.tsx       ← Template HTML (fontes, metadata)
│           │   └── globals.css      ← Design system (glassmorphism, scrollbar, glow)
│           ├── src/components/
│           │   ├── MessageCard.tsx   ← Card de mensagem individual
│           │   ├── StatsCard.tsx     ← Card de métrica com ícone
│           │   └── ContactForm.tsx   ← Formulário POST /contacts
│           ├── src/lib/
│           │   └── api.ts           ← Cliente HTTP + SSE (Retrofit do TypeScript)
│           └── src/types/
│               └── index.ts         ← Interfaces TypeScript (Message, DashboardStats)
│
└── Zettelkasten/              ← Notas técnicas atômicas
    └── fix-nextjs-node24-oom.md
```

---

## 7. 🛠️ Troubleshooting: Problemas Conhecidos

### 7.1 ⚠️ Node.js v24 + Turbopack = OOM

**Sintoma**: `Fatal process out of memory: Re-embedded builtins: set permissions` ao rodar `npm run dev`.

**Causa**: Node v24 é uma versão instável (Current, não LTS). O Turbopack do Next.js 16 cria múltiplos V8 Isolates que falham na alocação de memória nesta versão.

**Solução aplicada**: Alterado `package.json` para usar `--webpack`:
```json
"dev": "next dev --webpack"
```

**Solução definitiva**: Instalar Node.js **v22 LTS** via `nvm install 22`.

### 7.2 🔌 SSE Crash: ObjectDisposedException

**Sintoma**: O container do backend para com `System.ObjectDisposedException: Cannot write to the response body`.

**Causa**: Quando o browser desconecta (fechar aba, recarregar), o .NET tenta escrever em um stream que já foi descartado.

**Solução aplicada**: Implementamos `CancellationToken` vinculado ao `context.RequestAborted`. Quando o browser desconecta:
1. O token é cancelado.
2. O handler SSE para de escrever.
3. A subscrição Redis é removida (`Unsubscribe`).

### 7.3 🌐 Localhost vs IP de Rede

**Sintoma**: Dashboard funciona no PC mas não em dispositivos na mesma rede.

**Causa**: O frontend usa `localhost:5000` como URL da API. Em outro dispositivo, `localhost` aponta para ele mesmo.

**Solução**: Criar `.env.local` na pasta `frontend/`:
```env
NEXT_PUBLIC_API_URL=http://<SEU-IP>:5000/api
```

### 7.4 🐳 Conflito de Porta do Redis (Docker vs Local)

**Sintoma**: `docker-compose up` falha com `Bind for 0.0.0.0:6379 failed: port is already allocated`.

**Causa**: Já existe um Redis rodando localmente na porta 6379.

**Solução**: Escolha uma das opções:
- **Opção A**: Pare o Redis local antes de subir o Docker.
- **Opção B**: Comente o mapeamento de porta no `docker-compose.yml`:
  ```yaml
  redis:
    image: redis:alpine
    # ports:
    #   - "6379:6379"
  ```
  O backend Docker continuará acessando o Redis interno pela rede `dashboard-network`.

### 7.5 🔧 .NET SDK não encontrado

**Sintoma**: `dotnet run` falha com `No .NET SDKs were found`.

**Causa**: Somente os Runtimes do .NET estão instalados, não o SDK completo.

**Solução**: Use Docker para compilar e rodar o backend:
```powershell
cd lab/dashboard-tests
docker-compose up --build
```
O Dockerfile usa a imagem `mcr.microsoft.com/dotnet/sdk:8.0` para compilar e `aspnet:8.0` para rodar.

---

*Documentação gerada via SDD Pipeline — @engineer + @mentor.*
