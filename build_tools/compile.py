#!/usr/bin/env python3
import argparse
import concurrent.futures
import hashlib
import json
import os
import re
import sqlite3
import shutil
import subprocess
import tempfile
from pathlib import Path
from typing import Optional

from pydantic import ValidationError

from artifacts_schema import CommandArtifact, validate_artifact
from common import load_config, project_root, resolve_path


def init_db(db_path: Path) -> sqlite3.Connection:
    db_path.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(db_path, timeout=60)
    conn.execute("PRAGMA busy_timeout = 60000")
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
        CREATE TABLE IF NOT EXISTS command_resource_links (
            command_id TEXT NOT NULL,
            resource TEXT NOT NULL,
            verb TEXT NOT NULL,
            command_kind TEXT NOT NULL,
            source TEXT NOT NULL,
            evidence TEXT,
            PRIMARY KEY (command_id, resource, verb, command_kind, source),
            FOREIGN KEY(command_id) REFERENCES commands(id)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS command_field_links (
            command_id TEXT NOT NULL,
            resource TEXT NOT NULL,
            field TEXT NOT NULL,
            field_kind TEXT NOT NULL,
            relation TEXT NOT NULL,
            flag_name TEXT NOT NULL,
            match_kind TEXT NOT NULL,
            modifier TEXT,
            PRIMARY KEY (command_id, resource, field, relation, flag_name),
            FOREIGN KEY(command_id) REFERENCES commands(id)
        )
        """
    )
    command_field_columns = {row[1] for row in conn.execute("PRAGMA table_info(command_field_links)")}
    if "modifier" not in command_field_columns:
        conn.execute("ALTER TABLE command_field_links ADD COLUMN modifier TEXT")
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS command_summary_dimensions (
            command_id TEXT NOT NULL,
            summary_resource TEXT NOT NULL,
            name TEXT NOT NULL,
            source_path TEXT NOT NULL,
            PRIMARY KEY (command_id, summary_resource, name),
            FOREIGN KEY(command_id) REFERENCES commands(id),
            FOREIGN KEY(summary_resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS command_summary_metrics (
            command_id TEXT NOT NULL,
            summary_resource TEXT NOT NULL,
            name TEXT NOT NULL,
            source_path TEXT NOT NULL,
            PRIMARY KEY (command_id, summary_resource, name),
            FOREIGN KEY(command_id) REFERENCES commands(id),
            FOREIGN KEY(summary_resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS command_filter_paths (
            command_id TEXT NOT NULL,
            resource TEXT NOT NULL,
            flag_name TEXT NOT NULL,
            path TEXT NOT NULL,
            target_resource TEXT NOT NULL,
            target_field TEXT,
            hop_count INTEGER NOT NULL,
            match_kind TEXT NOT NULL,
            modifier TEXT,
            source TEXT NOT NULL,
            PRIMARY KEY (command_id, flag_name, path),
            FOREIGN KEY(command_id) REFERENCES commands(id)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS summary_dimensions (
            summary_resource TEXT NOT NULL,
            name TEXT NOT NULL,
            kind TEXT NOT NULL,
            source_path TEXT NOT NULL,
            PRIMARY KEY (summary_resource, name, kind),
            FOREIGN KEY(summary_resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS summary_metrics (
            summary_resource TEXT NOT NULL,
            name TEXT NOT NULL,
            source_path TEXT NOT NULL,
            PRIMARY KEY (summary_resource, name),
            FOREIGN KEY(summary_resource) REFERENCES resources(name)
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
            label_fields TEXT,
            server_types TEXT
        )
        """
    )
    resource_columns = {row[1] for row in conn.execute("PRAGMA table_info(resources)")}
    if "server_types" not in resource_columns:
        conn.execute("ALTER TABLE resources ADD COLUMN server_types TEXT")
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
        CREATE TABLE IF NOT EXISTS summary_resource_targets (
            summary_resource TEXT NOT NULL,
            primary_resource TEXT NOT NULL,
            condition TEXT,
            PRIMARY KEY (summary_resource, primary_resource, condition),
            FOREIGN KEY(summary_resource) REFERENCES resources(name),
            FOREIGN KEY(primary_resource) REFERENCES resources(name)
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS summary_sources (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            summary_resource TEXT NOT NULL,
            repo_name TEXT NOT NULL,
            file_path TEXT NOT NULL,
            FOREIGN KEY(summary_resource) REFERENCES resources(name)
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
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_graph_edges AS
        SELECT
            resource AS source_resource,
            field AS relationship,
            target_resource,
            'relationship' AS edge_kind,
            NULL AS condition
        FROM resource_field_targets
        UNION ALL
        SELECT
            summary_resource AS source_resource,
            'summarizes' AS relationship,
            primary_resource AS target_resource,
            'summary' AS edge_kind,
            condition
        FROM summary_resource_targets;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_filter_path_neighbors AS
        SELECT
            resource AS source_resource,
            target_resource,
            path,
            target_field,
            hop_count,
            match_kind,
            modifier,
            COUNT(DISTINCT command_id) AS command_count,
            COUNT(DISTINCT flag_name) AS flag_count
        FROM command_filter_paths
        GROUP BY resource, target_resource, path, target_field, hop_count, match_kind, modifier;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS summary_resource_features AS
        SELECT
            summary_resource,
            'summary_dimension' AS feature_kind,
            name AS feature_name,
            kind AS feature_detail,
            source_path
        FROM summary_dimensions
        UNION ALL
        SELECT
            summary_resource,
            'summary_metric' AS feature_kind,
            name AS feature_name,
            NULL AS feature_detail,
            source_path
        FROM summary_metrics;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS summary_resource_neighbors AS
        SELECT
            summary_resource AS source_resource,
            primary_resource AS target_resource,
            condition,
            'summary' AS edge_kind
        FROM summary_resource_targets
        UNION ALL
        SELECT
            primary_resource AS source_resource,
            summary_resource AS target_resource,
            condition,
            'summary_of' AS edge_kind
        FROM summary_resource_targets;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_feature_links AS
        SELECT
            resource,
            'command_field' AS feature_kind,
            field AS feature_name,
            relation AS feature_detail,
            command_id AS source_ref
        FROM command_field_links
        UNION ALL
        SELECT
            summary_resource AS resource,
            'summary_dimension' AS feature_kind,
            name AS feature_name,
            kind AS feature_detail,
            source_path AS source_ref
        FROM summary_dimensions
        UNION ALL
        SELECT
            summary_resource AS resource,
            'summary_metric' AS feature_kind,
            name AS feature_name,
            NULL AS feature_detail,
            source_path AS source_ref
        FROM summary_metrics
        UNION ALL
        SELECT
            resource,
            'filter_target' AS feature_kind,
            target_resource AS feature_name,
            path AS feature_detail,
            command_id AS source_ref
        FROM command_filter_paths;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_feature_similarity AS
        WITH distinct_features AS (
            SELECT DISTINCT
                resource,
                feature_kind,
                feature_name,
                COALESCE(feature_detail, '') AS feature_detail
            FROM resource_feature_links
        )
        SELECT
            a.resource AS source_resource,
            b.resource AS target_resource,
            a.feature_kind,
            COUNT(*) AS shared_features
        FROM distinct_features a
        JOIN distinct_features b
            ON a.feature_kind = b.feature_kind
            AND a.feature_name = b.feature_name
            AND a.feature_detail = b.feature_detail
            AND a.resource < b.resource
        GROUP BY a.resource, b.resource, a.feature_kind;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_metapath_similarity AS
        SELECT
            source_resource,
            target_resource,
            feature_kind AS path_kind,
            shared_features
        FROM resource_feature_similarity;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_neighbor_components AS
        WITH symmetric_components AS (
            SELECT
                source_resource,
                target_resource,
                CASE feature_kind
                    WHEN 'command_field' THEN 'shared_command_field'
                    WHEN 'summary_dimension' THEN 'shared_summary_dimension'
                    WHEN 'summary_metric' THEN 'shared_summary_metric'
                    WHEN 'filter_target' THEN 'shared_filter_target'
                    ELSE feature_kind
                END AS component_kind,
                shared_features AS component_count,
                NULL AS detail
            FROM resource_feature_similarity
        ),
        direct_components AS (
            SELECT
                source_resource,
                target_resource,
                CASE edge_kind
                    WHEN 'summary' THEN 'summary'
                    ELSE 'relationship'
                END AS component_kind,
                1 AS component_count,
                CASE
                    WHEN edge_kind = 'summary' AND condition IS NOT NULL THEN condition
                    ELSE relationship
                END AS detail
            FROM resource_graph_edges
            UNION ALL
            SELECT
                source_resource,
                target_resource,
                'filter_path' AS component_kind,
                flag_count AS component_count,
                path AS detail
            FROM resource_filter_path_neighbors
        )
        SELECT source_resource, target_resource, component_kind, component_count, detail
        FROM direct_components
        UNION ALL
        SELECT source_resource, target_resource, component_kind, component_count, detail
        FROM symmetric_components
        UNION ALL
        SELECT target_resource AS source_resource,
               source_resource AS target_resource,
               component_kind,
               component_count,
               detail
        FROM symmetric_components;
        """
    )
    conn.execute(
        """
        CREATE VIEW IF NOT EXISTS resource_neighbor_scores AS
        WITH weights(component_kind, weight) AS (
            VALUES
                ('relationship', 3.0),
                ('summary', 2.5),
                ('filter_path', 1.5),
                ('shared_command_field', 1.0),
                ('shared_summary_dimension', 0.8),
                ('shared_summary_metric', 0.8),
                ('shared_filter_target', 0.5)
        ),
        components AS (
            SELECT
                source_resource,
                target_resource,
                component_kind,
                CASE WHEN component_count > 10 THEN 10 ELSE component_count END AS capped_count
            FROM resource_neighbor_components
        ),
        scored AS (
            SELECT
                c.source_resource,
                c.target_resource,
                c.component_kind,
                c.capped_count,
                w.weight,
                (c.capped_count * w.weight) AS component_score
            FROM components c
            JOIN weights w ON w.component_kind = c.component_kind
        )
        SELECT
            source_resource,
            target_resource,
            SUM(component_score) AS score,
            SUM(capped_count) AS evidence_count,
            SUM(CASE WHEN component_kind = 'relationship' THEN capped_count ELSE 0 END) AS relationship_count,
            SUM(CASE WHEN component_kind = 'summary' THEN capped_count ELSE 0 END) AS summary_count,
            SUM(CASE WHEN component_kind = 'filter_path' THEN capped_count ELSE 0 END) AS filter_path_count,
            SUM(CASE WHEN component_kind = 'shared_command_field' THEN capped_count ELSE 0 END)
                AS shared_command_field_count,
            SUM(CASE WHEN component_kind = 'shared_summary_dimension' THEN capped_count ELSE 0 END)
                AS shared_summary_dimension_count,
            SUM(CASE WHEN component_kind = 'shared_summary_metric' THEN capped_count ELSE 0 END)
                AS shared_summary_metric_count,
            SUM(CASE WHEN component_kind = 'shared_filter_target' THEN capped_count ELSE 0 END)
                AS shared_filter_target_count
        FROM scored
        GROUP BY source_resource, target_resource;
        """
    )
    return conn


def find_repo_root(config: dict, name: str) -> Path:
    for repo in config.get("repos", []):
        if repo.get("name") == name:
            return resolve_path(config.get("_config_dir", "."), repo.get("path", ""))
    raise ValueError(f"Repo named {name!r} not found in config.yaml")


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
        server_types = data.get("server_types", [])
        conn.execute(
            "INSERT INTO resources (name, label_fields, server_types) VALUES (?, ?, ?)",
            (
                resource_name,
                json.dumps(label_fields),
                json.dumps(server_types) if server_types else None,
            ),
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

        for common_attr in ("created-at", "updated-at"):
            conn.execute(
                """
                INSERT OR IGNORE INTO resource_fields (resource, name, kind, description, is_label)
                VALUES (?, ?, ?, ?, ?)
                """,
                (
                    resource_name,
                    common_attr,
                    "attribute",
                    None,
                    0,
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


def load_summary_map(summary_map_path: Path) -> dict:
    if not summary_map_path.exists():
        return {}
    return json.loads(summary_map_path.read_text(encoding="utf-8"))


def upsert_summary_map(conn: sqlite3.Connection, summary_map: dict) -> None:
    conn.execute("DELETE FROM summary_resource_targets")
    conn.execute("DELETE FROM summary_sources")

    summaries = summary_map.get("summaries", {})
    if not isinstance(summaries, dict):
        return

    for summary_resource, data in summaries.items():
        if not isinstance(data, dict):
            continue

        for primary in data.get("primary_resources", []) or []:
            conn.execute(
                """
                INSERT OR IGNORE INTO summary_resource_targets
                (summary_resource, primary_resource, condition)
                VALUES (?, ?, ?)
                """,
                (summary_resource, primary, None),
            )

        for condition in data.get("conditions", []) or []:
            if not isinstance(condition, dict):
                continue
            condition_filter = condition.get("filter")
            condition_json = json.dumps(condition_filter) if condition_filter else None
            for primary in condition.get("primary_resources", []) or []:
                conn.execute(
                    """
                    INSERT OR IGNORE INTO summary_resource_targets
                    (summary_resource, primary_resource, condition)
                    VALUES (?, ?, ?)
                    """,
                    (summary_resource, primary, condition_json),
                )

        for source in data.get("sources", []) or []:
            if not isinstance(source, dict):
                continue
            conn.execute(
                """
                INSERT INTO summary_sources (summary_resource, repo_name, file_path)
                VALUES (?, ?, ?)
                """,
                (summary_resource, source.get("repo_name"), source.get("file_path")),
            )


def normalize_flag_name(flag_name: str) -> str:
    name = flag_name.strip()
    while name.startswith("-"):
        name = name[1:]
    return name.lower().replace("_", "-")


def match_flag_to_field(
    flag_name: str, fields: dict[str, str]
) -> tuple[Optional[str], Optional[str], Optional[str]]:
    normalized = normalize_flag_name(flag_name)
    if normalized in fields:
        return normalized, "exact", None

    if normalized.startswith("not-"):
        base = normalized[4:]
        if base in fields:
            return base, "negation", "not"

    if normalized.startswith("is-"):
        base = normalized[3:]
        if base in fields:
            return base, "presence", "is"

    for suffix, modifier in (("-min", "min"), ("-max", "max"), ("-before", "before"), ("-after", "after")):
        if normalized.endswith(suffix):
            base = normalized[: -len(suffix)]
            if base in fields:
                return base, "range", modifier
            if base.endswith("-id"):
                base_strip = base[:-3]
                if base_strip in fields:
                    return base_strip, "range_strip_id", modifier
            if base.endswith("-ids"):
                base_strip = base[:-4]
                if base_strip in fields:
                    return base_strip, "range_strip_id", modifier

    if normalized.endswith("-id"):
        base = normalized[:-3]
        if base in fields:
            return base, "strip_id", None
    if normalized.endswith("-ids"):
        base = normalized[:-4]
        if base in fields:
            return base, "strip_id", None

    return None, None, None


def extract_json_object(payload: str) -> Optional[str]:
    payload = payload.strip()
    if not payload:
        return None
    if payload.startswith("{") and payload.endswith("}"):
        return payload
    start = payload.find("{")
    end = payload.rfind("}")
    if start == -1 or end == -1 or end <= start:
        return None
    return payload[start : end + 1]


def llm_cache_key(
    resource: str,
    relation: str,
    flags: list[dict[str, str]],
    fields: list[tuple[str, str]],
    model: Optional[str],
) -> str:
    payload = {
        "resource": resource,
        "relation": relation,
        "flags": flags,
        "fields": fields,
        "model": model,
    }
    raw = json.dumps(payload, sort_keys=True)
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def load_llm_cache(path: Optional[Path]) -> dict[str, dict[str, Optional[str]]]:
    if not path:
        return {}
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except OSError:
        return {}
    except json.JSONDecodeError:
        return {}
    if not isinstance(data, dict):
        return {}
    cache: dict[str, dict[str, Optional[str]]] = {}
    for key, value in data.items():
        if isinstance(value, dict):
            cache[key] = {str(k): v for k, v in value.items()}
    return cache


def save_llm_cache(path: Optional[Path], cache: dict[str, dict[str, Optional[str]]]) -> None:
    if not path:
        return
    payload = json.dumps(cache, sort_keys=True, indent=2)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(payload + "\n", encoding="utf-8")


def run_llm_flag_mapping(
    resource: str,
    relation: str,
    flags: list[dict[str, str]],
    fields: list[tuple[str, str]],
    model: Optional[str],
) -> dict[str, Optional[str]]:
    if not flags:
        return {}

    allowed_lines = [f"- {name} ({kind})" for name, kind in fields]
    flag_lines = []
    for flag in flags:
        desc = flag.get("description") or ""
        flag_lines.append(f"- {flag['name']}: {desc}".strip())

    prompt = "\n".join(
        [
            "You map CLI flags to resource fields.",
            f"Resource: {resource}",
            f"Relation: {relation}",
            "",
            "Allowed fields (name + kind):",
            *allowed_lines,
            "",
            "Flags to map:",
            *flag_lines,
            "",
            "Return JSON ONLY (no prose).",
            "Schema: {\"--flag-name\": \"field-name\" | null, ...}",
            "Rules:",
            "- Only use field names from the allowed list.",
            "- If no good match exists, use null.",
        ]
    )

    fd, output_path = tempfile.mkstemp(prefix="codex-flag-map-", suffix=".json")
    os.close(fd)
    Path(output_path).unlink(missing_ok=True)
    cmd = [
        "codex",
        "exec",
        "--output-last-message",
        output_path,
        "--dangerously-bypass-approvals-and-sandbox",
        "-",
    ]
    if model:
        cmd.extend(["--model", model])
    try:
        result = subprocess.run(
            cmd,
            input=prompt,
            capture_output=True,
            text=True,
            check=False,
        )
    except OSError:
        return {}

    if result.returncode != 0:
        return {}

    try:
        raw = Path(output_path).read_text(encoding="utf-8", errors="ignore")
    except OSError:
        raw = ""
    Path(output_path).unlink(missing_ok=True)
    payload = extract_json_object(raw)
    if not payload:
        return {}
    try:
        data = json.loads(payload)
    except json.JSONDecodeError:
        return {}
    if not isinstance(data, dict):
        return {}
    return {str(k): v for k, v in data.items()}


def build_command_field_links(
    conn: sqlite3.Connection,
    llm_enabled: bool,
    llm_model: Optional[str],
    llm_workers: int,
    llm_cache: dict[str, dict[str, Optional[str]]],
    llm_cache_path: Optional[Path],
) -> None:
    conn.execute("DELETE FROM command_field_links")

    resource_fields: dict[str, dict[str, str]] = {}
    for resource, name, kind in conn.execute(
        "SELECT resource, name, kind FROM resource_fields"
    ):
        resource_fields.setdefault(resource, {})[name] = kind

    flags_by_command: dict[str, list[dict[str, str]]] = {}
    for command_id, name, description in conn.execute(
        "SELECT command_id, name, description FROM flags"
    ):
        flags_by_command.setdefault(command_id, []).append(
            {"name": name, "description": description}
        )

    command_links = conn.execute(
        "SELECT command_id, resource, verb, command_kind FROM command_resource_links"
    ).fetchall()

    unmatched_by_command: list[dict[str, str]] = []
    unmatched_by_group: dict[tuple[str, str], dict[str, str]] = {}

    for command_id, resource, verb, command_kind in command_links:
        if command_kind == "view":
            relation = "filters_by" if verb == "list" else "selects_field"
        elif command_kind == "do":
            relation = "sets_field" if verb in ("create", "update") else None
        elif command_kind == "summarize":
            relation = "summary_param"
        else:
            relation = None

        if not relation:
            continue

        fields = resource_fields.get(resource)
        if not fields:
            continue

        flags = flags_by_command.get(command_id, [])

        for flag in flags:
            field_name, match_kind, modifier = match_flag_to_field(flag["name"], fields)
            if field_name:
                conn.execute(
                    """
                    INSERT OR IGNORE INTO command_field_links
                    (command_id, resource, field, field_kind, relation, flag_name, match_kind, modifier)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                    """,
                    (
                        command_id,
                        resource,
                        field_name,
                        fields[field_name],
                        relation,
                        flag["name"],
                        match_kind or "exact",
                        modifier,
                    ),
                )
            else:
                unmatched_by_command.append(
                    {
                        "command_id": command_id,
                        "resource": resource,
                        "relation": relation,
                        "flag_name": flag["name"],
                    }
                )
                group_key = (resource, relation)
                if group_key not in unmatched_by_group:
                    unmatched_by_group[group_key] = {}
                unmatched_by_group[group_key].setdefault(
                    flag["name"], flag.get("description") or ""
                )

    if not llm_enabled or not unmatched_by_group:
        return

    tasks: list[tuple[str, str, list[dict[str, str]], list[tuple[str, str]], str]] = []
    mappings: dict[tuple[str, str], dict[str, Optional[str]]] = {}
    cache_changed = False
    cache_hits = 0
    cache_misses = 0

    for (resource, relation), flags_map in unmatched_by_group.items():
        fields = sorted(resource_fields.get(resource, {}).items())
        flags = [{"name": name, "description": desc} for name, desc in sorted(flags_map.items())]
        key = llm_cache_key(resource, relation, flags, fields, llm_model)
        cached = llm_cache.get(key)
        if cached is not None:
            mappings[(resource, relation)] = cached
            cache_hits += 1
        else:
            tasks.append((resource, relation, flags, fields, key))
            cache_misses += 1

    print(
        "LLM flag mapping: {} tasks ({} cache hits, {} cache misses, {} workers)".format(
            len(tasks),
            cache_hits,
            cache_misses,
            llm_workers,
        )
    )

    if tasks:
        with concurrent.futures.ThreadPoolExecutor(max_workers=llm_workers) as executor:
            future_map = {
                executor.submit(
                    run_llm_flag_mapping,
                    resource,
                    relation,
                    flags,
                    fields,
                    llm_model,
                ): (resource, relation, key)
                for resource, relation, flags, fields, key in tasks
            }
            completed = 0
            for future in concurrent.futures.as_completed(future_map):
                resource, relation, key = future_map[future]
                try:
                    result = future.result()
                except Exception:
                    result = {}
                mappings[(resource, relation)] = result
                llm_cache[key] = result
                cache_changed = True
                completed += 1
                print(
                    "LLM flag mapping: completed {}/{} ({})".format(
                        completed,
                        len(tasks),
                        f"{resource}:{relation}",
                    )
                )

    if cache_changed:
        save_llm_cache(llm_cache_path, llm_cache)

    for entry in unmatched_by_command:
        mapping = mappings.get((entry["resource"], entry["relation"]), {})
        mapped_field = mapping.get(entry["flag_name"])
        fields = resource_fields.get(entry["resource"], {})
        if mapped_field and mapped_field in fields:
            conn.execute(
                """
                INSERT OR IGNORE INTO command_field_links
                (command_id, resource, field, field_kind, relation, flag_name, match_kind, modifier)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    entry["command_id"],
                    entry["resource"],
                    mapped_field,
                    fields[mapped_field],
                    entry["relation"],
                    entry["flag_name"],
                    "llm",
                    None,
                ),
            )


def extract_method_block(lines: list[str], method_name: str) -> list[str]:
    start_pattern = re.compile(rf"^\s*def\s+{re.escape(method_name)}\b")
    def_pattern = re.compile(r"^\s*def\b")
    end_pattern = re.compile(r"^\s*end\b")
    collecting = False
    depth = 0
    block: list[str] = []
    for line in lines:
        if not collecting:
            if start_pattern.search(line):
                collecting = True
                depth = 1
            continue
        if def_pattern.search(line):
            depth += 1
        if end_pattern.search(line):
            depth -= 1
            if depth == 0:
                break
        block.append(line)
    return block


def extract_constant_block(lines: list[str], const_name: str) -> list[str]:
    start_pattern = re.compile(rf"\b{re.escape(const_name)}\b\s*=")
    collecting = False
    depth = 0
    block: list[str] = []
    for line in lines:
        if not collecting and start_pattern.search(line):
            collecting = True
        if not collecting:
            continue
        block.append(line)
        depth += line.count("[")
        depth -= line.count("]")
        if collecting and depth <= 0:
            break
    return block


def extract_attribute_names(block: list[str]) -> set[str]:
    text = "\n".join(block)
    return set(re.findall(r'Attribute\.new\(\s*"([^"]+)"', text))


def extract_metric_names(block: list[str]) -> set[str]:
    text = "\n".join(block)
    return set(re.findall(r'Metric\.new\(\s*"([^"]+)"', text))


def extract_constant_attributes(block: list[str]) -> set[str]:
    text = "\n".join(block)
    return set(re.findall(r'attribute:\\s*"([^"]+)"', text))


def build_summary_dimensions_metrics(conn: sqlite3.Connection, server_root: Path) -> None:
    conn.execute("DELETE FROM summary_dimensions")
    conn.execute("DELETE FROM summary_metrics")

    summary_sources = conn.execute(
        "SELECT summary_resource, repo_name, file_path FROM summary_sources"
    ).fetchall()

    summary_files: dict[str, list[str]] = {}
    for summary_resource, repo_name, file_path in summary_sources:
        if repo_name != "server":
            continue
        if not file_path.endswith(".rb"):
            continue
        summary_files.setdefault(summary_resource, []).append(file_path)

    for summary_resource, files in summary_files.items():
        server_path = None
        for candidate in files:
            if "/models/" in candidate and "summary" in candidate:
                server_path = candidate
                break
        if server_path is None:
            server_path = files[0]
        abs_path = server_root / server_path
        try:
            lines = abs_path.read_text(encoding="utf-8", errors="ignore").splitlines()
        except OSError:
            continue

        dimensions: set[str] = set()
        metrics: set[str] = set()

        group_block = extract_constant_block(lines, "GROUP_BY_ATTRIBUTES")
        dimensions |= extract_constant_attributes(group_block)

        metrics_block = extract_constant_block(lines, "METRICS")
        metrics |= extract_constant_attributes(metrics_block)

        summary_attr_block = extract_method_block(lines, "summary_attributes")
        dimensions |= extract_attribute_names(summary_attr_block)

        attributes_block = extract_method_block(lines, "attributes")
        dimensions |= extract_attribute_names(attributes_block)

        metrics_method_block = extract_method_block(lines, "metrics")
        metrics |= extract_metric_names(metrics_method_block)

        for name in sorted(dimensions):
            conn.execute(
                """
                INSERT OR IGNORE INTO summary_dimensions
                (summary_resource, name, kind, source_path)
                VALUES (?, ?, ?, ?)
                """,
                (summary_resource, name, "group_by", server_path),
            )

        for name in sorted(metrics):
            conn.execute(
                """
                INSERT OR IGNORE INTO summary_metrics
                (summary_resource, name, source_path)
                VALUES (?, ?, ?)
                """,
                (summary_resource, name, server_path),
            )


def build_command_summary_links(conn: sqlite3.Connection) -> None:
    conn.execute("DELETE FROM command_summary_dimensions")
    conn.execute("DELETE FROM command_summary_metrics")

    summary_dimensions: dict[str, list[tuple[str, str]]] = {}
    for summary_resource, name, source_path in conn.execute(
        "SELECT summary_resource, name, source_path FROM summary_dimensions"
    ):
        summary_dimensions.setdefault(summary_resource, []).append((name, source_path))

    summary_metrics: dict[str, list[tuple[str, str]]] = {}
    for summary_resource, name, source_path in conn.execute(
        "SELECT summary_resource, name, source_path FROM summary_metrics"
    ):
        summary_metrics.setdefault(summary_resource, []).append((name, source_path))

    for command_id, resource in conn.execute(
        "SELECT command_id, resource FROM command_resource_links WHERE command_kind = 'summarize'"
    ):
        for name, source_path in summary_dimensions.get(resource, []):
            conn.execute(
                """
                INSERT OR IGNORE INTO command_summary_dimensions
                (command_id, summary_resource, name, source_path)
                VALUES (?, ?, ?, ?)
                """,
                (command_id, resource, name, source_path),
            )
        for name, source_path in summary_metrics.get(resource, []):
            conn.execute(
                """
                INSERT OR IGNORE INTO command_summary_metrics
                (command_id, summary_resource, name, source_path)
                VALUES (?, ?, ?, ?)
                """,
                (command_id, resource, name, source_path),
            )


def pluralize_resource_name(base: str) -> str:
    if base.endswith("s"):
        return base
    if base.endswith("y") and len(base) > 1 and base[-2] not in "aeiou":
        return base[:-1] + "ies"
    return base + "s"


def candidate_resource_names(base: str, resources: set[str]) -> list[str]:
    candidates = []
    if base in resources:
        candidates.append(base)
    plural = pluralize_resource_name(base)
    if plural in resources and plural not in candidates:
        candidates.append(plural)
    return candidates


def build_relationship_graph(conn: sqlite3.Connection) -> dict[str, list[tuple[str, str]]]:
    adjacency: dict[str, list[tuple[str, str]]] = {}
    for resource, field, target_resource in conn.execute(
        "SELECT resource, field, target_resource FROM resource_field_targets"
    ):
        adjacency.setdefault(resource, []).append((field, target_resource))
    return adjacency


def shortest_paths(
    adjacency: dict[str, list[tuple[str, str]]], start: str, max_hops: int
) -> dict[str, list[str]]:
    paths: dict[str, list[str]] = {start: []}
    queue: list[str] = [start]
    while queue:
        current = queue.pop(0)
        path = paths[current]
        if len(path) >= max_hops:
            continue
        for rel_name, target in adjacency.get(current, []):
            if target in paths:
                continue
            paths[target] = path + [rel_name]
            queue.append(target)
    return paths


def parse_flag_base_modifier(flag_name: str) -> tuple[str, Optional[str]]:
    base = normalize_flag_name(flag_name)
    if base.startswith("not-"):
        return base[4:], "not"
    if base.startswith("is-"):
        return base[3:], "is"
    for suffix, modifier in (
        ("-min", "min"),
        ("-max", "max"),
        ("-before", "before"),
        ("-after", "after"),
    ):
        if base.endswith(suffix):
            return base[: -len(suffix)], modifier
    return base, None


def build_command_filter_paths(conn: sqlite3.Connection, max_hops: int = 3) -> None:
    conn.execute("DELETE FROM command_filter_paths")

    resources = {row[0] for row in conn.execute("SELECT name FROM resources")}

    attribute_map: dict[str, set[str]] = {}
    for resource, name in conn.execute(
        "SELECT resource, name FROM resource_fields WHERE kind = 'attribute'"
    ):
        attribute_map.setdefault(resource, set()).add(name)

    adjacency = build_relationship_graph(conn)

    unmapped_flags = conn.execute(
        """
        SELECT f.command_id, f.name, crl.resource
        FROM flags f
        JOIN command_resource_links crl ON crl.command_id = f.command_id
        WHERE crl.command_kind = 'view' AND crl.verb = 'list'
        AND NOT EXISTS (
            SELECT 1 FROM command_field_links cfl
            WHERE cfl.command_id = f.command_id AND cfl.flag_name = f.name
        )
        """
    ).fetchall()

    for command_id, flag_name, resource in unmapped_flags:
        base, modifier = parse_flag_base_modifier(flag_name)
        if base.endswith("-id"):
            base = base[:-3]
        elif base.endswith("-ids"):
            base = base[:-4]

        if base in {rel for rel, _ in adjacency.get(resource, [])}:
            for rel_name, target in adjacency.get(resource, []):
                if rel_name != base:
                    continue
                conn.execute(
                    """
                    INSERT OR IGNORE INTO command_filter_paths
                    (command_id, resource, flag_name, path, target_resource, target_field, hop_count, match_kind, modifier, source)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                    """,
                    (
                        command_id,
                        resource,
                        flag_name,
                        rel_name,
                        target,
                        None,
                        1,
                        "rel",
                        modifier,
                        "deterministic",
                    ),
                )
            continue

        candidates = candidate_resource_names(base, resources)
        paths = shortest_paths(adjacency, resource, max_hops)

        matched_any = False
        if candidates:
            min_hop = None
            for candidate in candidates:
                if candidate in paths and candidate != resource:
                    hop_count = len(paths[candidate])
                    if min_hop is None or hop_count < min_hop:
                        min_hop = hop_count
            if min_hop is not None:
                for candidate in candidates:
                    if candidate in paths and len(paths[candidate]) == min_hop:
                        path = ".".join(paths[candidate])
                        conn.execute(
                            """
                            INSERT OR IGNORE INTO command_filter_paths
                            (command_id, resource, flag_name, path, target_resource, target_field, hop_count, match_kind, modifier, source)
                            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                            """,
                            (
                                command_id,
                                resource,
                                flag_name,
                                path,
                                candidate,
                                None,
                                min_hop,
                                "rel_resource",
                                modifier,
                                "deterministic",
                            ),
                        )
                        matched_any = True

        if matched_any:
            continue

        attr_matches: list[tuple[str, list[str]]] = []
        for target_resource, attrs in attribute_map.items():
            if base not in attrs:
                continue
            if target_resource not in paths:
                continue
            if target_resource == resource:
                path_parts = []
            else:
                path_parts = paths[target_resource]
            attr_matches.append((target_resource, path_parts))

        if not attr_matches:
            continue

        min_hop = min(len(path) for _, path in attr_matches)
        for target_resource, path_parts in attr_matches:
            if len(path_parts) != min_hop:
                continue
            path = ".".join(path_parts + [base]) if path_parts else base
            conn.execute(
                """
                INSERT OR IGNORE INTO command_filter_paths
                (command_id, resource, flag_name, path, target_resource, target_field, hop_count, match_kind, modifier, source)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    command_id,
                    resource,
                    flag_name,
                    path,
                    target_resource,
                    base,
                    min_hop,
                    "rel_attr",
                    modifier,
                    "deterministic",
                ),
            )


def build_command_resource_links(conn: sqlite3.Connection, repo_root: Path) -> None:
    conn.execute("DELETE FROM command_resource_links")

    resources = {row[0] for row in conn.execute("SELECT name FROM resources")}
    commands = conn.execute("SELECT id, full_path FROM commands").fetchall()
    cli_sources: dict[str, list[str]] = {}
    for command_id, file_path in conn.execute(
        "SELECT command_id, file_path FROM sources WHERE repo_name = 'cli'"
    ):
        cli_sources.setdefault(command_id, []).append(file_path)

    endpoint_pattern = re.compile(r"\"/v1/([a-z0-9-]+)\"")

    for command_id, full_path in commands:
        tokens = full_path.split()
        if not tokens:
            continue
        kind = tokens[0]
        if kind == "view" and len(tokens) >= 3 and tokens[-1] in ("list", "show"):
            resource = " ".join(tokens[1:-1])
            if resource in resources:
                conn.execute(
                    """
                    INSERT OR IGNORE INTO command_resource_links
                    (command_id, resource, verb, command_kind, source, evidence)
                    VALUES (?, ?, ?, ?, ?, ?)
                    """,
                    (command_id, resource, tokens[-1], "view", "full_path", None),
                )
            continue

        if kind == "do" and len(tokens) >= 3:
            resource = " ".join(tokens[1:-1])
            verb = tokens[-1]
            if resource in resources:
                conn.execute(
                    """
                    INSERT OR IGNORE INTO command_resource_links
                    (command_id, resource, verb, command_kind, source, evidence)
                    VALUES (?, ?, ?, ?, ?, ?)
                    """,
                    (command_id, resource, verb, "do", "full_path", None),
                )
            continue

        if kind == "summarize" and len(tokens) >= 3 and tokens[-1] == "create":
            resource = None
            evidence = None
            for path in cli_sources.get(command_id, []):
                abs_path = repo_root / path
                try:
                    content = abs_path.read_text(encoding="utf-8", errors="ignore")
                except OSError:
                    continue
                match = endpoint_pattern.search(content)
                if match:
                    resource = match.group(1)
                    evidence = path
                    break
            if resource is None:
                alias = " ".join(tokens[1:-1])
                candidate = f"{alias}s"
                if candidate in resources:
                    resource = candidate
                    evidence = "alias_plural"
            if resource and resource in resources:
                conn.execute(
                    """
                    INSERT OR IGNORE INTO command_resource_links
                    (command_id, resource, verb, command_kind, source, evidence)
                    VALUES (?, ?, ?, ?, ?, ?)
                    """,
                    (command_id, resource, "create", "summarize", "cli_endpoint", evidence),
                )


def main() -> None:
    parser = argparse.ArgumentParser(description="Compile Cartographer artifacts into SQLite")
    parser.add_argument("--config", default="config.yaml", help="Path to config.yaml")
    parser.add_argument(
        "--resource-map",
        default=None,
        help="Path to resource_map.json (defaults to internal/cli/resource_map.json)",
    )
    parser.add_argument(
        "--summary-map",
        default=None,
        help="Path to summary_map.json (defaults to internal/cli/summary_map.json)",
    )
    parser.add_argument(
        "--llm",
        action="store_true",
        help="Use LLM fallback to map unmapped flags to resource fields",
    )
    parser.add_argument(
        "--llm-model",
        default="gpt-5.2-codex",
        help="Model name for LLM fallback",
    )
    parser.add_argument(
        "--llm-workers",
        type=int,
        default=8,
        help="Number of parallel LLM workers (default: 8)",
    )
    parser.add_argument(
        "--llm-cache",
        default=None,
        help="Path to LLM cache JSON (defaults to cartographer_out/db/llm_flag_cache.json)",
    )
    args = parser.parse_args()

    config = load_config(args.config)
    root_out = project_root(config)
    server_root = find_repo_root(config, "server")
    artifacts_root = root_out / "artifacts"
    db_path = root_out / "db" / "knowledge.sqlite"
    repo_root = Path(__file__).resolve().parents[1]
    resource_map_path = Path(args.resource_map) if args.resource_map else repo_root / "internal/cli/resource_map.json"
    summary_map_path = Path(args.summary_map) if args.summary_map else repo_root / "internal/cli/summary_map.json"
    llm_cache_path = (
        Path(args.llm_cache)
        if args.llm_cache
        else root_out / "db" / "llm_flag_cache.json"
    )

    conn = init_db(db_path)
    resource_map = load_resource_map(resource_map_path)
    upsert_resource_map(conn, resource_map)
    summary_map = load_summary_map(summary_map_path)
    upsert_summary_map(conn, summary_map)
    build_command_resource_links(conn, repo_root)
    build_summary_dimensions_metrics(conn, server_root)
    build_command_summary_links(conn)
    build_command_filter_paths(conn, max_hops=3)
    llm_cache = load_llm_cache(llm_cache_path if args.llm else None)
    build_command_field_links(
        conn,
        args.llm,
        args.llm_model,
        args.llm_workers,
        llm_cache,
        llm_cache_path if args.llm else None,
    )
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
    embedded_db_path = repo_root / "internal" / "cli" / "knowledge_db" / "knowledge.sqlite"
    embedded_db_path.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(db_path, embedded_db_path)
    print(f"Compiled {inserted} artifacts into {db_path} (skipped {skipped})")


if __name__ == "__main__":
    main()
