# vd_stats

Exibe todos os containers agrupados pela pasta do projeto, com uso de CPU em tempo real, memória (em GB) e uso de disco por projeto — tudo a partir da sua máquina local, sem necessidade de agentes no servidor.

```
╭─────────────────────────────────────────────────╮
│ VPS: my-server  |  14:32:01                     │
╰─────────────────────────────────────────────────╯
 Container                    CPU      Memória     Disco
 ────────────────────────────────────────────────────
 📁 /opt/sgo                                       2.6M
    sgo-web-1                0.03%    0.128 GB
    sgo-redis-1              8.67%    0.003 GB

 📁 /srv/my-app                                    305M
    app-web                  3.29%    0.118 GB
    app-worker               0.53%    0.300 GB
    app-redis                1.24%    0.004 GB
    app-beat                 ✗ Down   (exited)
╭─────────────────────────────────────────────────╮
│ Ctrl+C / Ctrl+D para sair  ↑↓ para rolar       │
╰─────────────────────────────────────────────────╯
```

**Recursos**
- Agrupa containers pela pasta do projeto `docker-compose`
- Mostra uso de CPU, memória (GB) e disco por projeto
- CPU codificada por cores: verde < 5%, amarelo 5–20%, vermelho > 20%
- Containers que estão fora do ar aparecem em vermelho com seu status
- Atualização automática a cada 2 segundos
- Rolagem quando o conteúdo excede a altura do terminal (setas ↑↓)
- Sair com `Ctrl+C` ou `Ctrl+D`

---

## Requisitos

- Python 3.8+
- Acesso via chave SSH à sua VPS (autenticação por senha não é suportada)
- Docker rodando na VPS (nenhuma configuração necessária no lado do servidor)

---

## Instalação

### Linux / macOS

**1. Clone o repositório**

```bash
git clone https://github.com/seu-usuario/vd-stats.git
cd vd-stats
```

**2. Configure sua conexão VPS**

```bash
cp vd_stats/config.example.py vd_stats/config.py
```

Edite `vd_stats/config.py` com seus valores:

```python
SSH_HOST = "ip-da-sua-vps"
SSH_USER = "root"
SSH_KEY_PATH = "~/.ssh/id_rsa"   # caminho para sua chave privada
SSH_PORT = 22
REFRESH_INTERVAL = 2             # segundos entre as atualizações
```

**3. Execute o instalador**

```bash
bash install.sh
```

Isso cria um ambiente virtual em `~/.venvs/vd_stats` e registra o comando `vd_stats` em `~/.local/bin`.

**4. Certifique-se de que `~/.local/bin` esteja no seu PATH**

Verifique com:
```bash
echo $PATH | grep -q ".local/bin" && echo "ok" || echo "adicione ao PATH"
```

Se não estiver lá, adicione ao seu `~/.bashrc` ou `~/.zshrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Depois recarregue: `source ~/.bashrc`

**5. Executar**

```bash
vd_stats
```

---

### Windows (via WSL)

O recurso de rolagem depende de `tty` e `termios`, que são módulos exclusivos para Unix. Terminais nativos do Windows não são suportados. A abordagem recomendada é o **WSL (Windows Subsystem for Linux)**.

**1. Habilite o WSL** (se ainda não estiver habilitado)

Abra o PowerShell como Administrador e execute:
```powershell
wsl --install
```

Reinicie sua máquina quando solicitado.

**2. Abra um terminal WSL e siga as instruções do Linux acima**

Todos os passos são idênticos dentro do WSL. Suas chaves SSH do Windows estão acessíveis em `/mnt/c/Users/SeuNome/.ssh/`.

Você pode apontar o `SSH_KEY_PATH` para:
```python
SSH_KEY_PATH = "/mnt/c/Users/SeuNome/.ssh/id_rsa"
```

---

## Referência de configuração

| Variável | Descrição | Padrão |
|---|---|---|
| `SSH_HOST` | IP ou hostname da sua VPS | — |
| `SSH_USER` | Usuário SSH | `root` |
| `SSH_KEY_PATH` | Caminho para sua chave SSH privada | `~/.ssh/id_rsa` |
| `SSH_PORT` | Porta SSH | `22` |
| `REFRESH_INTERVAL` | Segundos entre as atualizações de dados | `2` |

---

## Como funciona

1. Conecta-se à VPS via SSH usando [paramiko](https://www.paramiko.org/)
2. Executa `docker ps -a` e `docker inspect` para coletar nomes de containers, status e suas pastas de projeto (através da label `com.docker.compose.project.working_dir` definida pelo Docker Compose)
3. Executa `docker stats --no-stream` para obter CPU e memória em tempo real por container em execução
4. Executa `du -sh` em cada pasta de projeto para uso de disco
5. Agrupa tudo por pasta e renderiza o dashboard usando [rich](https://github.com/Textualize/rich)

Containers não gerenciados pelo Docker Compose (sem label de projeto) aparecem sob um grupo chamado **Outros**.

---

## Atualizando

```bash
cd vd-stats
git pull
bash install.sh
```

---

## Licença

MIT
