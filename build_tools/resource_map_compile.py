#!/usr/bin/env python3
import argparse
import json
import re
from pathlib import Path
from typing import Dict, List, Optional, Set

from common import load_config, project_root, resolve_path


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


def find_repo_root(config: Dict[str, object], name: str) -> Optional[Path]:
    for repo in config.get("repos", []):
        if repo.get("name") == name:
            return resolve_path(config.get("_config_dir", "."), repo.get("path", ""))
    return None


def scan_versionable_models(server_root: Path) -> Set[str]:
    versionable: Set[str] = set()
    class_pattern = re.compile(r"^\s*class\s+([A-Za-z0-9_:]+)")
    extend_pattern = re.compile(r"\bextend\s+Versionable\b")
    for path in (server_root / "app/models").glob("**/*.rb"):
        text = path.read_text("utf-8", errors="ignore")
        if "extend Versionable" not in text:
            continue
        current_classes: List[str] = []
        for line in text.splitlines():
            match = class_pattern.match(line)
            if match:
                current_classes.append(match.group(1).split("::")[-1])
            if extend_pattern.search(line) and current_classes:
                versionable.add(current_classes[-1])
    return versionable


def scan_version_changes_optional_features(server_root: Path) -> Dict[str, List[str]]:
    resource_optional: Dict[str, List[str]] = {}
    class_pattern = re.compile(r"^\s*class\s+([A-Za-z0-9_:]+)\s*<")
    optional_pattern = re.compile(r"request_optional_features\.include\?\(\"([^\"]+)\"\)")
    for path in (server_root / "app/resources/v1").glob("**/*_resource.rb"):
        text = path.read_text("utf-8", errors="ignore")
        if "def version_changes" not in text:
            continue
        resource_class = None
        for line in text.splitlines():
            match = class_pattern.match(line)
            if match:
                resource_class = match.group(1).split("::")[-1]
                break
        if not resource_class:
            continue
        features = sorted(set(optional_pattern.findall(text)))
        if features:
            resource_optional[resource_class] = features
    return resource_optional


IRREGULAR_SINGULAR = {
    "people": "person",
    "children": "child",
    "men": "man",
    "women": "woman",
    "feet": "foot",
    "teeth": "tooth",
    "geese": "goose",
    "mice": "mouse",
    "oxen": "ox",
}


def singularize(word: str) -> str:
    if word in IRREGULAR_SINGULAR:
        return IRREGULAR_SINGULAR[word]
    if word.endswith("ies") and len(word) > 3:
        return word[:-3] + "y"
    if word.endswith(("ses", "xes", "zes", "ches", "shes")) and len(word) > 2:
        return word[:-2]
    if word.endswith("s") and not word.endswith("ss") and len(word) > 1:
        return word[:-1]
    return word


def camel_to_words(name: str) -> List[str]:
    s1 = re.sub(r"(.)([A-Z][a-z]+)", r"\1_\2", name)
    s2 = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", s1)
    return s2.lower().split("_")


def model_to_server_type(model: str) -> str:
    words = camel_to_words(model)
    if not words:
        return ""
    last = words[-1]
    if last.endswith("y") and len(last) > 1 and last[-2] not in "aeiou":
        last = last[:-1] + "ies"
    elif last.endswith(("s", "x", "z", "ch", "sh")):
        last = last + "es"
    elif last not in {"equipment", "news", "series", "species", "data"}:
        last = last + "s"
    words[-1] = last
    return "-".join(words)


def server_type_to_model_candidates(server_type: str) -> List[str]:
    parts = server_type.split("-")
    if not parts:
        return []
    last = parts[-1]
    candidates = [last, singularize(last)]
    models = []
    for candidate in dict.fromkeys(candidates):
        words = parts[:-1] + [candidate]
        snake = "_".join(words)
        model = "".join(word.title() for word in snake.split("_") if word)
        models.append(model)
    return models


def apply_version_changes_metadata(
    resources: Dict[str, object],
    versionable_models: Set[str],
    optional_features_by_resource_class: Dict[str, List[str]],
) -> None:
    optional_features_by_server_type: Dict[str, List[str]] = {}
    for resource_class, features in optional_features_by_resource_class.items():
        model_name = resource_class.replace("Resource", "")
        server_type = model_to_server_type(model_name)
        if server_type:
            optional_features_by_server_type[server_type] = features

    for resource_name, data in resources.items():
        server_types = data.get("server_types", [])
        supports = False
        for server_type in server_types:
            for candidate in server_type_to_model_candidates(server_type):
                if candidate in versionable_models:
                    supports = True
                    break
            if supports:
                break
        data["version_changes"] = bool(supports)

        optional_features: List[str] = []
        for server_type in server_types:
            optional_features.extend(optional_features_by_server_type.get(server_type, []))
        if optional_features:
            data["version_changes_optional_features"] = sorted(set(optional_features))


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
    server_root = find_repo_root(config, "server")
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

    if server_root and server_root.exists():
        versionable_models = scan_versionable_models(server_root)
        optional_features = scan_version_changes_optional_features(server_root)
        apply_version_changes_metadata(resources, versionable_models, optional_features)

    write_json(output_path, merged)
    print(f"Merged {len(artifacts)} artifacts into {output_path}")


if __name__ == "__main__":
    main()
