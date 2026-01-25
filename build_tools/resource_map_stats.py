#!/usr/bin/env python3
import argparse
import json
import time
from pathlib import Path
from typing import Dict, Tuple


def load_json(path: Path) -> Dict[str, object]:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def artifact_stats(artifacts_dir: Path) -> Tuple[int, int, float, float, float]:
    files = list(artifacts_dir.glob("*.json"))
    count = len(files)
    invalid = 0
    mtimes = []
    for path in files:
        mtimes.append(path.stat().st_mtime)
        try:
            load_json(path)
        except Exception:
            invalid += 1
    if not mtimes:
        return count, invalid, 0.0, 0.0, 0.0
    start = min(mtimes)
    end = max(mtimes)
    elapsed = max(end - start, 0.0)
    rate = count / elapsed if elapsed > 0 else 0.0
    return count, invalid, elapsed, start, rate


def infer_project_root(config_path: Path) -> Path:
    if not config_path.exists():
        return Path("cartographer_out").resolve()
    try:
        content = config_path.read_text(encoding="utf-8")
    except OSError:
        return Path("cartographer_out").resolve()
    for line in content.splitlines():
        stripped = line.strip()
        if stripped.startswith("project_root:"):
            _, value = stripped.split("project_root:", 1)
            value = value.strip().strip("\"'") or "cartographer_out"
            base = config_path.parent
            return (base / value).resolve()
    return Path("cartographer_out").resolve()


def main() -> None:
    parser = argparse.ArgumentParser(description="Stats for resource_map swarm progress")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument(
        "--root",
        help="Override project root (defaults to project_root from config.yaml or ./cartographer_out)",
    )
    args = parser.parse_args()

    if args.root:
        root_out = Path(args.root).expanduser().resolve()
    else:
        root_out = infer_project_root(Path(args.config).resolve())
    base_dir = root_out / "resource_map"
    artifacts_dir = base_dir / "artifacts"
    pending_dir = base_dir / "queue" / "pending"
    processing_dir = base_dir / "queue" / "processing"

    artifacts, invalid, elapsed, start, rate = artifact_stats(artifacts_dir)
    pending = len(list(pending_dir.glob("batch_*.json")))
    processing = len(list(processing_dir.glob("batch_*.json")))
    remaining = pending + processing
    eta = remaining / rate if rate > 0 else 0.0

    print(f"artifacts={artifacts}")
    print(f"invalid={invalid}")
    print(f"pending={pending}")
    print(f"processing={processing}")
    if start:
        print("first={}".format(time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(start))))
        print("last={}".format(time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(start + elapsed))))
    print(f"elapsed_sec={elapsed:.1f}")
    print("rate={:.3f} per sec ({:.1f}/min)".format(rate, rate * 60))
    print(f"remaining_batches={remaining}")
    eta_min = eta / 60 if eta > 0 else 0.0
    print(f"eta_min={eta_min:.1f}")
    print("----")


if __name__ == "__main__":
    main()
