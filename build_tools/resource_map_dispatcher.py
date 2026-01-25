#!/usr/bin/env python3
import argparse
import os
from pathlib import Path
from typing import Dict, List, Optional

from common import load_config, project_root, resolve_path, sha256_hex, write_json


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


def main() -> None:
    parser = argparse.ArgumentParser(description="Create resource-map batches from server resources")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    server_root = find_repo_root(config, "server")
    resource_dir = server_root / "app" / "resources" / "v1"
    if not resource_dir.exists():
        raise FileNotFoundError(f"Resource directory not found: {resource_dir}")

    root_out = project_root(config)
    queue_dir = root_out / "resource_map" / "queue" / "pending"
    processing_dir = root_out / "resource_map" / "queue" / "processing"
    queue_dir.mkdir(parents=True, exist_ok=True)
    processing_dir.mkdir(parents=True, exist_ok=True)

    for existing in queue_dir.glob("batch_*.json"):
        existing.unlink()
    for existing in processing_dir.glob("batch_*.json"):
        existing.unlink()

    batches = 0
    for dirpath, _dirnames, filenames in os.walk(resource_dir):
        for filename in filenames:
            if not filename.endswith("_resource.rb"):
                continue
            resource_file = Path(dirpath) / filename
            try:
                content = resource_file.read_text(encoding="utf-8", errors="ignore")
            except OSError:
                content = ""
            if any(line.strip().startswith("abstract") for line in content.splitlines()):
                continue
            rel_path = resource_file.relative_to(server_root)
            resource_hint = resource_hint_from_path(resource_file)
            batch_id = sha256_hex(str(rel_path))
            sources = gather_related_files(server_root, resource_file)
            payload = {
                "batch_id": f"batch_resource_{batch_id}",
                "resource_hint": resource_hint,
                "resource_file": str(rel_path),
                "sources": sources,
            }
            write_json(queue_dir / f"batch_resource_{batch_id}.json", payload)
            batches += 1

    print(f"Wrote {batches} resource-map batches to {queue_dir}")


if __name__ == "__main__":
    main()
