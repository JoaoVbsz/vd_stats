#!/usr/bin/env bash
set -e

VENV_DIR="$HOME/.venvs/vd_stats"

echo "→ Criando virtualenv em $VENV_DIR..."
python3 -m venv "$VENV_DIR"

echo "→ Instalando dependências..."
"$VENV_DIR/bin/pip" install -e "$(dirname "$0")" -q

WRAPPER="$HOME/.local/bin/vd_stats"
mkdir -p "$HOME/.local/bin"

cat > "$WRAPPER" <<EOF
#!/usr/bin/env bash
exec "$VENV_DIR/bin/vd_stats" "\$@"
EOF

chmod +x "$WRAPPER"

echo "✓ Instalado! Certifique-se de que ~/.local/bin está no seu PATH."
echo "  Execute: vd_stats"
