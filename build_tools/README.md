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

`compile.py` also loads `internal/cli/resource_map.json` into these tables:
- `resources` (includes `server_types`)
- `resource_fields`
- `resource_field_targets`

`compile.py` also loads `internal/cli/summary_map.json` into:
- `summary_resource_targets`
- `summary_sources`
- `resource_graph_edges` (view; unions relationships + summary links)

`compile.py` also builds command-to-resource links:
- `command_resource_links`

`compile.py` also builds flag-to-field semantics:
- `command_field_links` (use `--llm` to enable LLM fallback for unmapped flags)
  - `--llm-workers` controls parallelism (default 8)
  - `--llm-cache` sets the cache path (default `cartographer_out/db/llm_flag_cache.json`)

`compile.py` also extracts summary dimensions + metrics:
- `summary_dimensions`
- `summary_metrics`
- `command_summary_dimensions`
- `command_summary_metrics`

`compile.py` also builds filter-path links (multi-hop):
- `command_filter_paths`

`compile.py` also builds neighborhood mining views:
- `resource_filter_path_neighbors` (aggregated filter-path neighbors)
- `summary_resource_features` (summary dimensions + metrics)
- `summary_resource_neighbors` (summary ↔ primary resources)
- `resource_feature_links` (actionable features for similarity)
- `resource_feature_similarity` (bipartite projection by feature kind)
- `resource_metapath_similarity` (metapath similarity by feature kind)
- `resource_neighbor_components` (component-level evidence)
- `resource_neighbor_scores` (weighted neighborhood ranking)

It creates a `command_resources` view that links list/show commands to their
resource by parsing `commands.full_path`. Example query:

```
SELECT c.full_path, cr.resource, cr.verb, r.label_fields
FROM command_resources cr
JOIN commands c ON c.id = cr.command_id
JOIN resources r ON r.name = cr.resource
ORDER BY c.full_path;
```

## Resource map pipeline (for CLI --fields)
1) Run `build_tools/resource_map_dispatcher.py` to create resource batches.
2) Run `build_tools/resource_map_worker.py` (or `resource_map_swarm.py`) to generate artifacts.
3) Run `build_tools/resource_map_compile.py` to merge into `internal/cli/resource_map.json`.

## Deterministic show backfill
Use this to backfill show-command artifacts without an agent (uses CLI help + resource_map):
1) Ensure `internal/cli/resource_map.json` is up to date.
2) Run `python3 build_tools/deterministic_show_backfill.py`.

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
