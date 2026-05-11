## Dificuldade 01: Linguagem de programação nova

Eu sou um dev Java, o único framework que eu sei, até o momento, é Spring Boot - eu até estava tentando aprender Django, mas dado os projetos que eu tinha em mente, eu deixei pra depois. Ao ver os requisitos do projeto, já identifiquei minha primeira dificuldade (não obstáculo): aprender uma linguagem de programação nova, dentre elas, Golang estava na minha lista de "Aprender mais tarde", só que "a hora é agora", eu pensei. Como resolvi isso? Peguei tutoriais curtos de Go, pois como sei lógica de programação (e todo dev deveria), é fácil reconhecer semelhanças estruturais de cada línugagem, além disso, também usei IA para sanar dúvidas (estilo mentor e aluno) sobre pontualidades da linguagem - como as tão conhecidas "Goroutines". Eu até pensei em usar C++, por ter um projeto que (mais tarde) vai usar a linguagem [tenho um projeto pessoal de programação embarcada].

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
