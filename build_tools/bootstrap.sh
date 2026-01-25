#!/usr/bin/env bash
set -euo pipefail

python3 -m venv .venv
source .venv/bin/activate
python3 -m pip install -r build_tools/requirements.txt
echo "Bootstrap complete. Activate with: source .venv/bin/activate"
