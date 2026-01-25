#!/usr/bin/env python3
import argparse
import json
import os
import re
import subprocess
from pathlib import Path
from typing import Dict, List, Optional

from artifacts_schema import validate_artifact
from common import cli_bin_path, load_config, project_root, resolve_path, sha256_hex, write_json
from dispatcher import crawl_commands

GLOBAL_FLAG_NAMES = {
    "json",
    "limit",
    "offset",
    "sort",
    "base-url",
    "token",
    "no-auth",
    "fields",
    "help",
}


def find_repo_root(config: Dict[str, object], name: str) -> Path:
    for repo in config.get("repos", []):
        if repo.get("name") == name:
            return resolve_path(config.get("_config_dir", "."), repo.get("path", ""))
    raise ValueError(f"Repo named {name!r} not found in config.yaml")


def pluralize_token(token: str) -> str:
    if token.endswith(("s", "x", "z", "ch", "sh")):
        return token + "es"
    if token.endswith("y") and len(token) > 1 and token[-2] not in "aeiou":
        return token[:-1] + "ies"
    return token + "s"


def resource_hint_from_path(path: Path) -> str:
    stem = path.stem
    if stem.endswith("_resource"):
        stem = stem[: -len("_resource")]
    parts = stem.split("_")
    if parts:
        parts[-1] = pluralize_token(parts[-1])
    return "-".join(parts)


def gather_related_files(server_root: Path, resource_file: Path) -> List[str]:
    rel_path = resource_file.relative_to(server_root)
    stem = resource_file.stem
    if stem.endswith("_resource"):
        stem = stem[: -len("_resource")]
    model_path = server_root / "app" / "models" / f"{stem}.rb"
    policy_path = server_root / "app" / "policies" / f"{stem}_policy.rb"
    serializer_path = server_root / "app" / "serializers" / f"{stem}_serializer.rb"
    related = [rel_path]
    for candidate in (model_path, policy_path, serializer_path):
        if candidate.exists():
            related.append(candidate.relative_to(server_root))
    return [str(path) for path in related]


def is_abstract_resource(path: Path) -> bool:
    try:
        content = path.read_text(encoding="utf-8", errors="ignore")
    except OSError:
        return False
    return any(line.strip().startswith("abstract") for line in content.splitlines())


def build_resource_sources(server_root: Path) -> Dict[str, List[str]]:
    resource_dir = server_root / "app" / "resources" / "v1"
    sources: Dict[str, List[str]] = {}
    for dirpath, _dirnames, filenames in os.walk(resource_dir):
        for filename in filenames:
            if not filename.endswith("_resource.rb"):
                continue
            resource_file = Path(dirpath) / filename
            if is_abstract_resource(resource_file):
                continue
            resource_hint = resource_hint_from_path(resource_file)
            sources[resource_hint] = gather_related_files(server_root, resource_file)
    return sources


def load_help_text(cli_bin: Path, full_path: str) -> str:
    tokens = full_path.split()
    cmd = [str(cli_bin)] + tokens + ["--help"]
    result = subprocess.run(cmd, capture_output=True, text=True, check=False)
    return (result.stdout or "") + (result.stderr or "")


def parse_description(help_text: str) -> str:
    for line in help_text.splitlines():
        stripped = line.strip()
        if stripped:
            return stripped
    return ""


def map_flag_type(flag_type: Optional[str], description: str) -> str:
    if not flag_type:
        return "boolean"
    normalized = flag_type.lower()
    if normalized in {"string"}:
        return "string"
    if normalized in {"stringarray", "stringslice"}:
        return "array"
    if normalized in {"int", "int64"}:
        return "integer"
    if normalized in {"bool", "boolean"}:
        return "boolean"
    if "comma-separated" in description.lower():
        return "array"
    return "string"


def parse_flags(help_text: str) -> List[Dict[str, object]]:
    flags: List[Dict[str, object]] = []
    in_flags = False
    required_section = False
    for line in help_text.splitlines():
        stripped = line.strip()
        if stripped == "FLAGS:":
            in_flags = True
            required_section = False
            continue
        if stripped in {"GLOBAL FLAGS:", "CONFIGURATION:", "USAGE:", "EXAMPLES:", "LEARN MORE:"}:
            if in_flags:
                break
            continue
        if stripped.startswith("Required flags"):
            required_section = True
            continue
        if stripped.startswith("Optional flags"):
            required_section = False
            continue
        if not in_flags:
            continue
        match = re.match(r"^--([a-z0-9-]+)(?:\s+([^\s]+))?\s+(.*)$", stripped)
        if not match:
            continue
        name = match.group(1)
        if name in GLOBAL_FLAG_NAMES:
            continue
        flag_type = match.group(2)
        description = match.group(3).strip()
        required = required_section or "(required)" in description
        description = description.replace("(required)", "").strip()
        flags.append(
            {
                "name": f"--{name}",
                "aliases": None,
                "required": required,
                "type": map_flag_type(flag_type, description),
                "description": description,
                "default": None,
                "validation": None,
            }
        )
    flags.sort(key=lambda item: item["name"])
    return flags


def resolve_cli_source(resource: str) -> Optional[str]:
    filename = resource.replace("-", "_") + "_show.go"
    path = Path("internal/cli") / filename
    if path.exists():
        return str(path)
    return None


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Deterministically backfill show command artifacts from CLI help"
    )
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    cli_bin = cli_bin_path(config)
    if not cli_bin.exists():
        raise FileNotFoundError(f"CLI binary not found at {cli_bin}")

    server_root = find_repo_root(config, "server")
    resource_sources = build_resource_sources(server_root)

    commands = crawl_commands(cli_bin)
    show_commands = [
        cmd
        for cmd in commands
        if cmd.get("full_path", "").startswith("view ")
        and cmd.get("full_path", "").split()[-1] == "show"
    ]

    root_out = project_root(config)
    artifacts_dir = root_out / "artifacts"
    generated = 0

    for cmd in show_commands:
        full_path = cmd["full_path"]
        tokens = full_path.split()
        if len(tokens) < 3:
            continue
        resource = tokens[1]
        description = cmd.get("description") or ""
        help_text = load_help_text(cli_bin, full_path)
        if not description:
            description = parse_description(help_text)

        flags = parse_flags(help_text)
        sources: List[Dict[str, str]] = []
        cli_source = resolve_cli_source(resource)
        if cli_source:
            sources.append({"repo_name": "cli", "file_path": cli_source})
        for path in resource_sources.get(resource, []):
            sources.append({"repo_name": "server", "file_path": path})

        command_id = sha256_hex(full_path)
        output_path = artifacts_dir / resource / f"{command_id}.json"
        artifact = {
            "id": command_id,
            "full_path": full_path,
            "description": description or full_path,
            "permissions": None,
            "side_effects": None,
            "validation_notes": None,
            "flags": flags,
            "sources": sources,
        }
        validate_artifact(artifact)
        write_json(output_path, artifact)
        generated += 1

    print(f"Wrote {generated} deterministic show artifacts to {artifacts_dir}")


if __name__ == "__main__":
    main()
