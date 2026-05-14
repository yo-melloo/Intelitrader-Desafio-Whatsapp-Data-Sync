import redis
import json
import time
import random
import os

print("🔥 Iniciando Chaos Tester (Filtro 3)...")

# Aguarda o Redis iniciar
time.sleep(5)

# Conecta ao Redis
redis_host = os.environ.get("REDIS_HOST", "redis")
r = redis.Redis(host=redis_host, port=6379, decode_responses=True)

try:
    r.ping()
    print(f"✅ Conectado ao Redis em {redis_host}:6379")
except Exception as e:
    print(f"❌ Falha ao conectar ao Redis: {e}")
    exit(1)

def disparar_contato(nome, numero, cenario):
    payload = {"name": nome, "number": numero}
    r.publish("contacts:insert", json.dumps(payload))
    print(f"[{cenario}] Disparado: {nome} - {numero}")

print("\n--- TESTE 1: A Bomba de Carga (Stress) ---")
for i in range(1, 51):
    disparar_contato(f"Clone {i}", f"551199999{i:04d}", "STRESS")
    time.sleep(0.05) # Disparo super rápido

time.sleep(2)

print("\n--- TESTE 2: O Contato Fantasma (Invalid Data) ---")
# Missing name
r.publish("contacts:insert", json.dumps({"number": "5511900000000"}))
print("[FANTASMA] Disparado: Sem nome")
# Missing number
r.publish("contacts:insert", json.dumps({"name": "Sem Numero"}))
print("[FANTASMA] Disparado: Sem numero")
# Empty strings
disparar_contato("", "", "FANTASMA")

time.sleep(2)

print("\n--- TESTE 3: Injeção SQL/Shell (Security) ---")
disparar_contato("Douglas' --", "123456", "SECURITY")
disparar_contato("$(reboot)", "123456", "SECURITY")
disparar_contato("Pwned\"; rm -rf /", "123456", "SECURITY")

time.sleep(2)

print("\n--- TESTE 4: Buffer Overflow (Limites) ---")
nome_gigante = "A" * 5000
disparar_contato("Overflow", nome_gigante, "OVERFLOW")

print("\n✅ Testes de caos finalizados. Verifique os logs do Agente Go no emulador!")
