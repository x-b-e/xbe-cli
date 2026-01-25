#!/usr/bin/env python3
import argparse
import json
import sqlite3
from pathlib import Path

from pydantic import ValidationError

from artifacts_schema import CommandArtifact, validate_artifact
from common import load_config, project_root


def init_db(db_path: Path) -> sqlite3.Connection:
    db_path.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(db_path)
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS commands (
            id TEXT PRIMARY KEY,
            full_path TEXT NOT NULL,
            description TEXT NOT NULL,
            permissions TEXT,
            side_effects TEXT,
            validation_notes TEXT
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS flags (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            command_id TEXT NOT NULL,
            name TEXT NOT NULL,
            aliases TEXT,
            required INTEGER NOT NULL,
            type TEXT NOT NULL,
            description TEXT NOT NULL,
            default_value TEXT,
            validation TEXT,
            FOREIGN KEY(command_id) REFERENCES commands(id)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS sources (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            command_id TEXT NOT NULL,
            repo_name TEXT NOT NULL,
            file_path TEXT NOT NULL,
            FOREIGN KEY(command_id) REFERENCES commands(id)
        )
        """
    )
    return conn


def upsert_artifact(conn: sqlite3.Connection, artifact: CommandArtifact) -> None:
    conn.execute(
        """
        INSERT OR REPLACE INTO commands (
            id, full_path, description, permissions, side_effects, validation_notes
        ) VALUES (?, ?, ?, ?, ?, ?)
        """,
        (
            artifact.id,
            artifact.full_path,
            artifact.description,
            artifact.permissions,
            artifact.side_effects,
            artifact.validation_notes,
        ),
    )
    conn.execute("DELETE FROM flags WHERE command_id = ?", (artifact.id,))
    conn.execute("DELETE FROM sources WHERE command_id = ?", (artifact.id,))

    for flag in artifact.flags:
        conn.execute(
            """
            INSERT INTO flags (
                command_id, name, aliases, required, type, description, default_value, validation
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                artifact.id,
                flag.name,
                json.dumps(flag.aliases) if flag.aliases else None,
                1 if flag.required else 0,
                flag.type,
                flag.description,
                flag.default,
                flag.validation,
            ),
        )
    for source in artifact.sources:
        conn.execute(
            """
            INSERT INTO sources (
                command_id, repo_name, file_path
            ) VALUES (?, ?, ?)
            """,
            (
                artifact.id,
                source.repo_name,
                source.file_path,
            ),
        )


def main() -> None:
    parser = argparse.ArgumentParser(description="Compile Cartographer artifacts into SQLite")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    artifacts_root = root_out / "artifacts"
    db_path = root_out / "db" / "knowledge.sqlite"

    conn = init_db(db_path)
    inserted = 0
    skipped = 0

    for artifact_path in artifacts_root.rglob("*.json"):
        try:
            data = json.loads(artifact_path.read_text(encoding="utf-8"))
            artifact = validate_artifact(data)
        except (json.JSONDecodeError, ValidationError) as exc:
            skipped += 1
            print(f"Skipping {artifact_path}: {exc}")
            continue
        upsert_artifact(conn, artifact)
        inserted += 1

    conn.commit()
    conn.close()
    print(f"Compiled {inserted} artifacts into {db_path} (skipped {skipped})")


if __name__ == "__main__":
    main()
