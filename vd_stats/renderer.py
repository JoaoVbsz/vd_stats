import datetime
import os
import signal
import sys
import termios
import threading
import tty
from typing import Dict, List, Optional, Tuple

from rich import box
from rich.align import Align
from rich.console import Console, Group
from rich.live import Live
from rich.panel import Panel
from rich.table import Table
from rich.text import Text


_MEM_UNITS = {
    "B": 1, "KB": 1e3, "KIB": 1024,
    "MB": 1e6, "MIB": 1024 ** 2,
    "GB": 1e9, "GIB": 1024 ** 3,
    "TB": 1e12, "TIB": 1024 ** 4,
}


def _format_gb(mem_str: Optional[str]) -> str:
    import re
    if not mem_str:
        return "—"
    m = re.match(r"([\d.]+)\s*([a-zA-Z]+)", mem_str.strip())
    if not m:
        return mem_str
    value, unit = float(m.group(1)), m.group(2).upper()
    bytes_val = value * _MEM_UNITS.get(unit, 1)
    return f"{bytes_val / 1e9:.3f} GB"


def _cpu_style(cpu_str: Optional[str]) -> str:
    try:
        val = float((cpu_str or "0").rstrip("%"))
        if val >= 20:
            return "bold red"
        if val >= 5:
            return "yellow"
        return "green"
    except ValueError:
        return "white"


class Renderer:
    def __init__(self):
        self._live: Optional[Live] = None
        self._scroll = 0
        self._max_scroll = 0
        self._lock = threading.Lock()
        self._stop_event = threading.Event()
        self._old_term = None
        self._key_thread: Optional[threading.Thread] = None

    def __enter__(self):
        self._live = Live(
            renderable="",
            screen=True,
            auto_refresh=False,
            console=Console(),
        )
        self._live.__enter__()
        self._key_thread = threading.Thread(target=self._key_loop, daemon=True)
        self._key_thread.start()
        return self

    def __exit__(self, *args):
        self._stop_event.set()
        self._restore_term()
        if self._live:
            self._live.__exit__(*args)

    def update(self, grouped_data: Dict, hostname: str, disk_usage: Dict = None):
        term_height = self._live.console.height
        with self._lock:
            scroll = self._scroll
        renderable, max_scroll = self._build(
            grouped_data, hostname, scroll, term_height, disk_usage or {}
        )
        with self._lock:
            self._max_scroll = max_scroll
            if scroll > max_scroll:
                self._scroll = max_scroll
        self._live.update(renderable)
        self._live.refresh()

    # ------------------------------------------------------------------

    def _build(
        self,
        grouped_data: Dict,
        hostname: str,
        scroll: int,
        term_height: int,
        disk_usage: Dict,
    ) -> Tuple:
        now = datetime.datetime.now().strftime("%H:%M:%S")

        all_rows: List[Tuple] = []
        for path, containers in grouped_data.items():
            all_rows.append(("project", path, None))
            for c in containers:
                all_rows.append(("container", c["name"], c))

        # 3 lines header panel + 3 lines footer panel + table header row = 7
        available = max(1, term_height - 7)
        max_scroll = max(0, len(all_rows) - available)
        visible = all_rows[scroll: scroll + available]

        table = Table(
            box=box.SIMPLE_HEAD,
            show_header=True,
            expand=True,
            header_style="bold dim",
            padding=(0, 1),
        )
        table.add_column("Container", no_wrap=True, min_width=30)
        table.add_column("CPU", justify="right", width=8)
        table.add_column("Memória", justify="right", width=12)
        table.add_column("Disco", justify="right", width=10)

        for kind, label, data in visible:
            if kind == "project":
                disk = disk_usage.get(label, "—")
                table.add_row(
                    Text(f"📁 {label}", style="bold blue"),
                    Text(""),
                    Text(""),
                    Text(disk, style="bold magenta"),
                )
            else:
                c = data
                if c["status"] == "running":
                    table.add_row(
                        Text(f"   {c['name']}", style="white"),
                        Text(c["cpu"] or "—", style=_cpu_style(c["cpu"])),
                        Text(_format_gb(c["mem"]), style="cyan"),
                        Text(""),
                    )
                else:
                    table.add_row(
                        Text(f"   {c['name']}", style="red"),
                        Text("✗ Down", style="bold red"),
                        Text(f"({c['status']})", style="red dim"),
                        Text(""),
                    )

        scroll_hint = ""
        if max_scroll > 0:
            end = min(scroll + available, len(all_rows))
            scroll_hint = f"   [dim]↑↓ para rolar  [{scroll + 1}–{end}/{len(all_rows)}][/]"

        renderable = Group(
            Panel(
                Align(
                    f"[bold cyan]VPS:[/] [cyan]{hostname}[/]  [dim]|[/]  [white]{now}[/]",
                    align="left",
                ),
                style="blue",
                height=3,
            ),
            table,
            Panel(
                f"[dim]Ctrl+C / Ctrl+D para sair[/]{scroll_hint}",
                style="dim",
                height=3,
            ),
        )

        return renderable, max_scroll

    # ------------------------------------------------------------------

    def _key_loop(self):
        fd = sys.stdin.fileno()
        if not os.isatty(fd):
            return
        try:
            self._old_term = termios.tcgetattr(fd)
            # setcbreak: desativa echo e modo canônico MAS preserva OPOST
            # (processamento de saída), que o rich precisa para renderizar corretamente.
            # setraw desabilita OPOST e quebraria o display.
            tty.setcbreak(fd)
            while not self._stop_event.is_set():
                ch = sys.stdin.read(1)
                if ch in ("", "\x04"):  # EOF ou Ctrl+D
                    os.kill(os.getpid(), signal.SIGINT)
                    break
                if ch == "\x1b":
                    seq = sys.stdin.read(2)
                    with self._lock:
                        if seq == "[A":  # Arrow up
                            self._scroll = max(0, self._scroll - 1)
                        elif seq == "[B":  # Arrow down
                            self._scroll = min(self._max_scroll, self._scroll + 1)
        except Exception:
            pass
        finally:
            self._restore_term()

    def _restore_term(self):
        if self._old_term:
            try:
                termios.tcsetattr(sys.stdin.fileno(), termios.TCSADRAIN, self._old_term)
            except Exception:
                pass
            self._old_term = None
