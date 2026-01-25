#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import tempfile
from pathlib import Path
from typing import Dict, List, Optional

from artifacts_schema import validate_artifact
from common import cli_bin_path, load_config, load_json, project_root, sha256_hex, write_json

SERVER_EXTENSIONS = {
    ".rb",
    ".sql",
}

CLI_EXTENSIONS = {
    ".go",
}

CLIENT_EXTENSIONS = {
    ".js",
}

# NOTE: repo_map.json is intentionally hand-curated. If structure changes, edit
# build_tools/repo_map.json directly and update its structure_hint text.


def select_batch(pending_dir: Path, batch_arg: Optional[str]) -> Optional[Path]:
    if batch_arg:
        return Path(batch_arg).resolve()
    candidates = sorted(pending_dir.glob("batch_*.json"))
    return candidates[0] if candidates else None


def claim_batch(pending_dir: Path, processing_dir: Path) -> Optional[Path]:
    candidates = sorted(pending_dir.glob("batch_*.json"))
    for candidate in candidates:
        processing_path = processing_dir / candidate.name
        try:
            os.rename(candidate, processing_path)
            return processing_path
        except FileNotFoundError:
            continue
        except OSError:
            continue
    return None


def file_matches(noun: str, path: Path, match_contents: bool) -> bool:
    variants = build_variants(noun)
    filename = path.name.lower()
    if any(variant in filename for variant in variants):
        return True
    if not match_contents:
        return False
    try:
        if path.stat().st_size > 1_000_000:
            return False
        content = path.read_text(encoding="utf-8", errors="ignore").lower()
        return any(variant in content for variant in variants)
    except OSError:
        return False


def build_variants(noun: str) -> List[str]:
    variants = set()
    base = noun.lower()
    variants.add(base)

    def add_singular(word: str) -> None:
        if word.endswith("ies") and len(word) > 3:
            variants.add(word[:-3] + "y")
        elif word.endswith("sses"):
            variants.add(word[:-2])
        elif word.endswith("s") and len(word) > 1:
            variants.add(word[:-1])

    def add_plural(word: str) -> None:
        if word.endswith(("s", "x", "z", "ch", "sh")):
            variants.add(word + "es")
        elif word.endswith("y") and len(word) > 1 and word[-2] not in "aeiou":
            variants.add(word[:-1] + "ies")
        else:
            variants.add(word + "s")

    add_singular(base)
    add_plural(base)

    if "-" in base:
        tokens = base.split("-")
        for token in tokens:
            add_singular(token)
            add_plural(token)
            variants.add(token)
        variants.add("".join(tokens))

    return sorted(variants)


def gather_source_files(
    repo_map: Dict[str, Dict], noun: str, allowed_repos: Optional[List[str]] = None
) -> List[Dict[str, str]]:
    matches: List[Dict[str, str]] = []
    for repo_name, info in repo_map.items():
        if allowed_repos and repo_name not in allowed_repos:
            continue
        extensions = CLIENT_EXTENSIONS
        match_contents = False
        if repo_name == "server":
            extensions = SERVER_EXTENSIONS
            match_contents = False
        elif repo_name == "cli":
            extensions = CLI_EXTENSIONS
            match_contents = False
        repo_root = Path(info.get("root_path", ""))
        for rel_path in info.get("search_paths", []):
            base = repo_root / rel_path
            if not base.exists():
                continue
            for dirpath, _dirnames, filenames in os.walk(base):
                for filename in filenames:
                    file_path = Path(dirpath) / filename
                    if file_path.suffix not in extensions:
                        continue
                    if not file_matches(noun, file_path, match_contents):
                        continue
                    matches.append(
                        {
                            "repo_name": repo_name,
                            "file_path": str(file_path.relative_to(repo_root)),
                            "abs_path": str(file_path),
                        }
                    )
    return matches


def load_help_text(cli_bin: Path, full_path: str) -> str:
    tokens = full_path.split()
    cmd = [str(cli_bin)] + tokens + ["--help"]
    result = subprocess.run(cmd, capture_output=True, text=True, check=False)
    return (result.stdout or "") + (result.stderr or "")

def schema_path() -> Path:
    return Path(__file__).resolve().parent / "artifacts_schema.py"


def build_agent_prompt(
    noun: str,
    commands: List[Dict[str, str]],
    sources: List[Dict[str, str]],
    help_texts: Dict[str, str],
    repo_map: Dict[str, Dict[str, str]],
    targets: List[Dict[str, str]],
) -> str:
    command_lines = []
    for cmd in commands:
        command_lines.append(f"- {cmd['full_path']}: {cmd.get('description','')}".strip())

    help_sections = []
    for cmd in commands:
        help_text = help_texts.get(cmd["full_path"], "").strip()
        if help_text:
            help_sections.append(f"## {cmd['full_path']} --help\n{help_text}")

    sources_by_repo: Dict[str, List[str]] = {}
    for src in sources:
        sources_by_repo.setdefault(src["repo_name"], []).append(src["file_path"])
    source_lines: List[str] = []
    for repo_name in sorted(sources_by_repo.keys()):
        source_lines.append(f"{repo_name}:")
        for path in sorted(set(sources_by_repo[repo_name])):
            source_lines.append(f"  - {path}")
    source_paths = "\n".join(source_lines)

    repo_hints = []
    for repo_name, info in repo_map.items():
        hint = info.get("structure_hint", "")
        if hint:
            repo_hints.append(f"{repo_name}: {hint}")

    schema_file = schema_path()
    schema_text = ""
    try:
        schema_text = schema_file.read_text(encoding="utf-8").strip()
    except OSError:
        schema_text = ""
    schema_section = [
        "Schema (Pydantic) file:",
        str(schema_file),
        "Open this file and follow it exactly.",
        "",
    ]
    if schema_text:
        schema_section = schema_section + [
            "Schema (inline):",
            schema_text,
            "",
        ]

    target_lines = []
    for target in targets:
        target_lines.append(
            f"- {target['full_path']} -> {target['output_path']} (id {target['id']})"
        )

    return "\n".join(
        [
            "CONTEXT",
            "You are a worker agent in a MapReduce pipeline that builds a Knowledge Database for the XBE CLI.",
            "Artifacts are later merged into SQLite for offline docs and agent context.",
            "",
            "TASK",
            "Read CLI help + relevant sources and write one CommandArtifact JSON file per command.",
            "",
            "EXTRACTION PRIORITIES",
            "1) Use help text for flags/usage/required fields.",
            "2) Use server policy/resource/model logic for permissions/side effects/validation notes when explicit.",
            "3) If something is unknown or not explicit, set it to null (do not guess).",
            "",
            "RULES",
            "- Validate output against the schema file (and inline schema) before writing.",
            "- Exclude global flags (--json, --limit, --offset, --sort, --base-url, --token, --no-auth).",
            "- Avoid procedural instructions unless help text explicitly says so.",
            "- Use the Sources list below as your starting set.",
            "- You may open additional repo files referenced by those sources if needed.",
            "- If you use additional files, add them to the artifact sources.",
            "- If both server and cli sources are provided, include at least one from each in every artifact.",
            "- Do not mention created_by in validation_notes; it is set server-side (not via API input).",
            "- Use the exact full_path from Output targets (including hyphens/spaces). Do not normalize or rename commands.",
            "- When a policy file is relevant, set permissions based on the policy and its permission methods; if it is not clearly stated, set permissions to null.",
            "- validation_notes should include the union of all flag-level validation notes plus any non-flag validation rules.",
            "- Include examples at the flag or filter level when they are known from help text or sources.",
            "",
            "OUTPUT CONTRACT",
            "- Write one JSON file per command to the specified output_path.",
            "- Also return JSON ONLY (no prose) with this exact shape:",
            "  {\"status\":\"ok\",\"written\":[\"<id>\",\"<id>\"]}",
            "  where written is the list of command IDs you wrote (from Output targets).",
            "",
            f"Noun: {noun}",
            "",
            "Commands:",
            *command_lines,
            "",
            "Output targets:",
            *target_lines,
            "",
            *schema_section,
            "",
            "Repo structure hints:",
            *repo_hints,
            "",
            "Sources (grouped by repo):",
            source_paths or "(none found)",
            "",
            "Help text:",
            *help_sections,
        ]
    )


def run_codex_agent(
    prompt: str, model: Optional[str], reasoning: Optional[str]
) -> Dict[str, object]:
    fd, output_path = tempfile.mkstemp(prefix="codex-artifacts-", suffix=".json")
    os.close(fd)
    output_file = Path(output_path)

    cmd = [
        "codex",
        "exec",
        "--output-last-message",
        str(output_file),
        "--dangerously-bypass-approvals-and-sandbox",
        "-",
    ]
    if model:
        cmd.extend(["--model", model])
    if reasoning:
        cmd.extend(["-c", f'model_reasoning_effort="{reasoning}"'])

    result = subprocess.run(cmd, input=prompt, capture_output=True, text=True, check=False)
    if result.returncode != 0:
        raise RuntimeError(
            f"codex exec failed: {result.returncode}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )

    payload = {"raw": output_file.read_text(encoding="utf-8")}
    output_file.unlink(missing_ok=True)
    return payload


def run_claude_agent(prompt: str, model: Optional[str]) -> Dict[str, object]:
    cmd = [
        "claude",
        "--print",
        "--dangerously-skip-permissions",
        "--output-format",
        "json",
    ]
    if model:
        cmd.extend(["--model", model])
    cmd.append(prompt)

    result = subprocess.run(cmd, capture_output=True, text=True, check=False)
    if result.returncode != 0:
        raise RuntimeError(
            f"claude failed: {result.returncode}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )

    return {"raw": result.stdout}


def validate_artifact_file(
    path: Path,
    expected_id: str,
    expected_full_path: str,
    repo_map: Dict[str, Dict],
    allowed_repos: Optional[set[str]],
    required_repos: set[str],
) -> None:
    data = json.loads(path.read_text(encoding="utf-8"))
    artifact = validate_artifact(data)
    if artifact.id != expected_id:
        raise ValueError(f"{path} id mismatch (expected {expected_id})")
    if artifact.full_path != expected_full_path:
        raise ValueError(f"{path} full_path mismatch (expected {expected_full_path})")
    for source in artifact.sources:
        repo_name = source.repo_name
        if repo_name not in repo_map:
            raise ValueError(f"{path} has unknown repo source: {repo_name}")
        if allowed_repos and repo_name not in allowed_repos:
            raise ValueError(f"{path} has source from repo not in scope: {repo_name}")
        repo_root = Path(repo_map[repo_name].get("root_path", "")).resolve()
        source_path = (repo_root / source.file_path).resolve()
        try:
            source_path.relative_to(repo_root)
        except ValueError as exc:
            raise ValueError(f"{path} has source outside repo root: {source.file_path}") from exc
        if not source_path.exists():
            raise ValueError(f"{path} has missing source file: {source.file_path}")
        allowed_exts = CLIENT_EXTENSIONS
        if repo_name == "server":
            allowed_exts = SERVER_EXTENSIONS
        elif repo_name == "cli":
            allowed_exts = CLI_EXTENSIONS
        if source_path.suffix not in allowed_exts:
            raise ValueError(f"{path} has source with disallowed extension: {source.file_path}")
        search_roots = [
            (repo_root / rel_path).resolve()
            for rel_path in repo_map[repo_name].get("search_paths", [])
        ]
        if search_roots:
            if not any(
                source_path == root or str(source_path).startswith(str(root) + os.sep)
                for root in search_roots
            ):
                raise ValueError(
                    f"{path} has source outside search_paths: {source.file_path}"
                )
    if required_repos:
        repos_present = {source.repo_name for source in artifact.sources}
        missing_repos = sorted(required_repos - repos_present)
        if missing_repos:
            raise ValueError(f"{path} missing sources for repos: {missing_repos}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Process Cartographer batches")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument("--batch", help="Path to a batch file (defaults to first pending batch)")
    parser.add_argument(
        "--loop",
        action="store_true",
        help="Keep claiming and processing batches until none remain",
    )
    parser.add_argument(
        "--agent",
        default="codex",
        choices=["codex", "claude"],
        help="Agent runner (codex or claude)",
    )
    parser.add_argument(
        "--agent-model",
        default="gpt-5.2-codex",
        help="Model name for agent runner",
    )
    parser.add_argument(
        "--reasoning",
        default="medium",
        choices=["low", "medium", "high", "minimal"],
        help="Reasoning effort for codex models (low/medium/high/minimal)",
    )
    parser.add_argument(
        "--repos",
        default="server,cli",
        help="Comma-separated repo names to search (default: server,cli)",
    )
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    pending_dir = root_out / "queue" / "pending"
    processing_dir = root_out / "queue" / "processing"
    artifacts_dir = root_out / "artifacts"

    processing_dir.mkdir(parents=True, exist_ok=True)

    def process_batch(processing_path: Path) -> None:
        batch = load_json(processing_path)
        noun = batch.get("noun", "misc")
        commands = batch.get("commands", [])
        print(f"Processing {processing_path.name} (noun={noun}, commands={len(commands)})")

        repo_map = load_json(Path(__file__).with_name("repo_map.json"))
        allowed_repos = None
        if args.repos:
            allowed_repos = [name.strip() for name in args.repos.split(",") if name.strip()]
            repo_map = {name: repo_map[name] for name in allowed_repos if name in repo_map}
        sources = gather_source_files(repo_map, noun, allowed_repos)
        required_repos = set()
        if allowed_repos and "server" in allowed_repos and "cli" in allowed_repos:
            required_repos = {"server", "cli"}
        allowed_repo_set = set(allowed_repos) if allowed_repos else None

        cli_bin = cli_bin_path(config)
        help_texts = {
            cmd["full_path"]: load_help_text(cli_bin, cmd["full_path"]) for cmd in commands
        }

        targets = []
        for command in commands:
            full_path = command["full_path"]
            command_id = sha256_hex(full_path)
            output_path = artifacts_dir / noun / f"{command_id}.json"
            targets.append(
                {
                    "id": command_id,
                    "full_path": full_path,
                    "output_path": str(output_path),
                }
            )

        prompt = build_agent_prompt(noun, commands, sources, help_texts, repo_map, targets)
        expected_count = len(targets)
        try:
            for target in targets:
                path = Path(target["output_path"])
                if path.exists():
                    path.unlink()
            if args.agent == "codex":
                payload = run_codex_agent(prompt, args.agent_model, args.reasoning)
            else:
                payload = run_claude_agent(prompt, args.agent_model)
            _ = payload
            missing = []
            for target in targets:
                path = Path(target["output_path"])
                if not path.exists():
                    missing.append(str(path))
                    continue
                validate_artifact_file(
                    path,
                    target["id"],
                    target["full_path"],
                    repo_map,
                    allowed_repo_set,
                    required_repos,
                )
            if missing:
                raise RuntimeError(f"Agent did not write {len(missing)} artifacts: {missing}")
            if processing_path.parent == processing_dir and processing_path.exists():
                processing_path.unlink()
            print(f"Processed {expected_count} artifacts for {noun}")
        except Exception:
            for target in targets:
                path = Path(target["output_path"])
                if path.exists():
                    path.unlink()
            if processing_path.parent == processing_dir:
                pending_path = pending_dir / processing_path.name
                try:
                    os.rename(processing_path, pending_path)
                except OSError:
                    pass
            raise

    if args.batch:
        batch_path = Path(args.batch).resolve()
        if not batch_path.exists():
            print("Batch not found.")
            return
        if batch_path.parent == pending_dir:
            processing_path = processing_dir / batch_path.name
            os.rename(batch_path, processing_path)
        else:
            processing_path = batch_path
        process_batch(processing_path)
        return

    if args.loop:
        while True:
            processing_path = claim_batch(pending_dir, processing_dir)
            if not processing_path:
                print("No pending batches found.")
                return
            process_batch(processing_path)
    else:
        processing_path = claim_batch(pending_dir, processing_dir)
        if not processing_path:
            print("No pending batches found.")
            return
        process_batch(processing_path)


if __name__ == "__main__":
    main()
