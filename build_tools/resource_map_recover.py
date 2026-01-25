#!/usr/bin/env python3
import argparse
import json
import os
from pathlib import Path
from typing import Dict, List

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


def is_abstract_resource(path: Path) -> bool:
    try:
        content = path.read_text(encoding="utf-8", errors="ignore")
    except OSError:
        return False
    return any(line.strip().startswith("abstract") for line in content.splitlines())


def validate_artifact(path: Path) -> bool:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return False
    if not isinstance(data, dict):
        return False
    required_keys = ["resource", "server_types", "label_fields", "attributes", "relationships"]
    for key in required_keys:
        if key not in data:
            return False
    if not isinstance(data.get("relationships"), dict):
        return False
    return True


def main() -> None:
    parser = argparse.ArgumentParser(description="Recover invalid or missing resource-map artifacts")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    server_root = find_repo_root(config, "server")
    resource_dir = server_root / "app" / "resources" / "v1"
    if not resource_dir.exists():
        raise FileNotFoundError(f"Resource directory not found: {resource_dir}")

    root_out = project_root(config)
    base_dir = root_out / "resource_map"
    artifacts_dir = base_dir / "artifacts"
    invalid_dir = base_dir / "artifacts_invalid"
    pending_dir = base_dir / "queue" / "pending"
    processing_dir = base_dir / "queue" / "processing"
    pending_dir.mkdir(parents=True, exist_ok=True)
    processing_dir.mkdir(parents=True, exist_ok=True)
    invalid_dir.mkdir(parents=True, exist_ok=True)

    recovered = 0
    for batch_path in sorted(processing_dir.glob("batch_*.json")):
        dest = pending_dir / batch_path.name
        if dest.exists():
            batch_path.unlink()
            continue
        os.rename(batch_path, dest)
        recovered += 1

    requeued = 0
    invalid = 0
    for dirpath, _dirnames, filenames in os.walk(resource_dir):
        for filename in filenames:
            if not filename.endswith("_resource.rb"):
                continue
            resource_file = Path(dirpath) / filename
            if is_abstract_resource(resource_file):
                continue
            rel_path = resource_file.relative_to(server_root)
            artifact_id = sha256_hex(str(rel_path))
            artifact_path = artifacts_dir / f"{artifact_id}.json"
            if artifact_path.exists() and validate_artifact(artifact_path):
                continue
            if artifact_path.exists():
                invalid += 1
                invalid_path = invalid_dir / artifact_path.name
                if invalid_path.exists():
                    invalid_path.unlink()
                os.rename(artifact_path, invalid_path)
            resource_hint = resource_hint_from_path(resource_file)
            sources = gather_related_files(server_root, resource_file)
            payload = {
                "batch_id": f"batch_resource_{artifact_id}",
                "resource_hint": resource_hint,
                "resource_file": str(rel_path),
                "sources": sources,
            }
            pending_path = pending_dir / f"batch_resource_{artifact_id}.json"
            write_json(pending_path, payload)
            requeued += 1

    print(
        f"Recovered {recovered} processing batch(es); "
        f"requeued {requeued} resource(s); invalid artifacts moved: {invalid}"
    )


if __name__ == "__main__":
    main()
