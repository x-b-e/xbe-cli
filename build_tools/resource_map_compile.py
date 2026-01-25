#!/usr/bin/env python3
import argparse
import json
from pathlib import Path
from typing import Dict, List

from common import load_config, project_root


def ensure_list(value) -> List[str]:
    if not isinstance(value, list):
        return []
    return [item for item in value if isinstance(item, str)]


def merge_list(existing: List[str], incoming: List[str]) -> List[str]:
    result = list(existing)
    seen = set(result)
    for item in incoming:
        if item not in seen:
            result.append(item)
            seen.add(item)
    return result


def load_json(path: Path) -> Dict[str, object]:
    if not path.exists():
        return {}
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def write_json(path: Path, payload: Dict[str, object]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)
        handle.write("\n")


def validate_artifact(data: Dict[str, object]) -> None:
    required = ["resource", "server_types", "label_fields", "attributes", "relationships"]
    for key in required:
        if key not in data:
            raise ValueError(f"Missing {key} in artifact")
    if not isinstance(data["resource"], str) or not data["resource"].strip():
        raise ValueError("resource must be a non-empty string")
    if not isinstance(data["relationships"], dict):
        raise ValueError("relationships must be a map")


def main() -> None:
    parser = argparse.ArgumentParser(description="Merge resource-map artifacts into resource_map.json")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument(
        "--output",
        default="internal/cli/resource_map.json",
        help="Path to write merged resource_map.json",
    )
    parser.add_argument(
        "--base",
        default="internal/cli/resource_map.json",
        help="Base resource_map.json to merge into",
    )
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    artifacts_dir = root_out / "resource_map" / "artifacts"

    base_path = Path(args.base).resolve()
    output_path = Path(args.output).resolve()

    merged = load_json(base_path)
    resources = merged.get("resources", {})
    relationships = merged.get("relationships", {})
    if not isinstance(resources, dict):
        resources = {}
    if not isinstance(relationships, dict):
        relationships = {}

    artifacts = sorted(artifacts_dir.glob("*.json"))
    for artifact_path in artifacts:
        data = load_json(artifact_path)
        validate_artifact(data)
        resource = data["resource"]
        existing_resource = resources.get(resource, {})
        resources[resource] = {
            "server_types": merge_list(
                ensure_list(existing_resource.get("server_types")),
                ensure_list(data.get("server_types")),
            ),
            "label_fields": merge_list(
                ensure_list(existing_resource.get("label_fields")),
                ensure_list(data.get("label_fields")),
            ),
            "attributes": merge_list(
                ensure_list(existing_resource.get("attributes")),
                ensure_list(data.get("attributes")),
            ),
        }

        rel_map = relationships.get(resource, {})
        if not isinstance(rel_map, dict):
            rel_map = {}
        for rel_name, rel_data in data.get("relationships", {}).items():
            if not isinstance(rel_data, dict):
                continue
            existing_rel = rel_map.get(rel_name, {})
            rel_map[rel_name] = {
                "resources": merge_list(
                    ensure_list(existing_rel.get("resources")),
                    ensure_list(rel_data.get("resources")),
                )
            }
        relationships[resource] = rel_map

    merged = {
        "resources": resources,
        "relationships": relationships,
    }

    write_json(output_path, merged)
    print(f"Merged {len(artifacts)} artifacts into {output_path}")


if __name__ == "__main__":
    main()
