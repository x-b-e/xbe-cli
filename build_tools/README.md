# Cartographer Build Tools

## Manual repo map (required)
- `build_tools/repo_map.json` is **hand-written on purpose**.
- Do not regenerate it with a script. This repo’s structure is stable, and the goal
  is to give agents a human-readable guide that explains how the codebase is laid out.
- If the server/client structure changes, edit the file directly:
  1) Update `search_paths` to the most relevant directories.
  2) Rewrite `structure_hint` to teach the structure (what to read first and how to trace).
  3) Keep hints short, specific, and oriented toward an agent that is new to the repo.

## Pipeline order
1) Update `build_tools/repo_map.json` by hand (see above).
2) Run `build_tools/dispatcher.py` to create batches.
3) Run `build_tools/worker.py` to generate artifacts (uses an agent).
4) Run `build_tools/compile.py` to build `cartographer_out/db/knowledge.sqlite`.

## Bootstrap (Python deps)
If your system Python is externally managed (PEP 668), run:
`bash build_tools/bootstrap.sh`

## Worker agents
- Default agent is Codex (`--agent codex`); Claude is optional (`--agent claude`).
- Codex runs with the equivalent of "skip permissions":
  `codex exec --dangerously-bypass-approvals-and-sandbox` (non-interactive).
- Claude uses `--dangerously-skip-permissions` with `--output-format json`.
- Agents write artifact files directly; their response payload is only used as a status signal.
- Provide a model with `--agent-model` if you want to override the default.
- Default is server + CLI (`--repos server,cli`). Use comma-separated names to include others.

### Swarm
Use `build_tools/swarm.py` to run multiple workers in parallel. Each worker atomically
claims a batch from `queue/pending` and exits when none remain. Swarm runs until the
pending queue is empty, or until `--max-batches` is reached.
Swarm logs batch progress, rate, and ETA to stdout.

## Failure behavior
- If an agent run fails, the worker deletes any partial artifacts for that batch
  and moves the batch back to `queue/pending` so it can be retried.

## Crash recovery
- `build_tools/swarm.py` runs crash recovery before starting workers.

## Validation
- Artifact JSON is validated with Pydantic in both the worker and compiler.
- The schema of record is `build_tools/artifacts_schema.py`.

## Source filtering
- Server sources are limited to `.rb` and `.sql` files.
- CLI sources are limited to `.go` files.
- Client sources are limited to `.js` files.
