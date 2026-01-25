#!/usr/bin/env python3
import argparse
import re
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from common import cli_bin_path, load_config, project_root, write_json

VERBS = {
    "create",
    "list",
    "view",
    "show",
    "get",
    "delete",
    "update",
    "add",
    "remove",
    "set",
    "unset",
    "enable",
    "disable",
    "assign",
    "unassign",
    "attach",
    "detach",
    "generate",
    "sync",
    "upload",
    "download",
    "start",
    "stop",
    "help",
}

IGNORE_COMMANDS = {"help", "completion"}
COMMAND_PREFIXES = {"do", "view", "summarize"}


def parse_subcommands(help_text: str, allow_resource_lines: bool) -> List[Tuple[str, str]]:
    subcommands: List[Tuple[str, str]] = []
    mode: Optional[str] = None
    indent_level: Optional[int] = None
    headers = {
        "Available Commands:",
        "AVAILABLE COMMANDS:",
        "Commands:",
        "COMMANDS:",
    }
    resources_headers = {
        "Resources:",
        "RESOURCES:",
    }
    stop_headers = {
        "FLAGS:",
        "GLOBAL FLAGS:",
        "CONFIGURATION:",
        "USAGE:",
        "EXAMPLES:",
        "LEARN MORE:",
    }
    command_pattern = re.compile(r"^\s+([a-zA-Z0-9][\w-]*)\s+(.+)$")
    resource_pattern = re.compile(r"^\s+([a-zA-Z0-9][\w-]*)$")
    for line in help_text.splitlines():
        stripped = line.strip()
        if stripped in headers:
            mode = "commands"
            indent_level = None
            continue
        if stripped in resources_headers:
            mode = "resources"
            indent_level = None
            continue
        if stripped in stop_headers:
            mode = None
            indent_level = None
        if mode is None:
            continue
        if stripped == "":
            continue
        if stripped.endswith(":") or stripped.startswith("["):
            continue
        if stripped.startswith("Use ") or stripped.startswith("These flags"):
            continue
        match = command_pattern.match(line)
        if match:
            indent = len(line) - len(line.lstrip(" "))
            if indent_level is None:
                indent_level = indent
            if indent != indent_level:
                continue
            name = match.group(1)
            desc = match.group(2).strip()
            if name in IGNORE_COMMANDS:
                continue
            subcommands.append((name, desc))
            continue
        if allow_resource_lines and mode in {"resources", "commands"}:
            match = resource_pattern.match(line)
            if match:
                indent = len(line) - len(line.lstrip(" "))
                if indent_level is None:
                    indent_level = indent
                if indent != indent_level:
                    continue
                name = match.group(1)
                if name in IGNORE_COMMANDS:
                    continue
                subcommands.append((name, ""))
    return subcommands


def run_help(cli_bin: Path, tokens: List[str]) -> str:
    cmd = [str(cli_bin)] + tokens + ["--help"]
    result = subprocess.run(cmd, capture_output=True, text=True, check=False)
    return (result.stdout or "") + (result.stderr or "")


def crawl_commands(cli_bin: Path) -> List[Dict[str, str]]:
    commands: List[Dict[str, str]] = []
    visited = set()

    def walk(tokens: List[str], description: str) -> None:
        key = " ".join(tokens)
        if key in visited:
            return
        visited.add(key)

        help_text = run_help(cli_bin, tokens)
        allow_resource_lines = len(tokens) == 1 and tokens[0] in {"view", "do"}
        subcommands = parse_subcommands(help_text, allow_resource_lines)
        if not subcommands and not tokens:
            print("[dispatcher] No subcommands for: (root)")
        if not subcommands:
            if tokens:
                commands.append({"full_path": key, "description": description})
            return
        for name, desc in subcommands:
            walk(tokens + [name], desc)

    walk([], "")
    return commands


def anchor_noun(command_path: str) -> str:
    tokens = [token for token in command_path.split() if not token.startswith("-")]
    cleaned: List[str] = []
    for token in tokens:
        token = token.strip()
        if not token or token.startswith("<") or token.startswith("["):
            continue
        cleaned.append(token)
    if cleaned and cleaned[0] in COMMAND_PREFIXES:
        cleaned = cleaned[1:]
    filtered = [token for token in cleaned if token not in VERBS]
    if filtered:
        return filtered[0]
    return cleaned[0] if cleaned else "misc"


def main() -> None:
    parser = argparse.ArgumentParser(description="Group CLI commands into Cartographer batches")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    cli_bin = cli_bin_path(config)
    if not cli_bin.exists():
        raise FileNotFoundError(f"CLI binary not found at {cli_bin}")

    commands = crawl_commands(cli_bin)
    batches: Dict[str, List[Dict[str, str]]] = {}
    for command in commands:
        noun = anchor_noun(command["full_path"])
        batches.setdefault(noun, []).append(command)

    root_out = project_root(config)
    queue_dir = root_out / "queue" / "pending"
    processing_dir = root_out / "queue" / "processing"
    queue_dir.mkdir(parents=True, exist_ok=True)
    processing_dir.mkdir(parents=True, exist_ok=True)
    for existing in queue_dir.glob("batch_*.json"):
        existing.unlink()
    for existing in processing_dir.glob("batch_*.json"):
        existing.unlink()
    for noun, items in batches.items():
        payload = {
            "batch_id": f"batch_{noun}",
            "noun": noun,
            "commands": items,
        }
        write_json(queue_dir / f"batch_{noun}.json", payload)

    print(f"Wrote {len(batches)} batches to {queue_dir}")


if __name__ == "__main__":
    main()
