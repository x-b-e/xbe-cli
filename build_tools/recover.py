#!/usr/bin/env python3
import argparse
import os
from pathlib import Path

from common import load_config, load_json, project_root, sha256_hex


def recover_batch(batch_path: Path, pending_dir: Path, artifacts_dir: Path) -> None:
    batch = load_json(batch_path)
    noun = batch.get("noun", "misc")
    commands = batch.get("commands", [])
    for command in commands:
        full_path = command.get("full_path", "")
        if not full_path:
            continue
        command_id = sha256_hex(full_path)
        artifact_path = artifacts_dir / noun / f"{command_id}.json"
        if artifact_path.exists():
            artifact_path.unlink()

    destination = pending_dir / batch_path.name
    if destination.exists():
        index = 1
        while True:
            candidate = pending_dir / f"{batch_path.stem}_recovered_{index}.json"
            if not candidate.exists():
                destination = candidate
                break
            index += 1
    os.rename(batch_path, destination)


def main() -> None:
    parser = argparse.ArgumentParser(description="Recover crashed Cartographer batches")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    pending_dir = root_out / "queue" / "pending"
    processing_dir = root_out / "queue" / "processing"
    artifacts_dir = root_out / "artifacts"

    pending_dir.mkdir(parents=True, exist_ok=True)
    processing_dir.mkdir(parents=True, exist_ok=True)

    recovered = 0
    for batch_path in sorted(processing_dir.glob("batch_*.json")):
        recover_batch(batch_path, pending_dir, artifacts_dir)
        recovered += 1

    print(f"Recovered {recovered} batches from {processing_dir}")


if __name__ == "__main__":
    main()
