// The Atomic Unit of the Knowledge Base
export interface CommandArtifact {
  // IDENTITY
  id: string; // Hash of full_path
  full_path: string; // e.g., "cli network firewall create"
  description: string;

  // CONTEXT (Human Readable)
  // If not applicable, OMIT or set to null.
  permissions?: string;   // e.g. "Restricted to Network Admins."
  side_effects?: string;  // e.g. "Triggers billing event."
  validation_notes?: string; // e.g. "Names must be lowercase."

  // SYNTAX (Rigid Logic)
  flags: Flag[];

  // PROVENANCE (Audit Trail)
  sources: SourceRef[];
}

export interface Flag {
  name: string;      // e.g., "--vpc-id"
  aliases?: string[];
  required: boolean;
  type: "string" | "boolean" | "integer" | "array" | "enum";
  description: string;
  default?: string | null;
  validation?: string | null; // e.g. "Must be a UUIDv4."
}

export interface SourceRef {
  repo_name: string; // Matches config (e.g., "server")
  file_path: string; // Relative path in that repo
}
