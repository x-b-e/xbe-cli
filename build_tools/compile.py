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
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS resources (
            name TEXT PRIMARY KEY,
            label_fields TEXT
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS resource_fields (
            resource TEXT NOT NULL,
            name TEXT NOT NULL,
            kind TEXT NOT NULL,
            description TEXT,
            is_label INTEGER NOT NULL DEFAULT 0,
            PRIMARY KEY (resource, name),
            FOREIGN KEY(resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS resource_field_targets (
            resource TEXT NOT NULL,
            field TEXT NOT NULL,
            target_resource TEXT NOT NULL,
            PRIMARY KEY (resource, field, target_resource),
            FOREIGN KEY(resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS command_resources AS
        SELECT
            id AS command_id,
            full_path,
            CASE
                WHEN full_path LIKE 'view % list'
                    THEN substr(full_path, 6, instr(substr(full_path, 6), ' ') - 1)
                WHEN full_path LIKE 'view % show'
                    THEN substr(full_path, 6, instr(substr(full_path, 6), ' ') - 1)
            END AS resource,
            CASE
                WHEN full_path LIKE 'view % list' THEN 'list'
                WHEN full_path LIKE 'view % show' THEN 'show'
            END AS verb
        FROM commands
        WHERE full_path LIKE 'view % list' OR full_path LIKE 'view % show';
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


def load_resource_map(resource_map_path: Path) -> dict:
    if not resource_map_path.exists():
        raise FileNotFoundError(f"resource_map.json not found: {resource_map_path}")
    return json.loads(resource_map_path.read_text(encoding="utf-8"))


def upsert_resource_map(conn: sqlite3.Connection, resource_map: dict) -> None:
    conn.execute("DELETE FROM resource_field_targets")
    conn.execute("DELETE FROM resource_fields")
    conn.execute("DELETE FROM resources")

    resources = resource_map.get("resources", {})
    relationships = resource_map.get("relationships", {})

    for resource_name, data in resources.items():
        label_fields = data.get("label_fields", [])
        conn.execute(
            "INSERT INTO resources (name, label_fields) VALUES (?, ?)",
            (resource_name, json.dumps(label_fields)),
        )

        for attr in data.get("attributes", []):
            conn.execute(
                """
                INSERT INTO resource_fields (resource, name, kind, description, is_label)
                VALUES (?, ?, ?, ?, ?)
                """,
                (
                    resource_name,
                    attr,
                    "attribute",
                    None,
                    1 if attr in label_fields else 0,
                ),
            )

    for resource_name, rels in relationships.items():
        for rel_name, rel_info in rels.items():
            conn.execute(
                """
                INSERT OR IGNORE INTO resource_fields (resource, name, kind, description, is_label)
                VALUES (?, ?, ?, ?, ?)
                """,
                (resource_name, rel_name, "relationship", None, 0),
            )
            for target in rel_info.get("resources", []):
                conn.execute(
                    """
                    INSERT OR IGNORE INTO resource_field_targets (resource, field, target_resource)
                    VALUES (?, ?, ?)
                    """,
                    (resource_name, rel_name, target),
                )


def main() -> None:
    parser = argparse.ArgumentParser(description="Compile Cartographer artifacts into SQLite")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument(
        "--resource-map",
        default=None,
        help="Path to resource_map.json (defaults to internal/cli/resource_map.json)",
    )
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    artifacts_root = root_out / "artifacts"
    db_path = root_out / "db" / "knowledge.sqlite"
    repo_root = Path(__file__).resolve().parents[1]
    resource_map_path = Path(args.resource_map) if args.resource_map else repo_root / "internal/cli/resource_map.json"

    conn = init_db(db_path)
    resource_map = load_resource_map(resource_map_path)
    upsert_resource_map(conn, resource_map)
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
