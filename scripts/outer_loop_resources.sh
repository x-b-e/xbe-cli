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
CODEX_CMD="${CODEX_CMD:-codex}"
CLAUDE_CMD="${CLAUDE_CMD:-claude}"
AGENT="${AGENT:-codex}"
AGENT_SILENT="${AGENT_SILENT:-1}"
AGENT_LOG_FILE="${AGENT_LOG_FILE:-}"
LOG_FILE="${LOG_FILE:-$ROOT_DIR/logs/outer_loop_resources.log}"
RUN_TESTS="${RUN_TESTS:-1}"
OUTER_LOOP_MAX_ITER="${OUTER_LOOP_MAX_ITER:-}"
OUTER_LOOP_DRY_RUN="${OUTER_LOOP_DRY_RUN:-0}"
OUTER_LOOP_QUEUE_FILE="${OUTER_LOOP_QUEUE_FILE:-}"
OUTER_LOOP_WORKER_ID="${OUTER_LOOP_WORKER_ID:-}"
OUTER_LOOP_AUTO_COMMIT="${OUTER_LOOP_AUTO_COMMIT:-1}"
OUTER_LOOP_REFRESH_RESOURCE_DECISIONS="${OUTER_LOOP_REFRESH_RESOURCE_DECISIONS:-1}"
OUTER_LOOP_MODE="${OUTER_LOOP_MODE:-standalone}"

if [[ ! -f "$RESOURCE_FILE" ]]; then
  echo "RESOURCE_DECISIONS.md not found at $RESOURCE_FILE" >&2
  exit 1
fi
if [[ ! -f "$SERVER_ROOT/config/routes.rb" ]]; then
  echo "Server routes not found at $SERVER_ROOT/config/routes.rb (set XBE_SERVER_DIR or SERVER_ROOT)" >&2
  exit 1
fi

if [[ "$AGENT" == "codex" ]]; then
  if ! command -v "$CODEX_CMD" >/dev/null 2>&1; then
    echo "codex CLI not found (set CODEX_CMD if needed)" >&2
    exit 1
  fi
elif [[ "$AGENT" == "claude" ]]; then
  if ! command -v "$CLAUDE_CMD" >/dev/null 2>&1; then
    echo "claude CLI not found (set CLAUDE_CMD if needed)" >&2
    exit 1
  fi
else
  echo "Unknown AGENT=$AGENT (expected codex or claude)" >&2
  exit 1
fi

mkdir -p "$(dirname "$LOG_FILE")"

log() {
  local prefix=""
  if [[ -n "$OUTER_LOOP_WORKER_ID" ]]; then
    prefix="[worker ${OUTER_LOOP_WORKER_ID}] "
  fi
  printf "%s %s%s\n" "$(date '+%F %T')" "$prefix" "$*" | tee -a "$LOG_FILE"
}

refresh_resource_decisions() {
  if [[ "$OUTER_LOOP_REFRESH_RESOURCE_DECISIONS" == "0" ]]; then
    return 0
  fi
  python3 - "$RESOURCE_FILE" "$ROOT_DIR" "$SERVER_ROOT" <<'PY'
import os
import re
import sys

path = sys.argv[1]
root = sys.argv[2]
server_root = sys.argv[3]

text = open(path, "r", encoding="utf-8").read()
lines = text.splitlines()

def parse_bullets(section_title):
    start = None
    end = None
    for i, line in enumerate(lines):
        if line.startswith(f"## {section_title}"):
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

pending = parse_bullets("Pending Decisions")
skipped = parse_bullets("Skipped (intentional)")
not_reviewed = parse_bullets("Not Yet Reviewed")

# Build command resources from Cobra registrations
cli_dir = os.path.join(root, "internal", "cli")
server_routes = os.path.join(server_root, "config", "routes.rb")

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
server_resources = {name.replace("_", "-") for name in re.findall(r"jsonapi_resources\s+:([a-z0-9_]+)", routes_text)}

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

remaining = sorted(server_resources - implemented_server - skipped - pending - not_reviewed)

analytics_re = re.compile(r"(summary|summaries|export|exports|report|reports|statistics|metrics|comparison|comparisons)$")
integration_re = re.compile(r"(integration|importer-configuration|exporter-configuration|open-ai|native-app|ui-tour|keep-truckin|digital-fleet|go-motive|deere|samsara|geotab|verizon|tenna|teletrac|gps-insight|gauge|t3-equipmentshare|one-step-gps|temeda|gauge|ozinga|lehman-roberts|haskell-lemon|superior-bowen|curran|textractions)")
content_re = re.compile(r"(notification|notifications|post|posts|comment-reactions|follows|text-messages|communications|questions|answers|answer-|prompter|prompters|newsletters|press-releases|release-notes)")
project_finance_re = re.compile(r"(^project-|^tender-|^rate-|^invoice-|^retainer|^commitment|^bid|bidders|^proffer|profit-improvement|customer-commitments|broker-commitments)")
core_ops_re = re.compile(r"(^job-|^shift-|^time-card|^time-sheet|^material-|^transport-|^equipment-|^tractor-|^trailer-|^maintenance-|^work-order|^crew-|^labor-|^driver-|^inventory-|^geofence-|^site-|^parking-sites$|^trips$|^hos-|^lineup-|^service-event|^service-site|^service-type-|^unit-of-measure|^resource-unavailabilities$)")
org_admin_re = re.compile(r"(^broker-|^customer-|^trucker-|^business-unit-|^membership|^memberships$|settings$|^developer-|^developers$|^api-tokens$|^application-settings$|^platform-statuses$|^organization-|^user-|^contractors$|^trading-partners$|^vendor|^vendors$|^search-|^model-filter|^features$|^tags$|^taggings$)")

buckets = {
    "Highest (Core operations & scheduling)": [],
    "High (Project & commercial workflows)": [],
    "Medium (Org/admin & reference)": [],
    "Low (Analytics, exports, summaries)": [],
    "Lowest (Notifications, content, integrations & vendor-specific)": [],
}

for r in remaining:
    if analytics_re.search(r):
        buckets["Low (Analytics, exports, summaries)"].append(r)
    elif integration_re.search(r) or content_re.search(r):
        buckets["Lowest (Notifications, content, integrations & vendor-specific)"].append(r)
    elif project_finance_re.search(r):
        buckets["High (Project & commercial workflows)"] .append(r)
    elif core_ops_re.search(r):
        buckets["Highest (Core operations & scheduling)"] .append(r)
    elif org_admin_re.search(r):
        buckets["Medium (Org/admin & reference)"] .append(r)
    else:
        buckets["Medium (Org/admin & reference)"] .append(r)

for k in buckets:
    buckets[k] = sorted(buckets[k])

def replace_section(text, heading, body):
    pattern = re.compile(rf"(?ms)^## {re.escape(heading)}\n.*?(?=^## |\Z)")
    return pattern.sub(f"## {heading}\n\n{body}\n\n", text)

status_body = "\n".join([
    f"- Server resources (routes): {len(server_resources)}",
    f"- CLI command resources: {len(command_resources)}",
    f"- Server resources covered by commands: {len(implemented_server)}",
    f"- Remaining (after skips + pending + not yet reviewed): {len(remaining)}",
])

implemented_body = "```\n" + "\n".join(sorted(command_resources)) + "\n```"

remaining_parts = []
for key in [
    "Highest (Core operations & scheduling)",
    "High (Project & commercial workflows)",
    "Medium (Org/admin & reference)",
    "Low (Analytics, exports, summaries)",
    "Lowest (Notifications, content, integrations & vendor-specific)",
]:
    items = buckets[key]
    if items:
        remaining_parts.append(f"### {key} ({len(items)})\n\n```\n" + "\n".join(items) + "\n```")
    else:
        remaining_parts.append(f"### {key} (0)\n\n(none)")

remaining_body = "\n\n".join(remaining_parts)

text = replace_section(text, "Status Summary", status_body)
text = replace_section(text, "Implemented (CLI commands exist for these resources)", implemented_body)
text = replace_section(text, "Remaining (by priority)", remaining_body)

with open(path, "w", encoding="utf-8") as f:
    f.write(text)
PY
}

select_next_resource() {
  python3 - "$RESOURCE_FILE" <<'PY'
import sys
import re

path = sys.argv[1]
lines = open(path, "r", encoding="utf-8").read().splitlines()

# Work Queue parsing
start = None
end = None
for i, line in enumerate(lines):
    if line.startswith("## Work Queue"):
        start = i
        continue
    if start is not None and i > start and line.startswith("## "):
        end = i
        break
if start is None:
    print("")
    sys.exit(1)
if end is None:
    end = len(lines)

header_idx = None
sep_idx = None
for i in range(start, end):
    if lines[i].startswith("| Resource / Command"):
        header_idx = i
        sep_idx = i + 1
        break

rows = []
if header_idx is not None:
    for i in range(sep_idx + 1, end):
        line = lines[i].strip()
        if not line.startswith("|"):
            break
        parts = [p.strip() for p in line.strip("|").split("|")]
        if len(parts) >= 3:
            rows.append((parts[0], parts[1].lower(), parts[2]))

in_progress = [r for r in rows if r[0] != "_TBD_" and r[1] == "in progress"]
planned = [r for r in rows if r[0] != "_TBD_" and r[1] == "planned"]
blocked = {r[0] for r in rows if r[0] != "_TBD_" and r[1] == "blocked"}

# Not Yet Reviewed
not_reviewed = set()
start_nr = None
end_nr = None
for i, line in enumerate(lines):
    if line.startswith("## Not Yet Reviewed"):
        start_nr = i
        continue
    if start_nr is not None and i > start_nr and line.startswith("## "):
        end_nr = i
        break
if start_nr is not None:
    if end_nr is None:
        end_nr = len(lines)
    for i in range(start_nr + 1, end_nr):
        line = lines[i].strip()
        if line.startswith("-"):
            val = line.strip("- ").strip("`")
            if val:
                not_reviewed.add(val)

# Remaining list (ordered by appearance)
remaining = []
start_rem = None
end_rem = None
for i, line in enumerate(lines):
    if line.startswith("## Remaining"):
        start_rem = i
        continue
    if start_rem is not None and i > start_rem and line.startswith("## "):
        end_rem = i
        break
if start_rem is not None:
    if end_rem is None:
        end_rem = len(lines)
    in_code = False
    for i in range(start_rem + 1, end_rem):
        line = lines[i].strip()
        if line.startswith("```"):
            in_code = not in_code
            continue
        if in_code and line:
            remaining.append(line)

# Choose
if in_progress:
    print(in_progress[0][0])
    sys.exit(0)
if planned:
    print(planned[0][0])
    sys.exit(0)

for r in remaining:
    if r in blocked:
        continue
    if r in not_reviewed:
        continue
    print(r)
    sys.exit(0)

print("")
sys.exit(1)
PY
}

update_work_queue() {
  local resource="$1"
  local status="$2"
  local notes="$3"
  local today
  today="$(date +%F)"
  python3 - "$RESOURCE_FILE" "$resource" "$status" "$notes" "$today" <<PY
import sys
import re

path, resource, status, notes, today = sys.argv[1:6]
lines = open(path, "r", encoding="utf-8").read().splitlines()

# locate Work Queue section
start = None
end = None
for i, line in enumerate(lines):
    if line.startswith("## Work Queue"):
        start = i
        continue
    if start is not None and i > start and line.startswith("## "):
        end = i
        break
if start is None:
    sys.exit("Work Queue section not found")
if end is None:
    end = len(lines)

# update Last updated line
for i in range(start, end):
    if lines[i].startswith("Last updated:"):
        lines[i] = f"Last updated: {today}"
        break

# find table header
header_idx = None
sep_idx = None
for i in range(start, end):
    if lines[i].startswith("| Resource / Command"):
        header_idx = i
        sep_idx = i + 1
        break
if header_idx is None:
    sys.exit("Work Queue table header not found")

# data rows
data_start = sep_idx + 1
row_end = data_start
while row_end < end and lines[row_end].strip().startswith("|"):
    row_end += 1

rows = []
for i in range(data_start, row_end):
    parts = [p.strip() for p in lines[i].strip().strip("|").split("|")]
    if len(parts) >= 3:
        rows.append(parts[:3])

# remove placeholder
rows = [r for r in rows if r[0] != "_TBD_"]

# update/insert
found = False
for r in rows:
    if r[0] == resource:
        r[1] = status
        r[2] = notes
        found = True
        break
if not found:
    rows.append([resource, status, notes])

# rebuild rows
new_rows = rows
if not new_rows:
    new_rows = [["_TBD_", "planned", "—"]]

prefix = lines[:data_start]
suffix = lines[row_end:]

out_lines = prefix + ["| " + " | ".join(r) + " |" for r in new_rows] + suffix

with open(path, "w", encoding="utf-8") as f:
    f.write("\n".join(out_lines) + "\n")
PY
}

remove_from_work_queue() {
  local resource="$1"
  local today
  today="$(date +%F)"
  python3 - "$RESOURCE_FILE" "$resource" "$today" <<PY
import sys

path, resource, today = sys.argv[1:4]
lines = open(path, "r", encoding="utf-8").read().splitlines()

start = None
end = None
for i, line in enumerate(lines):
    if line.startswith("## Work Queue"):
        start = i
        continue
    if start is not None and i > start and line.startswith("## "):
        end = i
        break
if start is None:
    sys.exit("Work Queue section not found")
if end is None:
    end = len(lines)

for i in range(start, end):
    if lines[i].startswith("Last updated:"):
        lines[i] = f"Last updated: {today}"
        break

header_idx = None
sep_idx = None
for i in range(start, end):
    if lines[i].startswith("| Resource / Command"):
        header_idx = i
        sep_idx = i + 1
        break
if header_idx is None:
    sys.exit("Work Queue table header not found")

data_start = sep_idx + 1
row_end = data_start
while row_end < end and lines[row_end].strip().startswith("|"):
    row_end += 1

rows = []
for i in range(data_start, row_end):
    parts = [p.strip() for p in lines[i].strip().strip("|").split("|")]
    if len(parts) >= 3:
        if parts[0] != resource and parts[0] != "_TBD_":
            rows.append(parts[:3])

if not rows:
    rows = [["_TBD_", "planned", "—"]]

prefix = lines[:data_start]
suffix = lines[row_end:]

out_lines = prefix + ["| " + " | ".join(r) + " |" for r in rows] + suffix

with open(path, "w", encoding="utf-8") as f:
    f.write("\n".join(out_lines) + "\n")
PY
}

is_implemented() {
  local resource="$1"
  python3 - "$ROOT_DIR" "$SERVER_ROOT" "$resource" <<PY
import os
import re
import sys

root = sys.argv[1]
server_root = sys.argv[2]
resource = sys.argv[3]

cli_dir = os.path.join(root, "internal", "cli")
server_routes = os.path.join(server_root, "config", "routes.rb")

# Build command resources from Cobra registrations
var_re = re.compile(r"var\s+(\w+)\s*=\s*&cobra\.Command\s*{", re.M)
use_re = re.compile(r"Use:\s*\"([^\"]+)\"")

var_use = {}
for filename in os.listdir(cli_dir):
    if not filename.endswith(".go"):
        continue
    text = open(os.path.join(cli_dir, filename), "r", encoding="utf-8").read()
    for m in var_re.finditer(text):
        name = m.group(1)
        brace_start = text.find("{", m.end() - 1)
        if brace_start == -1:
            continue
        depth = 0
        end = None
        for i in range(brace_start, len(text)):
            ch = text[i]
            if ch == "{":
                depth += 1
            elif ch == "}":
                depth -= 1
                if depth == 0:
                    end = i
                    break
        if end is None:
            continue
        block = text[brace_start:end + 1]
        um = use_re.search(block)
        if um:
            var_use[name] = um.group(1)

add_re = re.compile(r"(viewCmd|doCmd|summarizeCmd)\.AddCommand\(([^\)]*)\)", re.S)
command_resources = set()
for filename in os.listdir(cli_dir):
    if not filename.endswith(".go"):
        continue
    text = open(os.path.join(cli_dir, filename), "r", encoding="utf-8").read()
    for m in add_re.finditer(text):
        args = [a.strip() for a in m.group(2).split(",") if a.strip()]
        for arg in args:
            if re.match(r"^[A-Za-z_][A-Za-z0-9_]*$", arg):
                use = var_use.get(arg)
                if use:
                    command_resources.add(use)

# Server resources
routes_text = open(server_routes, "r", encoding="utf-8").read()
server_resources = {name.replace("_", "-") for name in re.findall(r"jsonapi_resources\s+:([a-z0-9_]+)", routes_text)}

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

implemented = set()
for cmd in command_resources:
    if cmd in alias_map:
        implemented.add(alias_map[cmd])
    elif cmd in server_resources:
        implemented.add(cmd)

if resource in command_resources or resource in implemented:
    sys.exit(0)

sys.exit(1)
PY
}

pop_queue_resource() {
  python3 - "$OUTER_LOOP_QUEUE_FILE" <<'PY'
import fcntl
import sys

path = sys.argv[1]
try:
    with open(path, "r+") as f:
        fcntl.flock(f, fcntl.LOCK_EX)
        lines = [line.strip() for line in f.readlines()]
        lines = [line for line in lines if line]
        if not lines:
            print("")
            sys.exit(0)
        next_item = lines[0]
        f.seek(0)
        f.truncate()
        f.write("\n".join(lines[1:]) + ("\n" if len(lines) > 1 else ""))
        fcntl.flock(f, fcntl.LOCK_UN)
        print(next_item)
except FileNotFoundError:
    print("")
    sys.exit(0)
PY
}

run_agent() {
  local prompt="$1"
  local output_target="/dev/null"
  if [[ -n "$AGENT_LOG_FILE" ]]; then
    output_target="$AGENT_LOG_FILE"
  fi
  if [[ "$AGENT" == "codex" ]]; then
    if [[ "$AGENT_SILENT" == "1" ]]; then
      "$CODEX_CMD" exec --dangerously-bypass-approvals-and-sandbox -C "$ROOT_DIR" - <<EOF >"$output_target" 2>&1
$prompt
EOF
    else
      "$CODEX_CMD" exec --dangerously-bypass-approvals-and-sandbox -C "$ROOT_DIR" - <<EOF
$prompt
EOF
    fi
  else
    if [[ "$AGENT_SILENT" == "1" ]]; then
      "$CLAUDE_CMD" --dangerously-skip-permissions --add-dir "$ROOT_DIR" -p "$prompt" >"$output_target" 2>&1
    else
      "$CLAUDE_CMD" --dangerously-skip-permissions --add-dir "$ROOT_DIR" -p "$prompt"
    fi
  fi
}

run_tests() {
  if [[ "$RUN_TESTS" == "0" ]]; then
    log "RUN_TESTS=0; skipping go test ./..."
    return 0
  fi
  log "Running: go test ./..."
  set +e
  (cd "$ROOT_DIR" && go test ./...) 2>&1 | tee -a "$LOG_FILE"
  local status=${PIPESTATUS[0]}
  set -e
  return $status
}

while true; do
  refresh_resource_decisions
  if [[ -n "$OUTER_LOOP_QUEUE_FILE" ]]; then
    resource="$(pop_queue_resource || true)"
  else
    resource="$(select_next_resource || true)"
  fi
  if [[ -z "${resource}" ]]; then
    log "No remaining resources to process."
    exit 0
  fi

  if is_implemented "$resource"; then
    log "Already implemented: ${resource}"
    if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
      remove_from_work_queue "$resource"
    fi
    continue
  fi

  if [[ -n "$OUTER_LOOP_WORKER_ID" ]]; then
    log "[worker ${OUTER_LOOP_WORKER_ID}] Next resource: ${resource}"
  else
    log "Next resource: ${resource}"
  fi

  if [[ "$OUTER_LOOP_DRY_RUN" == "1" ]]; then
    log "DRY RUN: would process ${resource}"
    exit 0
  fi

  if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
    update_work_queue "$resource" "in progress" "started by outer loop"
  fi

  set +e
  agent_prompt=$(cat <<EOF
Implement resource: ${resource}

Follow the Resource Implementation Spec in RESOURCE_DECISIONS.md. Focus on this single resource.
- Inspect server resource/policy/model in ${SERVER_ROOT}.
- Add view/do commands and tests per spec.
- Update help text and README.
- Keep changes scoped to this resource.
- Run gofmt on modified Go files.
- Run go test ./... and fix failures before finishing.

If you get blocked, explain why in your final response and stop.
EOF
)
  run_agent "$agent_prompt"
  codex_status=$?
  set -e

  refresh_resource_decisions

  if [[ $codex_status -ne 0 ]]; then
    log "Codex session failed for ${resource} (exit ${codex_status})."
    if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
      update_work_queue "$resource" "blocked" "codex session failed"
    fi
    continue
  fi

  if ! run_tests; then
    log "Tests failed after ${resource}."
    if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
      update_work_queue "$resource" "blocked" "go test ./... failed"
    fi
    continue
  fi

  if is_implemented "$resource"; then
    log "Completed: ${resource}"
    if [[ "$OUTER_LOOP_AUTO_COMMIT" == "1" ]]; then
      if [[ -n "$(git -C "$ROOT_DIR" status --porcelain)" ]]; then
        git -C "$ROOT_DIR" add -A
        git -C "$ROOT_DIR" commit -m "Implement ${resource}"
      fi
    fi
    if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
      remove_from_work_queue "$resource"
    fi
  else
    log "Blocked or incomplete: ${resource}"
    if [[ "$OUTER_LOOP_MODE" != "worker" ]]; then
      update_work_queue "$resource" "blocked" "implementation incomplete"
    fi
  fi

  if [[ -n "${OUTER_LOOP_MAX_ITER}" ]]; then
    OUTER_LOOP_MAX_ITER=$((OUTER_LOOP_MAX_ITER - 1))
    if [[ "$OUTER_LOOP_MAX_ITER" -le 0 ]]; then
      log "Reached OUTER_LOOP_MAX_ITER limit. Stopping."
      exit 0
    fi
  fi

done
