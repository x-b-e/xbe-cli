#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESOURCE_FILE="$ROOT_DIR/RESOURCE_DECISIONS.md"
QUEUE_FILE_DEFAULT="$ROOT_DIR/.outer_loop_queue.txt"
WORKTREE_BASE="${OUTER_LOOP_WORKTREE_BASE:-$ROOT_DIR/.worktrees}"
WORKTREE_PREFIX="${OUTER_LOOP_WORKTREE_PREFIX:-outer-loop-worker}"
WORKERS="${OUTER_LOOP_WORKERS:-2}"
QUEUE_FILE="${OUTER_LOOP_QUEUE_FILE:-$QUEUE_FILE_DEFAULT}"
QUEUE_REBUILD="${OUTER_LOOP_QUEUE_REBUILD:-1}"
QUEUE_LIMIT="${OUTER_LOOP_QUEUE_LIMIT:-}"
REUSE_WORKTREES="${OUTER_LOOP_REUSE_WORKTREES:-0}"
LOG_FILE="${LOG_FILE:-$ROOT_DIR/logs/outer_loop_dispatcher.log}"
MERGE_QUEUE_FILE="${OUTER_LOOP_MERGE_QUEUE_FILE:-$ROOT_DIR/.outer_loop_merge_queue.txt}"

AGENT="${AGENT:-codex}"
CODEX_CMD="${CODEX_CMD:-codex}"
CLAUDE_CMD="${CLAUDE_CMD:-claude}"
RUN_TESTS="${RUN_TESTS:-1}"
OUTER_LOOP_MAX_ITER="${OUTER_LOOP_MAX_ITER:-}"
OUTER_LOOP_AUTO_COMMIT="${OUTER_LOOP_AUTO_COMMIT:-1}"
OUTER_LOOP_AUTO_MERGE="${OUTER_LOOP_AUTO_MERGE:-1}"
OUTER_LOOP_MERGE_STRATEGY="${OUTER_LOOP_MERGE_STRATEGY:-theirs}"
OUTER_LOOP_MAIN_BRANCH="${OUTER_LOOP_MAIN_BRANCH:-}"
XBE_SERVER_DIR="${XBE_SERVER_DIR:-${SERVER_ROOT:-}}"

mkdir -p "$(dirname "$LOG_FILE")"

log() {
  printf "%s %s\n" "$(date '+%F %T')" "$*" | tee -a "$LOG_FILE"
}

if [[ ! -f "$RESOURCE_FILE" ]]; then
  log "RESOURCE_DECISIONS.md not found at $RESOURCE_FILE"
  exit 1
fi

if [[ ! -x "$ROOT_DIR/scripts/outer_loop_resources.sh" ]]; then
  log "scripts/outer_loop_resources.sh not found or not executable"
  exit 1
fi

if [[ -z "$XBE_SERVER_DIR" ]]; then
  XBE_SERVER_DIR="$(cd "$ROOT_DIR/.." && pwd)/server"
fi

build_queue() {
  python3 - "$RESOURCE_FILE" "$QUEUE_FILE" "$QUEUE_LIMIT" <<'PY'
import sys

resource_file = sys.argv[1]
queue_file = sys.argv[2]
limit = sys.argv[3].strip()
limit_val = int(limit) if limit else None

lines = open(resource_file, "r", encoding="utf-8").read().splitlines()

def parse_work_queue(lines):
    start = end = None
    for i, line in enumerate(lines):
        if line.startswith("## Work Queue"):
            start = i
            continue
        if start is not None and i > start and line.startswith("## "):
            end = i
            break
    if start is None:
        return [], [], set()
    if end is None:
        end = len(lines)
    header_idx = sep_idx = None
    for i in range(start, end):
        if lines[i].startswith("| Resource / Command"):
            header_idx = i
            sep_idx = i + 1
            break
    if header_idx is None:
        return [], [], set()
    rows = []
    for i in range(sep_idx + 1, end):
        line = lines[i].strip()
        if not line.startswith("|"):
            break
        parts = [p.strip() for p in line.strip("|").split("|")]
        if len(parts) >= 3:
            rows.append((parts[0], parts[1].lower(), parts[2]))
    in_progress = [r[0] for r in rows if r[0] != "_TBD_" and r[1] == "in progress"]
    planned = [r[0] for r in rows if r[0] != "_TBD_" and r[1] == "planned"]
    blocked = {r[0] for r in rows if r[0] != "_TBD_" and r[1] == "blocked"}
    return in_progress, planned, blocked

def parse_remaining(lines):
    start = end = None
    for i, line in enumerate(lines):
        if line.startswith("## Remaining"):
            start = i
            continue
        if start is not None and i > start and line.startswith("## "):
            end = i
            break
    if start is None:
        return []
    if end is None:
        end = len(lines)
    remaining = []
    in_code = False
    for i in range(start + 1, end):
        line = lines[i].strip()
        if line.startswith("```"):
            in_code = not in_code
            continue
        if in_code and line:
            remaining.append(line)
    return remaining

in_progress, planned, blocked = parse_work_queue(lines)
remaining = parse_remaining(lines)

ordered = []
seen = set()
for item in in_progress + planned + remaining:
    if item in seen:
        continue
    if item in blocked:
        continue
    ordered.append(item)
    seen.add(item)

if limit_val is not None:
    ordered = ordered[:limit_val]

with open(queue_file, "w", encoding="utf-8") as f:
    f.write("\n".join(ordered) + ("\n" if ordered else ""))
PY
}

pop_merge_queue() {
  python3 - "$MERGE_QUEUE_FILE" <<'PY'
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

merge_commit() {
  local commit_sha="$1"
  if git -C "$ROOT_DIR" merge-base --is-ancestor "$commit_sha" "$OUTER_LOOP_MAIN_BRANCH"; then
    return 0
  fi
  if git -C "$ROOT_DIR" merge --no-ff -X "$OUTER_LOOP_MERGE_STRATEGY" -m "Merge $commit_sha" "$commit_sha"; then
    return 0
  fi

  if git -C "$ROOT_DIR" rev-parse -q --verify MERGE_HEAD >/dev/null; then
    if [[ "$OUTER_LOOP_MERGE_STRATEGY" == "ours" ]]; then
      git -C "$ROOT_DIR" checkout --ours -- .
    else
      git -C "$ROOT_DIR" checkout --theirs -- .
    fi
    git -C "$ROOT_DIR" add -A
    git -C "$ROOT_DIR" commit -m "Merge $commit_sha (auto-resolved)"
    return 0
  fi
  return 1
}

merge_monitor() {
  while true; do
    commit_sha="$(pop_merge_queue)"
    if [[ -n "$commit_sha" ]]; then
      if ! merge_commit "$commit_sha"; then
        log "Auto-merge failed for commit $commit_sha."
        return 1
      fi
      continue
    fi

    any_running=0
    for pid in "${pids[@]}"; do
      if kill -0 "$pid" 2>/dev/null; then
        any_running=1
        break
      fi
    done

    if [[ "$any_running" -eq 0 ]]; then
      commit_sha="$(pop_merge_queue)"
      if [[ -n "$commit_sha" ]]; then
        if ! merge_commit "$commit_sha"; then
          log "Auto-merge failed for commit $commit_sha."
          return 1
        fi
        continue
      fi
      break
    fi

    sleep 2
  done
}

if [[ "$QUEUE_REBUILD" == "1" || ! -f "$QUEUE_FILE" ]]; then
  build_queue
fi

mkdir -p "$WORKTREE_BASE"

if [[ -z "$OUTER_LOOP_MAIN_BRANCH" ]]; then
  OUTER_LOOP_MAIN_BRANCH="$(git -C "$ROOT_DIR" rev-parse --abbrev-ref HEAD)"
fi

if [[ "$OUTER_LOOP_AUTO_MERGE" == "1" ]]; then
  if [[ -n "$(git -C "$ROOT_DIR" status --porcelain)" ]]; then
    log "Main worktree has uncommitted changes; auto-merge is disabled until clean."
    exit 1
  fi
  git -C "$ROOT_DIR" checkout "$OUTER_LOOP_MAIN_BRANCH" >/dev/null
fi

declare -a pids=()

for i in $(seq 1 "$WORKERS"); do
  worktree_dir="$WORKTREE_BASE/$WORKTREE_PREFIX-$i"
  branch_name="$WORKTREE_PREFIX-$i"
  if [[ -d "$worktree_dir" ]]; then
    if [[ "$REUSE_WORKTREES" != "1" ]]; then
      log "Worktree exists at $worktree_dir (set OUTER_LOOP_REUSE_WORKTREES=1 to reuse)"
      exit 1
    fi
  else
    if git -C "$ROOT_DIR" show-ref --verify --quiet "refs/heads/$branch_name"; then
      git -C "$ROOT_DIR" worktree add "$worktree_dir" "$branch_name"
    else
      git -C "$ROOT_DIR" worktree add -b "$branch_name" "$worktree_dir"
    fi
  fi

  worker_script_dir="$worktree_dir/.outer_loop_scripts"
  mkdir -p "$worker_script_dir"
  cp "$ROOT_DIR/scripts/outer_loop_resources.sh" "$worker_script_dir/outer_loop_resources.sh"

  worker_log="$ROOT_DIR/logs/outer_loop_resources.worker-$i.log"
  mkdir -p "$(dirname "$worker_log")"

  (
    cd "$worktree_dir"
    AGENT="$AGENT" \
      CODEX_CMD="$CODEX_CMD" \
      CLAUDE_CMD="$CLAUDE_CMD" \
      RUN_TESTS="$RUN_TESTS" \
      OUTER_LOOP_MAX_ITER="$OUTER_LOOP_MAX_ITER" \
      OUTER_LOOP_QUEUE_FILE="$QUEUE_FILE" \
      OUTER_LOOP_MERGE_QUEUE_FILE="$MERGE_QUEUE_FILE" \
      OUTER_LOOP_WORKER_ID="$i" \
      OUTER_LOOP_MODE="worker" \
      OUTER_LOOP_REFRESH_RESOURCE_DECISIONS="0" \
      OUTER_LOOP_AUTO_COMMIT="$OUTER_LOOP_AUTO_COMMIT" \
      LOG_FILE="$worker_log" \
      XBE_SERVER_DIR="$XBE_SERVER_DIR" \
      bash ".outer_loop_scripts/outer_loop_resources.sh"
  ) &
  pids+=("$!")
done

merge_pid=""
if [[ "$OUTER_LOOP_AUTO_MERGE" == "1" ]]; then
  merge_monitor &
  merge_pid=$!
fi

failed=0
for pid in "${pids[@]}"; do
  if ! wait "$pid"; then
    failed=1
  fi
done

if [[ -n "$merge_pid" ]]; then
  if ! wait "$merge_pid"; then
    failed=1
  fi
fi

if [[ "$failed" -ne 0 ]]; then
  log "One or more workers failed."
  exit 1
fi

merge_branch() {
  local branch="$1"
  if git -C "$ROOT_DIR" merge --no-ff -X "$OUTER_LOOP_MERGE_STRATEGY" -m "Merge $branch" "$branch"; then
    return 0
  fi

  if git -C "$ROOT_DIR" rev-parse -q --verify MERGE_HEAD >/dev/null; then
    if [[ "$OUTER_LOOP_MERGE_STRATEGY" == "ours" ]]; then
      git -C "$ROOT_DIR" checkout --ours -- .
    else
      git -C "$ROOT_DIR" checkout --theirs -- .
    fi
    git -C "$ROOT_DIR" add -A
    git -C "$ROOT_DIR" commit -m "Merge $branch (auto-resolved)"
    return 0
  fi
  return 1
}

if [[ "$OUTER_LOOP_AUTO_MERGE" == "1" ]]; then
  if [[ -n "$(git -C "$ROOT_DIR" status --porcelain)" ]]; then
    log "Main worktree has uncommitted changes; auto-merge is disabled until clean."
    exit 1
  fi

  git -C "$ROOT_DIR" checkout "$OUTER_LOOP_MAIN_BRANCH"

  for i in $(seq 1 "$WORKERS"); do
    branch_name="$WORKTREE_PREFIX-$i"
    if ! git -C "$ROOT_DIR" show-ref --verify --quiet "refs/heads/$branch_name"; then
      continue
    fi

    ahead="$(git -C "$ROOT_DIR" rev-list --left-right --count "$OUTER_LOOP_MAIN_BRANCH...$branch_name" | awk '{print $2}')"
    if [[ "$ahead" == "0" ]]; then
      continue
    fi

    if ! merge_branch "$branch_name"; then
      log "Auto-merge failed for $branch_name."
      exit 1
    fi
  done
fi

log "All workers finished."
