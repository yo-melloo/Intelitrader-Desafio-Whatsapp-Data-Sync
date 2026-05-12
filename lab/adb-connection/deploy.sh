export CGO_ENABLED=1
export GOOS=android
export GOARCH=amd64
export CC="C:/Users/$(whoami)/AppData/Local/Android/Sdk/ndk/30.0.14904198/toolchains/llvm/prebuilt/windows-x86_64/bin/x86_64-linux-android35-clang"

echo "🔨 Compilando agente_sync para Android x86_64..."

go build -ldflags="-linkmode external -extldflags '-Wl,-z,common-page-size=16384 -Wl,-z,max-page-size=16384'" -o agente_sync

if [ $? -eq 0 ]; then
    echo "✅ Sucesso! Enviando para o emulador..."
    adb push agente_sync //data/local/tmp/
    #adb push query-copy.sql //data/local/tmp/      # não necessário quando //go:embed está em uso
    adb shell chmod +x //data/local/tmp/agente_sync
    echo '"🚀 Pronto para rodar: adb shell "cd /data/local/tmp/ && ./agente_sync"'
else
    echo "❌ Falha na compilação."
fi