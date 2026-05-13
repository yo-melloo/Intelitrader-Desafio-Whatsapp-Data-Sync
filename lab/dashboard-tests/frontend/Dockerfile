# - A imagem node:24.12.0-alpine é uma versão leve do Node.js baseada na Alpine Linux
# - É uma escolha popular para ambientes Docker devido ao seu tamanho reduzido e simplicidade
FROM node:24.12.0-alpine

# - Define o diretório de trabalho dentro do container em /app
WORKDIR /app

# - Copia os arquivos de dependências do projeto
COPY package*.json .

# - Executa o comando npm install dentro do container que baixa e instala todas as dependências do projeto
RUN npm install

# - Copia o restante dos arquivos do projeto para o diretório de trabalho
COPY . .

# - Expõe a porta 3000 do container (acessível via porta 80 do nginx)
EXPOSE 3000

# - Executa "npm run dev"
CMD ["npm", "run", "dev"]
