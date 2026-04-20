def fetch_containers(ssh):
    ids_out, _ = ssh.run("docker ps -aq 2>/dev/null")
    if not ids_out:
        return []

    # Inspect all containers: full_id | name | status | working_dir
    inspect_cmd = (
        "docker inspect $(docker ps -aq) "
        "--format '{{.Id}}|{{.Name}}|{{.State.Status}}"
        "|{{index .Config.Labels \"com.docker.compose.project.working_dir\"}}'"
        " 2>/dev/null"
    )
    inspect_out, _ = ssh.run(inspect_cmd)

    # Stats only for running containers: short_id | cpu | mem_used
    stats_cmd = (
        "docker stats --no-stream "
        "--format '{{.ID}}|{{.CPUPerc}}|{{.MemUsage}}' 2>/dev/null"
    )
    stats_out, _ = ssh.run(stats_cmd)

    stats_map = {}
    for line in stats_out.splitlines():
        parts = line.split("|", 2)
        if len(parts) == 3:
            short_id, cpu, mem_usage = parts
            stats_map[short_id.strip()] = {
                "cpu": cpu.strip(),
                "mem": mem_usage.split("/")[0].strip(),
            }

    containers = []
    for line in inspect_out.splitlines():
        parts = line.split("|", 3)
        if len(parts) < 4:
            continue

        full_id, name, status, working_dir = parts
        short_id = full_id[:12]
        name = name.lstrip("/").strip()
        working_dir = working_dir.strip() or None

        stat = stats_map.get(short_id, {})
        containers.append({
            "id": short_id,
            "name": name,
            "status": status.strip(),
            "cpu": stat.get("cpu"),
            "mem": stat.get("mem"),
            "working_dir": working_dir,
        })

    return containers


def fetch_disk_usage(ssh, paths):
    if not paths:
        return {}
    paths_str = " ".join(f'"{p}"' for p in paths)
    out, _ = ssh.run(f"du -sh {paths_str} 2>/dev/null")
    result = {}
    for line in out.splitlines():
        parts = line.split("\t", 1)
        if len(parts) == 2:
            size, path = parts
            result[path.strip()] = size.strip()
    return result
