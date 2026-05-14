## Dificuldade 01: Linguagem de programação nova

Eu sou um dev Java e Python, o único framework que eu sei, até o momento, é Spring Boot - eu até estava tentando aprender Django, mas dado os projetos que eu tinha em mente, eu deixei pra depois. Ao ver os requisitos do projeto, já identifiquei minha primeira dificuldade (não obstáculo): aprender uma linguagem de programação nova, dentre elas, Golang estava na minha lista de "Aprender mais tarde", só que "a hora é agora", eu pensei. Como resolvi isso? Peguei tutoriais curtos de Go, pois como sei lógica de programação (e todo dev deveria), é fácil reconhecer semelhanças estruturais de cada línugagem, principalmente por que a achei parecida com Python, além disso, também usei IA para sanar dúvidas (estilo mentor e aluno) sobre pontualidades da linguagem - como as tão conhecidas "Goroutines". Eu até pensei em usar C++, por ter um projeto que (mais tarde) vai usar a linguagem [tenho um projeto pessoal de programação embarcada].

O problema: Linguagem requisitada pelo desafio
Solução: Abordagem de aprendizado pragmático

---

## Dificuldade 02: Race Condition

O Golang é muito rápido. Enquanto eu testava uma lógica para o uso de Observers, eu notei que os índices de pesquisa ao Banco de Dados estavam "retrógados", testei também a inserção de valores repetidos e em velocidades diferentes, e o comportamento foi o mesmo, o ID não acompanhava a mudança do banco de dados, mesmo que o Observer disparasse com a mudança pois o Agente é muito rápido (ele chegava no banco de dados antes da mudança ser salva). Então eu pesquisei uma forma de adicionar uma fila, até pensei em delay em primeiro momento, mas pensei que talvez não fosse a melhor solução (e foi). Apanhei por horas, mas com pesquisa consegui identificar até erros que causei no código por conta de fadiga.

O problema: O agente chegava no banco de dados antes dos dados
Solução: criar uma lógica de fila + delay

---

## Dificuldade 03: Ambiente Android

Ser um SO baseado em Linux não significa que só conhecer Linux é suficiente pra saber lidar com Android, o sistema "me colocou de joelhos sobre sementes de milho" ao ver que meu agente programado nos testes anteriores não se comportou como o esperado na primeira tentativa, até aí tudo bem. Fiz um trabalho minuscioso de tentativa e erro, e com muita pesquisa. O Ambiente Android é bastante hostil a qualquer coisa que não for pré-definida pelo próprio sistema, pois ele tem restrições de acesso, abstrações de armazenamento (SQLite VFS, que o Android usa para estratégia de Write-back Cache: O registro só entrava de fato no banco de dados, quando o sistema passava a ficar ocioso ou houvesse demanda por memória RAM, onde de fato estavam os registros novos), e gerenciadores robustos de recursos como o OOM Killer (que "assassina" processos de baixa prioridade quando o sistema começa pedir arrego)

O probelma: Limitações e peculiaridades do Sistema Android
Solução: Readaptar a lógica do Agente sob a lógica base para funcionamento dentro do Android

---

## Dificulade 04: Material de estudo para Redis + Golang específicos para o desafio

Por incrível que pareça, eu não tive muitas dificuldades nesse processo, o que quis me travar no começo foi não encontrar tutoriais específicos para essa tarefa (como eu, sem conhecimentos de Go e Redis vou desenvolver uma implementação entre esses dois em um curto tempo?), anos atrás essa tarefa seria coisa de louco, eu até estava planejando usar um Agente de IA para realizar essa etapa com auxilio diretamente sobre o código, mas isso ia abrir brecha pra dívida técnica, foi aí que me lembrei que meu próprio Google está com modo de pesquisa com IA ativado, signifca que se eu pesquisar algo específico, ou realizar uma pesquisa simples e "induzir" a IA chegar no assunto que eu desejo, ela pode acabar agilizando o processo de pesquisa, que sem ela, poderia durar horas ou dias... E foi o que eu fiz: Enquanto gerava o agente de IA, ele me soltou uma palavra chave "Producers e Consumers", que quando pesquisei no modo de IA, serviu direitinho para fazer o drift até o exemplo de código em que me baseei para o desafio.

O problema: material de estudo não resolve a especifidade do projeto em curto prazo
Solução: polir pesquisa apenas para o necessário exigido para o desafio

---

## Dificuldade 05: SQL Joins

Não tive contato muito prático com Joins na minha pouca experiência, isso culminou em eu ter que, mais uma vez, ter um brainstorming com IA e fazer uma "investigação" no banco de dados do WhatsApp, como relatado em `processo.md`. Contornei essa dificuldade com essa "mentoria", mas sei muito bem que qualquer pessoa que delega estudos para IA permite que se crie uma "dívida técnica/cognitiva", coisa que eu mitiguei com um "teste investigativo": É possível encontrar `teste-sql-joins.md` aqui na pasta de documentação, ele foi um "dever de casa" que pedi para IA fazer, que respondi com minhas próprias palavras, e deletei e refiz manualmente os comentários na query com minha versão do entendimento, então acredito que essa parte da dificuldade eu tratei devidamente como um "autodidata" deveria o fazer.

---

## Dificuldade 06: Runtime Go

Durante o processo de implementação da query SQL no Agente, ao gerar o binário/daemon ficou provado que para a arquitetura Android x86_64, o Go exige o CGO_ENABLED=1 e o linker externo do NDK para lidar com o gerenciamento de memória (o Thread Local Storage (TLS) - especialmente as páginas de 16 KB do Android 17). Isso significa que, tive que rever o uso do NDK que, interpretei como desnecessário graças ao Go, para Necessário justamente por usar o Go. Veio a calhar pois um erro de interpretação da minha parte quase me faria forçar o agente ser independente dele.

O problema: A arquitetura x86_64 no Android 17 exige o linker do NDK para compatibilidade com o alinhamento de memória TLS.

A solução: Gerei um script para gerar o binário com CGO_ENABLED=1 e flags de linker específicas corrigiu o erro `android/amd64 requires external linking`

---

## Dificuldade 07: Consumo da Query SQL

Ainda durante o processo de implementação da query SQL no agente, ao rodar o daemon, a sincronia em tempo real parou. A começar que eu esqueci de tirar o limitador da query (LIMIT 20), e adaptar todo o resto. Enquanto discutia com a IA sobre o que estava acontecendo no código (eu já estava começando achar que quebrei o código todo), identificamos que se tratava de uma particularidade do Go: ele não processa o `ATTACH DATABASE` no começo da query, e não estava conseguindo injetar o lastProcessID nela.

O problema: O Go estava sendo redundante em tentar abrir várias conexões dos bancos de dados cada vez que o Watcher via um evento (pois ele disparava a query toda vez e toda vez ele anexava uma nova conexão do banco de dados), e o agente não lia/escrevia as mensagens.

A solução: Gerenciar a conexão no próprio agente, reconfigurar o daemon, e readptar a query para injeção da variável `lastIdProcess`

> Thread Local Storage (TLS) é método de gerenciamento de memória que permite a cada thread em um processo multithread a ter sua própria cópia de dados.

---

## Dificuldade 08: Aplicação externa

Como dev Java, eu não conhecia C# .NET, apesar de já ter tido uma rasa experiência a muitos anos, sem sucesso nas minhas tentativas. Ao ver os requisitos do projeto, já identifiquei como uma barreira a ser supéra: aprender não só uma linguagem de programação novas, como duas (Go e C#). Da mesma maneira que fiz com Go, farei com C#, ressaltando que parte do processo foi feita com auxílio de IA (como documentado em `processo.md`), voltado apenas para a resolução do projeto.

PS.: Fazer uma API REST com C# não foi um processo tão difícil, claro, tem suas particularidades. Acredito que um desenvolvedor Java (como eu) ou PHP conseguiria implementar algo assim em até 1h (dependendo do conhecimento prévio e da familiaridade com o Framework), e no meu caso, pesquisa e ação foi o caminho.

O problema: Barreira de novas linguagens de programação

Solução: Auxílio de IA, Rubber Duckering e tutoriais curtos de C# .NET

---

## Dificuldade 09: Estágio 1 de 2 do fim da aplicação

Desenvolvi os componentes necessários para o evento 1 da pipeline do desafio (Salvar e servir as mensagens do WhatsApp, no cliente - frontend - através do Redis), então simulei essa interação com serviços conteinerizados, as dificuldades que encontrei foram adaptabilidade das conexões entre os containers, e código gerado pelo agente no front-end.

O problema: quebras de protocolo por parte do agente:

- **Conflito de versões do runtime .NET**: O `Dockerfile` compilava o projeto com o SDK do .NET 10, mas o estágio de execução usava a imagem `aspnet:8.0`. O binário exigia o runtime 10 e o container só oferecia o 8, impedindo a inicialização.
- **Redis inacessível para o Agente Android**: A porta `6379` do container Redis não estava mapeada para o host. O Agente (no emulador Android, que usa `10.0.2.2` para alcançar o host) não conseguia estabelecer conexão.
- **Incompatibilidade de nomenclatura no JSON (CamelCase vs. snake_case)**: O backend C# serializava os campos em `camelCase` (ex: `nomeConversa`), mas a interface TypeScript do frontend definia os campos em `snake_case` (ex: `nome_conversa`). O Dashboard ficava vazio pois os campos nunca eram encontrados.
- **Campo `receivedAt` ausente no backend**: O frontend esperava um campo `receivedAt` no JSON que o backend nunca enviava, gerando inconsistência de tipo no TypeScript.
- **Timeout de 60 segundos na conexão SSE (erro 499)**: Sem mensagens novas, a conexão SSE ficava em silêncio e o Nginx/Windows a encerrava por inatividade, gerando o erro `Erro no EventSource: {}` no console do browser.
- **Bug de desinscrição coletiva no Redis**: Ao fechar o browser, o código executava `redis.Unsubscribe("general:hash-updates")`, cancelando a inscrição de **todos** os clientes conectados no canal, e não apenas do que havia desconectado.
- **Dupla inscrição no canal Redis**: Durante a depuração, uma chamada extra a `redis.Subscribe(handler)` foi acidentalmente introduzida, fazendo com que cada mensagem nova fosse processada e enviada duas vezes ao frontend.
- **Inconsistência de Serialização JSON (PascalCase vs camelCase)**: As mensagens enviadas via SSE eram serializadas manualmente sem definir uma política de nomes, resultando em chaves em `PascalCase` (`Id`, `Conteudo`). Como o frontend esperava `camelCase`, as mensagens novas apareciam sem conteúdo e causavam erro de chaves duplicadas no React (todas eram tratadas como `id: undefined`).

Solução: Depuração e testes com correções dos problemas identificados.

## Dificuldade 10: Etágio 2 de 2: Teste de Estresse e debugging com IA

E finalmente consegui desenvolver todos os componentes necessários para gerar o produto final do desafio, só que mais um teste precisava ser feito, e pra isso criei a pasta `integrated-solution` para essa finalidade:

- Investigar vulnerabilidades no código (SQL Injection)
- Avaliar boas práticas de programação (SOLID, Clean Code, etc.)
- Testar performance e estabilidade do Agente Go no Android (resiliente a OOM)
- Validar a estabilidade geral da aplicação (Conteineres dependentes de Redis e Nginx)

A dificuldade esteve em ter que escrever testes, e dessa vez falo de talvez testes unitários, integração... seja lá o que for. Como eu não tenho experiência escrevendo códigos de testes, minha idéia foi gerar um script com simulação de uso "malicioso" (tentativa de SQL Injection, e checagem de outras vulnerabilidades) e validar se a infraestrutura estava pronta de fato.

- O problema: validar brechas no código

- Solução: Gerar script de simulação e checagem.

---

## Considerações finais

O desafio está concluído, nessa "jornada de aprendizado", senti que minha dificuldade mais significativa foi a de domar funções assíncronas (Goroutines, Actions, e Tasks), que, até o momento não entraram muito bem na minha cabeça. Acredito que ao longo do tempo, e com a prática, essa dificuldade vai diminuindo. Sinto que hoje tenho uma visão muito mais clara de como funcionam as coisas. Fico feliz com o resultado final, e espero que ele possa ser aceito pelo pessoal da Intelitrader. :D
