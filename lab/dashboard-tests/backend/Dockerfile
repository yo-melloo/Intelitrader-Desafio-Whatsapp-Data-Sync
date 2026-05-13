# Estágio 1: Build
# Usa a imagem do .NET SDK 10.0 como base
# Define o diretório de trabalho dentro do container em /src
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

# Copia o projeto "./" e dependências "backend.csproj"
# Executa o comando dotnet restore dentro do container que baixa e instala todas as dependências do projeto
COPY ["backend.csproj", "./"]
RUN dotnet restore "backend.csproj"

# Copia o restante dos arquivos e compila
# Executa o comando dotnet publish dentro do container que compila o projeto e gera os arquivos necessários para rodar a aplicação em modo de produção
# A flag "/p:UseAppHost=false" garante que a aplicação seja executada usando o runtime do .NET, em vez de criar um executável nativo
COPY . .
RUN dotnet publish "backend.csproj" -c Release -o /app/publish /p:UseAppHost=false

# Estágio 2: Runtime
# A imagem aspnet:10.0 é menor que a sdk:10.0, e já contém tudo o que é necessário para rodar a aplicação
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app

# Copia os arquivos compilados do estágio de build para o runtime
COPY --from=build /app/publish .

# Porta padrão do ASP.NET
# Expõe a porta 5000 do container para o host
# O caractere "+" significa que o host pode ser qualquer um (local ou rede)
ENV ASPNETCORE_URLS=http://+:5000
EXPOSE 5000

# Comando que será executado quando o container for iniciado "dotnet backend.dll" - equivale a "./mvnw spring-boot:run" do Spring
ENTRYPOINT ["dotnet", "backend.dll"]
