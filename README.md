# VD Stats (v2)

Uma plataforma em tempo real para monitoramento de servidores (VPS), containers Docker e fluxo de balanceamento de carga (NGINX).

O sistema usa uma arquitetura de extração via SSH, agregação em memória no Go (redução de gargalo de banco de dados) e armazenamento no PostgreSQL para retenção de histórico. O Frontend (React/Vite) consulta essa API para exibição em painéis dinâmicos.

## Fluxo e Arquitetura do Projeto

1. **Backend (Go Engine):**
   - Estabelece conexões SSH seguras com os servidores VPS.
   - Executa comandos não-bloqueantes para extrair métricas de CPU, Memória e Disco de processos, containers (Docker) e NGINX (Logs de acesso).
   - Mantém agregação em memória (Mutex lock) para não sobrecarregar o banco com inserções contínuas, gravando lotes consolidados no PostgreSQL.
   - Fornece endpoints REST para o Frontend consumir métricas consolidadas e históricas.

2. **Frontend (React + Vite + Tailwind):**
   - Consome a API Go para alimentar gráficos dinâmicos de consumo de CPU, RAM e Disco de Containers.
   - Mapeia ativamente o fluxo e distribuição do Load Balancer em um painel específico.

3. **Database (PostgreSQL):**
   - Armazena metadados e séries temporais com alto desempenho de leitura.

## Como Executar (Ambiente de Desenvolvimento)

### 1. Configurar Variáveis de Ambiente
Crie um arquivo `.env` na raiz ou exporte as seguintes variáveis:
```bash
export DATABASE_URL="host=localhost user=SEU_USER password=SUA_SENHA dbname=vd_stats port=5432 sslmode=disable TimeZone=UTC"

# Servidores Alvo
export VPS1_IP="0.0.0.0"
export VPS2_IP="0.0.0.0"
export LB_IP="0.0.0.0"

# Autenticação
export SSH_USER="root"
export SSH_KEY_PATH="~/.ssh/sua_chave_aqui"
```

No Frontend, crie um arquivo `frontend/.env`:
```env
VITE_VPS1_IP="IP DA VPS 1"
VITE_VPS2_IP="IP DA VPS 2"
VITE_LB_IP="IP DO LOAD BALANCER"
```

### 2. Rodando o Backend
```bash
cd backend
go mod tidy
go run cmd/vd_stats/main.go
```

### 3. Rodando o Frontend
```bash
cd frontend
npm install
npm run dev
```

## Como funciona a extração dos dados (Load Balancer)
O backend conecta-se via stream SSH (`tail -f`) ao log de acesso (`access.log`) do NGINX do Load Balancer. Cada linha recebida é processada com base nas variáveis do bloco upstream:
* `upstream_addr`: Identifica para qual servidor de processamento o tráfego foi roteado. Se não houver processamento externo (ex: cache de estáticos, redirects), é capturado como tráfego Local.
* `server_name`: Identifica o projeto / domínio do request (`$host` config no Nginx).

*Este repositório foi otimizado e formatado para uso open-source.*
