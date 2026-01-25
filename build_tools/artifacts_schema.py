from typing import List, Literal, Optional

from pydantic import BaseModel, Field


class Flag(BaseModel):
    name: str
    aliases: Optional[List[str]] = None
    required: bool
    type: Literal["string", "boolean", "integer", "array", "enum"]
    description: str
    default: Optional[str] = None
    validation: Optional[str] = None


class SourceRef(BaseModel):
    repo_name: str
    file_path: str


class CommandArtifact(BaseModel):
    id: str
    full_path: str
    description: str
    permissions: Optional[str] = None
    side_effects: Optional[str] = None
    validation_notes: Optional[str] = None
    flags: List[Flag] = Field(default_factory=list)
    sources: List[SourceRef] = Field(default_factory=list)


def validate_artifact(data: dict) -> CommandArtifact:
    if hasattr(CommandArtifact, "model_validate"):
        return CommandArtifact.model_validate(data)
    return CommandArtifact.parse_obj(data)
