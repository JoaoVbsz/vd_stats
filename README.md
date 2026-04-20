# vd_stats

A terminal dashboard for monitoring Docker containers on a remote VPS via SSH.

Displays all containers grouped by project folder, with real-time CPU usage, memory (in GB), and disk usage per project — all from your local machine, no agent needed on the server.

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
│ Ctrl+C / Ctrl+D to exit   ↑↓ to scroll         │
╰─────────────────────────────────────────────────╯
```

**Features**
- Groups containers by their `docker-compose` project folder
- Shows CPU, memory (GB), and disk usage per project
- Color-coded CPU: green < 5%, yellow 5–20%, red > 20%
- Containers that are down appear in red with their status
- Auto-refresh every 2 seconds
- Scroll when content exceeds terminal height (↑↓ arrow keys)
- Exit with `Ctrl+C` or `Ctrl+D`

---

## Requirements

- Python 3.8+
- SSH key-based access to your VPS (password auth is not supported)
- Docker running on the VPS (no setup needed on the server side)

---

## Installation

### Linux / macOS

**1. Clone the repository**

```bash
git clone https://github.com/your-username/vd-stats.git
cd vd-stats
```

**2. Configure your VPS connection**

```bash
cp vd_stats/config.example.py vd_stats/config.py
```

Edit `vd_stats/config.py` with your values:

```python
SSH_HOST = "your-vps-ip"
SSH_USER = "root"
SSH_KEY_PATH = "~/.ssh/id_rsa"   # path to your private key
SSH_PORT = 22
REFRESH_INTERVAL = 2             # seconds between refreshes
```

**3. Run the installer**

```bash
bash install.sh
```

This creates a virtual environment at `~/.venvs/vd_stats` and registers the `vd_stats` command in `~/.local/bin`.

**4. Make sure `~/.local/bin` is in your PATH**

Check with:
```bash
echo $PATH | grep -q ".local/bin" && echo "ok" || echo "add to PATH"
```

If it is not there, add to your `~/.bashrc` or `~/.zshrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload: `source ~/.bashrc`

**5. Run**

```bash
vd_stats
```

---

### Windows (via WSL)

The scroll feature relies on `tty` and `termios`, which are Unix-only modules. Native Windows terminals are not supported. The recommended approach is **WSL (Windows Subsystem for Linux)**.

**1. Enable WSL** (if not already enabled)

Open PowerShell as Administrator and run:
```powershell
wsl --install
```

Restart your machine when prompted.

**2. Open a WSL terminal and follow the Linux instructions above**

All steps are identical inside WSL. Your SSH keys from Windows are accessible at `/mnt/c/Users/YourName/.ssh/`.

You can point `SSH_KEY_PATH` to:
```python
SSH_KEY_PATH = "/mnt/c/Users/YourName/.ssh/id_rsa"
```

---

## Configuration reference

| Variable | Description | Default |
|---|---|---|
| `SSH_HOST` | IP or hostname of your VPS | — |
| `SSH_USER` | SSH username | `root` |
| `SSH_KEY_PATH` | Path to your private SSH key | `~/.ssh/id_rsa` |
| `SSH_PORT` | SSH port | `22` |
| `REFRESH_INTERVAL` | Seconds between data refreshes | `2` |

---

## How it works

1. Connects to the VPS via SSH using [paramiko](https://www.paramiko.org/)
2. Runs `docker ps -a` and `docker inspect` to collect container names, statuses, and their project folders (via the `com.docker.compose.project.working_dir` label set by Docker Compose)
3. Runs `docker stats --no-stream` to get live CPU and memory per running container
4. Runs `du -sh` on each project folder for disk usage
5. Groups everything by folder and renders the dashboard using [rich](https://github.com/Textualize/rich)

Containers not managed by Docker Compose (no project label) appear under a group called **Outros**.

---

## Updating

```bash
cd vd-stats
git pull
bash install.sh
```

---

## License

MIT
