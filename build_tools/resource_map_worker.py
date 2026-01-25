#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import tempfile
from pathlib import Path
from typing import Dict, List, Optional

from common import load_config, project_root, resolve_path


def find_repo_root(config: Dict[str, object], name: str) -> Path:
    for repo in config.get("repos", []):
        if repo.get("name") == name:
            return resolve_path(config.get("_config_dir", "."), repo.get("path", ""))
    raise ValueError(f"Repo named {name!r} not found in config.yaml")


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


def run_codex_agent(prompt: str, model: Optional[str], reasoning: Optional[str]) -> None:
    fd, output_path = tempfile.mkstemp(prefix="codex-resource-map-", suffix=".json")
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
    output_file.unlink(missing_ok=True)
    if result.returncode != 0:
        raise RuntimeError(
            f"codex exec failed: {result.returncode}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )


def run_claude_agent(prompt: str, model: Optional[str]) -> None:
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


def validate_resource_map_artifact(path: Path) -> Dict[str, object]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError("Artifact must be a JSON object")
    required_keys = ["resource", "server_types", "label_fields", "attributes", "relationships"]
    for key in required_keys:
        if key not in data:
            raise ValueError(f"Missing required key: {key}")
    if not isinstance(data["resource"], str) or not data["resource"].strip():
        raise ValueError("resource must be a non-empty string")
    for list_key in ("server_types", "label_fields", "attributes"):
        if not isinstance(data[list_key], list):
            raise ValueError(f"{list_key} must be a list")
        if not all(isinstance(item, str) for item in data[list_key]):
            raise ValueError(f"{list_key} must only contain strings")
    if not isinstance(data["relationships"], dict):
        raise ValueError("relationships must be a map")
    for rel, rel_data in data["relationships"].items():
        if not isinstance(rel, str):
            raise ValueError("relationship keys must be strings")
        if not isinstance(rel_data, dict):
            raise ValueError(f"relationship {rel} must be an object")
        resources = rel_data.get("resources", [])
        if not isinstance(resources, list) or not all(isinstance(item, str) for item in resources):
            raise ValueError(f"relationship {rel} resources must be a list of strings")
    return data


def build_prompt(
    resource_hint: str,
    resource_file: str,
    sources: List[str],
    output_path: str,
    server_root: Path,
) -> str:
    source_lines = ["Sources (server repo):"]
    for path in sources:
        source_lines.append(f"  - {path}")
    source_block = "\n".join(source_lines)
    return "\n".join(
        [
            "You are a worker agent building a resource_map fragment for the XBE CLI.",
            "Your job is to read the provided server source files and write ONE JSON file",
            "that describes a single resource (attributes + has_one relationships only).",
            "This map powers the CLI --fields feature and a resource dependency graph.",
            "",
            "Rules:",
            "- Output MUST be strict JSON (double-quoted keys/strings). No comments or trailing commas.",
            "- Write ONLY the JSON object to the output file. Do not include prose.",
            "- Use kebab-case for attribute names (company-name) and relationship names.",
            "- Include ONLY attributes and has_one relationships (including create_only_has_one).",
            "- Do NOT include has_many relationships.",
            "- Use JSON:API server types (plural) in server_types.",
            "- Use CLI resource names (plural, kebab-case) in relationships.*.resources.",
            "- If a relationship is polymorphic, list all possible resources.",
            "- Choose 1-2 label_fields (name/title/label-style attributes) for this resource.",
            "- If uncertain, leave a list empty rather than guessing.",
            "",
            "Output schema (JSON object):",
            '{',
            '  "resource": "business-units",',
            '  "server_types": ["business-units"],',
            '  "label_fields": ["company-name"],',
            '  "attributes": ["company-name", "external-id"],',
            '  "relationships": {',
            '    "broker": { "resources": ["brokers"] },',
            '    "parent": { "resources": ["business-units"] }',
            "  }",
            "}",
            "",
            f"Resource hint (best guess): {resource_hint}",
            f"Server repo root: {server_root}",
            f"Primary resource file: {resource_file}",
            "",
            f"Write JSON to: {output_path}",
            "",
            source_block,
        ]
    )


def main() -> None:
    parser = argparse.ArgumentParser(description="Process resource-map batches")
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
    args = parser.parse_args()

    config = load_config(args.config)
    server_root = find_repo_root(config, "server")
    root_out = project_root(config)
    pending_dir = root_out / "resource_map" / "queue" / "pending"
    processing_dir = root_out / "resource_map" / "queue" / "processing"
    artifacts_dir = root_out / "resource_map" / "artifacts"

    processing_dir.mkdir(parents=True, exist_ok=True)

    def process_batch(processing_path: Path) -> None:
        batch = json.loads(processing_path.read_text(encoding="utf-8"))
        resource_hint = batch.get("resource_hint", "")
        resource_file = batch.get("resource_file", "")
        sources = batch.get("sources", [])
        batch_id = batch.get("batch_id", processing_path.stem)
        artifact_id = batch_id.replace("batch_resource_", "")
        output_path = artifacts_dir / f"{artifact_id}.json"
        print(f"Processing {processing_path.name} (resource_hint={resource_hint})")

        prompt = build_prompt(resource_hint, resource_file, sources, str(output_path), server_root)

        try:
            max_attempts = 2
            last_error: Optional[Exception] = None
            for attempt in range(1, max_attempts + 1):
                if output_path.exists():
                    output_path.unlink()
                attempt_prompt = prompt
                if attempt > 1:
                    attempt_prompt = (
                        prompt
                        + "\n\nIMPORTANT: Previous output was invalid JSON. "
                        + "Write strict JSON only (double quotes, no comments, no trailing commas)."
                    )
                if args.agent == "codex":
                    run_codex_agent(attempt_prompt, args.agent_model, args.reasoning)
                else:
                    run_claude_agent(attempt_prompt, args.agent_model)
                if not output_path.exists():
                    last_error = RuntimeError("Agent did not write the resource map artifact")
                    continue
                try:
                    _ = validate_resource_map_artifact(output_path)
                    last_error = None
                    break
                except Exception as exc:
                    last_error = exc
                    continue
            if last_error is not None:
                raise last_error
            if processing_path.parent == processing_dir and processing_path.exists():
                processing_path.unlink()
            print(f"Processed resource map artifact: {output_path.name}")
        except Exception:
            if output_path.exists():
                output_path.unlink()
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
