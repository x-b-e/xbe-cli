#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESOURCE_FILE="$ROOT_DIR/RESOURCE_DECISIONS.md"

GIT_COMMON_DIR="$(git -C "$ROOT_DIR" rev-parse --git-common-dir 2>/dev/null || true)"
if [[ -n "$GIT_COMMON_DIR" ]]; then
  if [[ "$GIT_COMMON_DIR" != /* ]]; then
    GIT_COMMON_DIR="$ROOT_DIR/$GIT_COMMON_DIR"
  fi
  MAIN_REPO_ROOT="$(cd "$GIT_COMMON_DIR/.." && pwd)"
  DEFAULT_SERVER_ROOT="$(cd "$MAIN_REPO_ROOT/.." && pwd)/server"
else
  DEFAULT_SERVER_ROOT="$(cd "$ROOT_DIR/.." && pwd)/server"
fi
SERVER_ROOT="${XBE_SERVER_DIR:-${SERVER_ROOT:-$DEFAULT_SERVER_ROOT}}"

RATE_WINDOW_HOURS="${RATE_WINDOW_HOURS:-4}"
MERGE_QUEUE_FILE="${OUTER_LOOP_MERGE_QUEUE_FILE:-$ROOT_DIR/.outer_loop_merge_queue.txt}"

python3 - "$ROOT_DIR" "$SERVER_ROOT" "$RESOURCE_FILE" "$RATE_WINDOW_HOURS" "$MERGE_QUEUE_FILE" "$DEFAULT_SERVER_ROOT" <<'PY'
import os
import re
import subprocess
import sys
import time

root = sys.argv[1]
server_root = sys.argv[2]
resource_file = sys.argv[3]
rate_window_hours = float(sys.argv[4])
merge_queue_file = sys.argv[5]
default_server_root = sys.argv[6]

def run_git(args):
    return subprocess.check_output(["git", "-C", root] + args, text=True).strip()

def parse_bullets(lines, title):
    start = end = None
    for i, line in enumerate(lines):
        if line.startswith(f"## {title}"):
            start = i
            continue
        if start is not None and i > start and line.startswith("## "):
            end = i
            break
    if start is None:
        return set()
    if end is None:
        end = len(lines)
    vals = set()
    for i in range(start + 1, end):
        line = lines[i].strip()
        if line.startswith("-"):
            val = line.strip("- ").strip("`")
            if val:
                vals.add(val)
    return vals

with open(resource_file, "r", encoding="utf-8") as f:
    lines = f.read().splitlines()

pending = parse_bullets(lines, "Pending Decisions")
skipped = parse_bullets(lines, "Skipped (intentional)")
not_reviewed = parse_bullets(lines, "Not Yet Reviewed")

cli_dir = os.path.join(root, "internal", "cli")

def load_server_resources(root_path):
    routes_path = os.path.join(root_path, "config", "routes.rb")
    if not os.path.isfile(routes_path):
        return None, routes_path
    routes_text = open(routes_path, "r", encoding="utf-8").read()
    resources = {name.replace("_", "-") for name in re.findall(r"jsonapi_resources\s+:([a-z0-9_]+)", routes_text)}
    return resources, routes_path

server_resources, server_routes = load_server_resources(server_root)
if not server_resources:
    if default_server_root and default_server_root != server_root:
        server_resources, server_routes = load_server_resources(default_server_root)
    if not server_resources:
        if server_routes is None or not os.path.isfile(server_routes):
            print(f"error: server routes not found at {server_routes}", file=sys.stderr)
        else:
            print(f"error: no jsonapi_resources found in {server_routes}", file=sys.stderr)
        sys.exit(1)

var_re = re.compile(r"var\s+(\w+)\s*=\s*&cobra\.Command\s*{", re.M)
use_re = re.compile(r"Use:\s*\"([^\"]+)\"")
var_use = {}
for filename in os.listdir(cli_dir):
    if not filename.endswith(".go"):
        continue
    text_file = open(os.path.join(cli_dir, filename), "r", encoding="utf-8").read()
    for m in var_re.finditer(text_file):
        name = m.group(1)
        brace_start = text_file.find("{", m.end() - 1)
        if brace_start == -1:
            continue
        depth = 0
        end = None
        for i in range(brace_start, len(text_file)):
            ch = text_file[i]
            if ch == "{":
                depth += 1
            elif ch == "}":
                depth -= 1
                if depth == 0:
                    end = i
                    break
        if end is None:
            continue
        block = text_file[brace_start:end + 1]
        um = use_re.search(block)
        if um:
            var_use[name] = um.group(1)

add_re = re.compile(r"(viewCmd|doCmd|summarizeCmd)\.AddCommand\(([^\)]*)\)", re.S)
command_resources = set()
for filename in os.listdir(cli_dir):
    if not filename.endswith(".go"):
        continue
    text_file = open(os.path.join(cli_dir, filename), "r", encoding="utf-8").read()
    for m in add_re.finditer(text_file):
        args = [a.strip() for a in m.group(2).split(",") if a.strip()]
        for arg in args:
            if re.match(r"^[A-Za-z_][A-Za-z0-9_]*$", arg):
                use = var_use.get(arg)
                if use:
                    command_resources.add(use)

routes_text = open(server_routes, "r", encoding="utf-8").read()

alias_map = {
    "lane-summary": "cycle-summaries",
    "shift-summary": "shift-summaries",
    "driver-day-summary": "driver-day-summaries",
    "job-production-plan-summary": "job-production-plan-summaries",
    "material-transaction-summary": "material-transaction-summaries",
    "device-location-event-summary": "device-location-event-summaries",
    "public-praise-summary": "public-praise-summaries",
    "transport-summary": "transport-summaries",
    "transport-order-efficiency-summary": "transport-order-efficiency-summaries",
    "ptp-summary": "project-transport-plan-summaries",
    "ptp-driver-summary": "project-transport-plan-driver-summaries",
    "ptp-trailer-summary": "project-transport-plan-trailer-summaries",
    "ptp-event-summary": "project-transport-plan-event-summaries",
    "ptp-event-time-summary": "project-transport-plan-event-time-summaries",
}

implemented_server = set()
for cmd in command_resources:
    if cmd in alias_map:
        implemented_server.add(alias_map[cmd])
    elif cmd in server_resources:
        implemented_server.add(cmd)

remaining = server_resources - implemented_server - skipped - pending - not_reviewed

# Unmerged implement commits on worker branches
worker_branches = []
branch_lines = run_git(["branch", "--list", "--format=%(refname:short)", "outer-loop-worker-*"]).splitlines()
for line in branch_lines:
    name = line.strip()
    if name:
        worker_branches.append(name)

unmerged_impl = set()
unmerged_ct = {}
for b in worker_branches:
    try:
        out = run_git(["log", "--format=%H %ct %s", f"main..{b}"])
    except subprocess.CalledProcessError:
        continue
    for line in out.splitlines():
        if " " not in line:
            continue
        parts = line.split(" ", 2)
        if len(parts) != 3:
            continue
        sha, ct, subj = parts
        if subj.startswith("Implement "):
            unmerged_impl.add(sha)
            try:
                unmerged_ct.setdefault(sha, int(ct))
            except ValueError:
                pass

# Rate window (unmerged commits only)
since_arg = f"{rate_window_hours} hours ago"
recent_unmerged_impl = set()
for b in worker_branches:
    try:
        out = run_git(["log", "--format=%H %ct %s", f"--since={since_arg}", f"main..{b}"])
    except subprocess.CalledProcessError:
        continue
    for line in out.splitlines():
        if " " not in line:
            continue
        parts = line.split(" ", 2)
        if len(parts) != 3:
            continue
        sha, _ct, subj = parts
        if subj.startswith("Implement "):
            recent_unmerged_impl.add(sha)

run_rate = 0.0
span_hours = 0.0
oldest_ct = None
newest_ct = None
if unmerged_ct:
    oldest_ct = min(unmerged_ct.values())
    newest_ct = max(unmerged_ct.values())
    span_hours = max(1e-9, (newest_ct - oldest_ct) / 3600)
    run_rate = len(unmerged_impl) / span_hours

merge_queue_len = 0
if merge_queue_file and os.path.exists(merge_queue_file):
    with open(merge_queue_file, "r", encoding="utf-8") as f:
        merge_queue_len = len([ln for ln in f.read().splitlines() if ln.strip()])

now = time.strftime("%Y-%m-%d %H:%M:%S")
print(f"timestamp: {now}")
implemented_total = len(implemented_server) + len(unmerged_impl)
print(f"implemented: {implemented_total}")
print(f"remaining: {len(remaining)}")
print(f"unmerged_worker_commits: {len(unmerged_impl)}")
print(f"remaining_after_unmerged: {max(0, len(remaining) - len(unmerged_impl))}")
print(f"rate_per_hour: {run_rate:.2f}")
if run_rate > 0:
    remaining_after_unmerged = max(0, len(remaining) - len(unmerged_impl))
    hours_left_after_unmerged = remaining_after_unmerged / run_rate
    print(f"eta_hours: {hours_left_after_unmerged:.1f}")
else:
    print("eta_hours: n/a")
PY
