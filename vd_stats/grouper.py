def group_by_project(containers):
    groups = {}

    for c in containers:
        key = c["working_dir"] or "Outros"
        if key not in groups:
            groups[key] = []
        groups[key].append({
            "name": c["name"],
            "status": c["status"],
            "cpu": c["cpu"],
            "mem": c["mem"],
        })

    # Named projects sorted alphabetically, "Outros" always last
    return dict(
        sorted(groups.items(), key=lambda kv: (kv[0] == "Outros", kv[0]))
    )
