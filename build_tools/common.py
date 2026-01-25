import hashlib
import json
from pathlib import Path
from typing import Any, Dict

import yaml


def load_config(config_path: str) -> Dict[str, Any]:
    config_file = Path(config_path).resolve()
    with config_file.open("r", encoding="utf-8") as handle:
        data = yaml.safe_load(handle)

    if not isinstance(data, dict):
        raise ValueError("config.yaml must be a mapping")

    base_dir = config_file.parent
    data["_config_dir"] = str(base_dir)
    return data


def resolve_path(base_dir: str, path_value: str) -> Path:
    path = Path(path_value)
    if not path.is_absolute():
        path = Path(base_dir) / path
    return path.resolve()


def project_root(config: Dict[str, Any]) -> Path:
    config_dir = config.get("_config_dir", ".")
    root_value = config.get("project_root", "./cartographer_out")
    return resolve_path(config_dir, root_value)


def cli_bin_path(config: Dict[str, Any]) -> Path:
    config_dir = config.get("_config_dir", ".")
    cli_value = config.get("cli_bin", "./bin/xbe-cli")
    return resolve_path(config_dir, cli_value)


def ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def write_json(path: Path, payload: Any) -> None:
    ensure_dir(path.parent)
    with path.open("w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)
        handle.write("\n")


def sha256_hex(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)
