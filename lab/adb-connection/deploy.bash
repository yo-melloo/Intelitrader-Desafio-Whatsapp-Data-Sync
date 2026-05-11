# Caso use o .bash para atualizar e inicializar o agente:
adb root

# Atualiza build:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o agent-sync

# Deploy para o Android:
adb push agent-sync //data/local/tmp/  

# Modifica permissões e executa binário via ADB:
adb shell "chmod +x /data/local/tmp/agent-sync && su root /data/local/tmp/agent-sync"