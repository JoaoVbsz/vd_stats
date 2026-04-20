import signal
import sys
import time

from vd_stats.config import (
    REFRESH_INTERVAL,
    SSH_HOST,
    SSH_KEY_PATH,
    SSH_PORT,
    SSH_USER,
)
from vd_stats.docker_fetcher import fetch_containers, fetch_disk_usage
from vd_stats.grouper import group_by_project
from vd_stats.renderer import Renderer
from vd_stats.ssh_client import SSHClient

_running = True


def _handle_exit(sig, frame):
    global _running
    _running = False
    sys.exit(0)


def main():
    signal.signal(signal.SIGINT, _handle_exit)
    signal.signal(signal.SIGTERM, _handle_exit)

    try:
        with Renderer() as renderer:
            with SSHClient(SSH_HOST, SSH_USER, SSH_KEY_PATH, SSH_PORT) as ssh:
                hostname, _ = ssh.run("hostname")

                while _running:
                    try:
                        containers = fetch_containers(ssh)
                        grouped = group_by_project(containers)
                        paths = [p for p in grouped if p != "Outros"]
                        disk_usage = fetch_disk_usage(ssh, paths)
                        renderer.update(grouped, hostname, disk_usage)
                    except Exception:
                        pass  # mantém o display vivo, tenta de novo no próximo ciclo

                    time.sleep(REFRESH_INTERVAL)

    except (KeyboardInterrupt, SystemExit):
        pass
    except Exception as e:
        print(f"Erro de conexão SSH: {e}", file=sys.stderr)
        sys.exit(1)
