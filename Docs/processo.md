### Etapa-001: Configuração do Ambiente

- [x] Instalou GO
- [x] Executou Olá Mundo em GO: Entendeu que a linguagem funciona usando `packages`
- [x] Instalou Android Studio
- [x] Instalou, executou e configurou um Emulador de Android
  - Primeira tentativa deu erro: Usei uma imagem Android que não dava acesso root (Pixel 14 Pro)
  - Segunda tentativa deu certo: Usei uma imagem Android que dava acesso root (Pixel 6 Pro)
  - Tentei instalar APK do WhatsApp Business, mas o emulador não instalou. Tentei do WhatsApp comum, e funcionou.
  - Personalizei o Android para entender o quanto ele é limitado na emulação, e na minha máquina (não é - o que é bom).
- [x] Baixou NDK (que não vem instalado por padrão)
- [x] Realizou teste de acesso ao root do sistema (manual), copiando o banco de dados do WhatsApp e analisando na máquina local
  - Usei SQLite para abrir o arquivo `msgstore.db` copiado para o armazenamento do meu computador
  - Identifiquei dezenas de tabelas, e pesquisei qual delas armazena as mensagens (literalmente uma tabela chamada `message`) [Informação vai servir mais tarde para construir a consulta SQL]
- [x] Criou container Docker para usar Redis na máquina local
- [x] Usou IA para criar o primeiro código `main.go`:
  - Código revisado (vide comentários adicionados)
  - Aprendeu buildar binários do Go
  - Aprendeu a dar push em arquivos via Adb
  - Ao ser buildado e pushado para o Android, o binário tenta conexão com o banco de dados do WhatsApp, e com o Redis (validando conexão)
- [x] Usou IA para reescrever o `.gitignore`, e manter boas práticas para evitar commitar segredos e dependências pesadas.
- [x] Usou IA para reescrever o `README.md`, e melhorar a descrição pré-escrita do projeto.
- [x] Gerou arquivo LICENSE com as condições de uso do projeto (restrito apenas para o teste).

> Commit: 8d45888 feat(ambiente): Configura repositório, documentação e prepara Agente GO para execução no Android

---

### Etapa-002: Implementar leitura do banco de dados

#### DECISÃO-TÉCNICA-001: Polling vs. Observer

Após analisar, entendi que Polling realizaria uma consulta periódica no banco de dados, enquanto o Observer iria aguardar alguma mudança acontecer para disparar a função do agente. Para me decidir entre um e outro, levei em conta que o critério da aplicação é ser em tempo real, se reduzisse o tempo de polling isso poderia ser prejucidial tanto à bateria e performance quanto à experiencia. Já que existe um processo rodando ativamente chamado inotify no kernel Linux (base do Android) que pode ser apontado para o Write-Ahead Logging (WAL) do banco de dados do WhatsApp, programar um Observer seria como programar um gatilho que só dispara quando necessário, economizando recursos.

Decisão: Usar Observer para "vigiar" banco de dados com o próprio "vigia do sistema" (inotify), e um time.Tick para fazer polling no banco de dados quando no android, para resistir ao Write-back Cache.

Consequência: Arquitetura do projeto a partir desse ponto vira um **Event-Driven**.

Problema identificado: Alguns programas podem disparar mais de uma notificação para uma simples operação, o que pode gerar consultas desnecessárias ao banco de dados

```bash
$ go run .

# Usei .txt nesse exemplo
2026/05/09 18:50:37 [WATCHER] Watcher adicionado ao arquivo: ".//db/teste-db.txt"

# Um único save, duas ações de Write
[WATCHER] Evento detectado

2026/05/09 18:50:40 [WATCHER] O arquivo foi modificado.
WRITE

[WATCHER] Evento detectado
2026/05/09 18:50:40 [WATCHER] O arquivo foi modificado.
WRITE


exit status 0xc000013a
```

Solução: Criado fila e filtro de mudanças. Durante o processo me deparei com uma race condition, relatado em `Docs/dificuldades.md`

```go
case event, ok := <-watcher.Events:
  if !ok {
    return
  }
  log.Println("[WATCHER] Evento detectado")

  if event.Has(fsnotify.Write) {
    if timer != nil {
      timer.Stop() // para timers anteriores, se existirem, para evitar logs repetidos
    }

    timer = time.AfterFunc(200*time.Millisecond, func() { // inicia um novo timer para gerar o log após um curto atraso
      buscarUltimaMensagem(database) // o log que exibe o conteúdo do registro encontrado no banco de dados está nessa função
    })
  }

  // ...

func buscarUltimaMensagem(db *sql.DB) { // Lógica de fila para buscar os próximos registros do banco de dados a partir do lastProcessID (global)

	query := "SELECT id, content FROM example WHERE id > ? ORDER BY id ASC" // Seleciona o próximo registro com ID maior que lastProcessID

	rows, err := db.Query(query, lastProcessID)
	if err != nil {
		log.Println("[SQL] Falha na consulta. Verifique a conexão com o banco de dados e a estrutura da tabela.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var conteudo string

		if err := rows.Scan(&id, &conteudo); err != nil {
			fmt.Printf("[SQL] Erro ao ler os dados da linha: %v.\n", err)
			continue
		}

		fmt.Printf("[WATCHER] Conteúdo adicionado - ID: %d | Conteúdo: %s\n", id, conteudo)

		lastProcessID = id // Atualiza o lastProcessID (global) para o ID do registro encontrado
	}

}

```

Resultado:

```bash

$ go run .

2026/05/10 01:04:48 [NATIVE AGENT] Iniciando o observador de banco de dados...
2026/05/10 01:04:48 [SQLITE3 DRIVER] Conexão com o banco de dados estabelecida com sucesso.
2026/05/10 01:04:48 [WATCHER] Watcher adicionado ao arquivo: ".//db/teste-wal.db-wal"
[WATCHER] Conteúdo adicionado - ID: 47 | Conteúdo: w
[WATCHER] Conteúdo adicionado - ID: 48 | Conteúdo: x
[WATCHER] Conteúdo adicionado - ID: 49 | Conteúdo: y
[WATCHER] Conteúdo adicionado - ID: 50 | Conteúdo: z

```

- [x] Desenvolveu lab (teste) para implementar a leitura do banco de dados
  - [x] Aprendeu importar e usar Sqlite em Go
  - [x] Entendeu as Goroutines e Loops infinitos
  - [x] Fez o uso de Observer/Watcher de arquivos para ler o WAL de um banco de dados, para criar "gatilhos"
  - [x] Planejou a estrutura do código que será executado no Agente (antes de implementar o Redis)
  - [x] Resolveu uma Race Condition
  - [x] Testou manualmente conexão com banco de dados via ADB
  - [x] Resolveu Chatters (logs repetitivos) que aconteciam por serem executados no lugar errado dos loops

> commit: c730008 test(agent): simula triggers de observers em estrutura semelhante ao WhatsApp

---

##### Mudando para Ambiente Android (identificando limitações)

Os testes anteriores foram feitos na máquina local, simulando estrutura semelhante ao do WhatsApp, conhecida até então. Ao realizar os testes em ambiente Android (Linux), me deparei com as seguintes limitações:

1. **Permissão de root** - O agente sofre de restrições em pastas do sistema, sendo visto como processo não prioritário.
   - Para contornar isso, o agente deve ser executado em modo de superusuário `su`, ou pode ser configurado como daemon/serviço do sistema com `magisk`.

2. **Hierarquia de permissões** - Para o banco de dados em um Android, dificilmente duas atividades podem acessar o banco de dados com permissão integral (edição), a menos que uma delas esteja apenas em readonly (apenas leitura).
   - O agente se conecta no banco agora com flags de acesso `readOnly` para apenas leitura e `Syncronous` para respeitar a ordem de I/O do Banco de Dados, e aguardar o registro ser devidamente salvo.
   - Detecção de alteração no WAL -> apenas observa o banco de dados -> Aguarda regisro ser salvo para disparar consulta SQL em seguida

3. **Dooze Mode** - o sistema precisa reconhecer o processo como um serviço prioritário e evitar "matar" ele.
   - Através do `oom_score_adj`, o próprio agente se configura como serviço de alta prioridade, então, quando o Android estiver economizando recursos (OMM Killer), o Agente não vai ser encerrado.

4. **Arquitetura de Abstrações de baixo consumo** - O Kernel do Android é otimizado para economizar bateria, então alguns eventos não são processados em tempo, salvando registros novo em cache para economizar CPU e bateria.
   - Para resolver isso, é adicionado um time.Tick que "cutuca" o banco de dados, e faz o Watcher disparar a requisição a procura de novos registros.

5. **Identificando o agente como um Daemon** - O comportamento do agente é equivalente ao de _serviços de background_ (Daemons: processos que operam no _espaço do usuário_ de forma assíncrona, gerenciando recursos e respondendo a eventos do sistema). -> Criei um `deploy.bash` para ter uma _"execução de disparo único"_ que atualiza o daemon, envia para o Android e configura as permissões, e em seguida exectua via adb root.

O código foi devidamente adaptado, agora o agente consegue se comportar exatamente como um serviço nativo do sistema, similar ao anteriormente testado no ambiente Windows. Durante os testes entre um dia e outro, percebi que o Go, ao fazer build do binário para Android, me forçava a linkar o NDK, mas em nenhum momento do meu código até então eu fiz uso, de fato, de nenhuma biblioteca do NDK. Embora eu tenha conseguido buildar com ele uma vez, o Agente funciona perfeitamente sem linkar com o NDK, o que culminou na quinta etapa (após pesquisar sobre, e revisar o código com uma IA externamente (não usei agentes/gemini/cloud code), a arquitetura do meu binário é o mesmo de um daemon, e o objetivo da revisão foi apenas saber se sem o NDK o código poderia quebrar em algum momento que eu não estivesse vendo).

- [x] Criou PoC (Proof of Case) para implementação do Agente/Daemon de monitoramento
  - [x] Identificou limitações do SO Android
  - [x] Pesquisou e testou soluções durante desenvolvimento do PoC
    - [x] Agente/Daemon não conseguia acompanhar as atualizações do Banco de Dados no Android
      - [x] Resolveu usando flag de acesso `readOnly` e `Syncronous`
    - [x] Pesquisou que outras limitações o Android impõe ao Agente (Definição de prioridade)
      - [x] Adicionou time.Ticks (polling) como fallback dos observers (evita que observers "durmam")
      - [x] Adicionou `oom.score.adj` para aumentar a prioridade do agente/daemon

> Commit 9297764 test(agent): adapta e testa agente em ambiente android para identificar limitações

---

### Etapa-003: Implementação do Redis e Modelagem de Dados

#### Fase-1: Implementação do Redis

Como nunca tive experiência com Redis, fiz uma pesquisa e olhei a documentação oficial: achei fácil, mas o problema do desafio pedia por algo que não estava explicito em nenhum tutorial ou parte da documentação - ao menos de início. Pensando em "encurtar o caminho" (no bom sentido), me aproveitei da maleabilidade da função Modo IA da pesquisa do Google para "produzir meu próprio tutorial", em uma sessão eu consegui aprender e consolidar teoria e prática de poncorrência e paralelismo, publisher e subscribers, e producers e consumers, além do CRUD básico de Redis. Não muito diferente do que já fiz com Go até agora.

- [x] Instalou Redis
- [x] Fez o primeiro CRUD com Redis
- [x] Aprendeu usar string hashes (para o desafio)
- [x] Fez a primeira estrutura de Pub/Sub
- [x] Fez a primeira esturua Producer/Consumer

> Commit: ddea034 test(redis): Executa primeiro CRUD com Redis, e estrutura pedida pelo desafio

#### Fase-2: Modelagem de Dados

Quando interagi a primeira vez com a tabela `message` do banco de dados do WhatsApp, percebi que "nem todos os dados estão lá", quase deixei um detalhe despercebido passar: o WhatsApp usa mais de um banco de dados para guardar informações. Em resumo, até dá para monitorar as mensagens do WhatsApp pela tabela, mas algumas informações que podem ajudar na identificação estão em outros bancos. Parte desse processo fiz com mais interação com IA, onde, através de tentativa e erro, realizei algumas queries SQL no banco de dados do WhatsApp (vale ressaltar que tenho conhecimentos básicos em SQL, mas ainda peco em joins, quando vi que ia ser necessário trabalhar com eles, resolvi "apelar" pro meu mentor virtual, ou seja, mesmo sem saber trabalhar com joins, eu consegui prever que precisava deles, por isso que essa parte etapa foi inteiramente no terminal), e identifiquei que vou precisar trabalhar com o `msgstore.db` e com `wa.db`.

Após encontrar a "query perfeita", fiz um trabalho investigativo para consolidação do conhecimento, o que resultou no quadro do projeto `WhatsAppSync Excalidraw.png`.

Ao tentar adaptar a query no Agente, me deparei com mais limitações, dessa vez com o Go: o Go é quem gerencia as conexões entre demais bancos de dados, o NDK era necessário, e a query pecisava de um pequena adaptação. Relatei todo o drama em `dificuldades.md`, mas em resumo, precisei de mais indução de IA para entender o que estava acontecendo.

- [x] Modelou os dados da base de dados do WhatsApp para a aplicação
  - [x] Investigou "pontos cegos" na interpretação das queries
  - [x] Aprendeu como funciona as JOINs
- [x] Corrigiu agente
  - [x] Corrigiu `pullMessages` para trabalhar com a query correta
  - [x] Consolidou processo de build e deploy do agente em um script bash
  - [x] Corrigiu a falta do NDK e referências para montar a build
  - [x] Configurou o agente para gerar o deploy sem precisar carregar o arquivo .sql junto `go://embbeded`

> Commit: 3ce8cb6 test(database): modela dados vindos do whatsapp e adpta agente

Sem dificuldades, implementei o push de mensagens para o Redis, revisarei essa etapa antes de entregar o desafio. O diferencial foi que entendi melhor como funcniona as Structs, e as usei para criar o DTO o JSON e String Hash da mensagem pra ser enviada para o Redis (apenas a String Hash).

- [x] Adaptou o agente para trabalhar com Redis
  - [x] Aprendeu criar JSONs, DTOs e Structs no Go
  - [x] Conseguiu enviar as mensagem em tempo real para o Redis
  - [x] Pré-configurou o agente para produzir JSONs (talvez seja descartado, ou implementado)

> Commit: 1407607 test(Redis): Adapta agente para salvar e publicar as mensagens no Redis

---

## Etapa-004: Aplicação Externa

#### Decisão-Técnica-002: Usar IA como Agente de programação para criar base da aplicação:

Nesta etapa, eu decidi que iria implementar uma **API REST** para consumir o Redis, e interagir com ele em tempo real. **Usei IA, e dessa vez como agente**, para auxiliar na criação da **base da API** (ele gerou com alguns erros, o que é perfeito para eu aprender corrigindo o código) em .NET Minimal API. Para contornar a dívida técnica/cognitiva, gerei com ele uma documentação de revisão da arquitetura (`TECH_GUIDE.md`), que serve como um guia técnico para entender as implicações e o funcionamento do ecossistema .NET e Next.js. Essa etapa só será entregue após a conclusão dos testes, que dependem do entendimento a estrutura gerada pelo agente, e por ajustes necessários para o objetivo final do desafio. Para o frontend, aproveitei que pedi ao agente para criar a base da API, o pedi para que desenvolvesse o front-end com a estrutura que eu já tinha desenvolvido com as Provas de Conceito em /lab, e eu o corrigi pois encontrei alguns erros - mais um processo de entendimento de Typescript/Next.js para o desafio, mas não foquei em me aprofundar em front-end.

#### Decisão-Técnica-003: Conteinerizar serviços:

Percebendo o aumento de tecnologias no projeto (C#, Go, Javascript/Typescript, Redis), resolvi conteinerizar os serviços para facilitar o desenvolvimento, e evitar possíveis conflitos entre dependências, versionamentos, e etc. Contei com IA para configurar Dockerfile do back-end, pois **nunca havia conteinerizado uma aplicação .NET** e corrigir os docker-composes.

### Back-end: Teste

Por ser familiarizado com Java e ter experiência mínima com C++, muito da sintaxe do C# é familiar para mim (Família C).

Na máquina local:

- [x] Instalou .Net
  - [x] Entendeu como funciona o .Net
    - [x] Fez o primeiro app em C# (console app)
      - [x] Imprimiu a mensagem "Hello World" no terminal
      - [x] Entendeu como funcionam propriedades e métodos
      - [x] Entendeu como funcionam herança
      - [x] Entendeu como funcionam polimorfismo
      - [x] Entendeu como funcionam encapsulamento
    - [x] Entendeu como funciona o .Net Minimal API
  - [x] Implementou API em .NET Minimal API
    - [x] Entendeu os passos da criação de uam API em Minimal API
      - [x] Montou um buider simples
      - [x] Configurou o CORS
      - [x] Montou a aplicação
      - [x] Criou endpoints GET
      - [x] Criou endpoints POST para o Redis
    - [x] Implementou Redis Pub/Sub
      - [x] Entendeu como funciona o Redis Pub/Sub no Minimal API
      - [x] Entendeu como funciona o processamento assíncrono das Tasks do Async/Await
      - [x] API de testes é capaz de salvar mensagens no Redis (Push)
      - [x] API de testes é capaz de carregar mensagens do Redis (Pull)
      - [x] API de testes é capaz de receber mensagens do Redis (Sub)
      - [x] API de testes é capaz de enviar mensagens para o Redis (Pub)
    - [x] API de testes simula o endpoint /contacts do desafio
      - [x] API de testes é capaz de criar um contato no Redis (POST)
      - [x] API de testes é capaz de buscar um contato no Redis (GET)
    - [x] Organizou a arquitetura da API para proximidade da base gerada com IA (arquitetura em Camadas/Modular)
    - [x] Entendeu como funciona SSE (Server-Sent Events) para interagir com Redis em tempo real.
    - [x] Aprendeu a usar DTOs em C#.
    - [x] Revisou e ajustou código gerado pelo agente

### Back-end: Implementação

A base da API foi gerada com IA e revisada manualmente. Os principais ajustes necessários para o ambiente Docker com Next.js e Redis foram:

- [x] Corrigiu conflito de versões no `Dockerfile`: build usava `sdk:10.0` mas runtime estava em `aspnet:8.0` — corrigido para `aspnet:10.0`
- [x] Corrigiu incompatibilidade de nomenclatura: C# serializa em `camelCase` (`nomeConversa`), frontend esperava `snake_case` (`nome_conversa`) — corrigido nos tipos TypeScript e adicionado alias no DTO
- [x] Adicionou campo `receivedAt` ao backend (o frontend esperava o campo, a IA não havia gerado)
- [x] Adicionou **Heartbeat SSE (ping a cada 15s)**: sem ele, o Nginx encerrava a conexão por inatividade com erro `499` após 60 segundos
- [x] Corrigiu bug de desinscrição coletiva no Redis: fechar uma aba cancelava o canal de **todos** os clientes — simplificado para `Timeout.Infinite` com `RequestAborted`
- [x] Corrigiu dupla inscrição no canal Redis que duplicava as mensagens no feed
- [x] Expôs porta `6379:6379` do Redis no `docker-compose.yml` para o Agente Android alcançar o host via `10.0.2.2`
- [x] **Tradução Avançada LID -> JID no Android**: Redigiu a query SQL para consulta no banco de dados do WhatsApp com múltiplas camadas de `COALESCE` para cruzar dados entre `msgstore.db` e `wa.db`.

### Modelagem de Dados e SQL Avançado

A extração de dados foi readaptada para lidar com a complexidade do esquema interno do WhatsApp (LID - Linked Identity), que mascara números de telefone reais, depois que incluí um terceiro - não salvo - número nos testes:

- [x] **Resolução de Identidade (LID to PN)**: Criou uma ponte de mapeamento via tabela `status_ranking` para traduzir IDs internos do WhatsApp (`@lid`) em números de telefone legíveis (`@s.whatsapp.net`).
- [x] **Hierarquia de Nomes (6 Camadas)**: Estabeleceu uma cadeia de prioridade para exibir o nome mais legível possível:
  1. Nome da Agenda (`display_name`)
  2. Nome via Mapeamento LID
  3. Nome Verificado (Business)
  4. Nome de Perfil (`wa_name`)
  5. Nome de Perfil via LID
  6. Fallback ("Desconhecido")

- [!] **Desafio Técnico: Bruteforce de Identidade**: Durante os testes com um terceiro número (não salvo), observou-se uma inconsistência na arquitetura do WhatsApp. Enquanto alguns contatos permitem o mapeamento LID -> PN (Phone Number), outros permanecem restritos apenas ao LID.
- [!] **Conclusão de Segurança**: Identificou-se que o WhatsApp está transitando para uma arquitetura mais segura onde o número de telefone pode estar ausente das tabelas locais (`wa_contacts` / `status_ranking`) para contatos fora da agenda, impossibilitando a resolução do número real via SQL local em 100% dos casos.

### Front-end: Refinamento e UX

O painel visual foi refinado com agente para garantir uma experiência sem erros de execução:

- [x] **Centralização Vertical do Dashboard**: Ajustou o grid principal (`lg:grid-cols-12`) com `items-center` para balancear visualmente a sidebar de estatísticas com o feed de mensagens.
- [x] **Correção de Hydration Mismatch (Next.js)**: Solucionou erro crítico de hidratação onde `Math.random()` gerava valores divergentes entre servidor e cliente. Implementou o padrão `isMounted` para garantir renderização segura de métricas simuladas.
- [x] **Micro-animações de Tráfego**: Adicionou feedback visual progressivo no "Monitor de Tráfego" para simular a atividade do Agente Android em tempo real.

### Resumo dos Marcos

s- **Infraestrutura**: A conteinerização via Docker unificou as tecnologias (.NET, Go, Next.js, Redis), eliminando conflitos de ambiente e garantindo portabilidade total do desafio.

- **Backend Robustez**: A implementação da API em .NET com Redis Pub/Sub e SSE transformou o fluxo de dados em um stream de baixa latência, com mecanismos de resiliência como Heartbeats e gestão de conexões.
- **Engenharia de Dados**: A superação do desafio de identidades (LID vs JID) via SQL avançado, permitindo a extração de nomes e números mesmo em uma arquitetura de banco de dados fragmentada.
- **Interface de Usuário**: A criação de um Dashboard responsivo que não apenas exibe dados, mas simula a telemetria do agente, proporcionando uma visão operacional em tempo real.

---

### Etapa-005: Reflexão e Triangulação Técnica

Esta etapa final consistiu em um processo de "espelhamento" com o agente de IA para nomear as práticas que apliquei instintivamente ao longo do desafio. Como um desenvolvedor que nunca cursou graduação, minha base é 100% fruto de autoconhecimento e uma jornada de 7 anos de pesquisas e estudos autodidatas (com alguns hiatos e meses recentes de imersão intensa).

#### Identificação de Padrões (Prática vs. Teoria)

Através da análise técnica do agente de IA sobre o código gerado, foi possível identificar que apliquei padrões de arquitetura de nível sênior por pura intuição técnica e necessidade lógica, mesmo sem conhecer seus nomes acadêmicos:

- [x] **Spike & Stabilize (PoC-Driven)**: O uso sistemático da pasta `/lab` para validar hipóteses isoladas (Redis, Observers, Android Root) antes da integração no código de produção.
- [x] **Strangler Fig Pattern**: A evolução do agente em paralelo à base original, substituindo e absorvendo funcionalidades de forma incremental e segura.
- [x] **Debounce Pattern**: O uso de timers (`time.AfterFunc`) para mitigar o "chattering" (múltiplas notificações) do sistema de arquivos Android, evitando consultas redundantes.
- [x] **Cursor-Based Pagination**: A lógica de busca via `lastProcessID` (cursor) em vez de offset, garantindo integridade cronológica e evitando duplicidade de mensagens.
- [x] **Graceful Degradation**: O sistema híbrido de Watcher + Polling Ticker que garante o funcionamento do sync mesmo sob limitações severas de economia de energia do Android (Doze Mode).
- [x] **DTO & Separation of Concerns**: A distinção clara entre structs de banco de dados (`MessageRow`) e objetos de transferência (`MessageDTO`) para isolar a camada de persistência da camada de transporte.

**Sinceridade Técnica**: Durante a entrevista, utilizarei da minha sinceridade para explicar que esses nomes foram identificados nesta etapa de reflexão com a IA. Isso prova que a prática e a habilidade de pesquisa (7 anos de autodidatismo) me levaram a implementar soluções de engenharia robustas que a academia apenas formaliza.

---

As etapas processo foram devidamente organizadas em um quadro Kanban usando Trello: https://trello.com/b/SuVJxaAJ/desafio-intelitrader
