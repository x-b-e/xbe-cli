#!/usr/bin/env python3
import argparse
import re
from pathlib import Path


def to_camel(resource: str) -> str:
    parts = re.split(r"[-_]+", resource)
    return "".join(part[:1].upper() + part[1:] for part in parts if part)


def resource_to_filename(resource: str) -> str:
    return resource.replace("-", "_") + "_show.go"


def parse_list_commands(cli_dir: Path) -> dict[str, str]:
    mapping: dict[str, str] = {}
    add_re = re.compile(r"(\w+Cmd)\.AddCommand\(")
    for path in cli_dir.glob("*_list.go"):
        stem = path.stem
        if stem.startswith("do_"):
            continue
        resource = stem[:-5].replace("_", "-")
        text = path.read_text(encoding="utf-8")
        match = add_re.search(text)
        if not match:
            continue
        mapping[resource] = match.group(1)
    return mapping


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Generate generic show commands for resources that have list commands"
    )
    parser.add_argument("--cli-dir", default="internal/cli", help="Path to internal/cli directory")
    parser.add_argument(
        "--force",
        action="store_true",
        help="Overwrite existing show files (use with care; may replace custom handlers)",
    )
    args = parser.parse_args()

    cli_dir = Path(args.cli_dir)
    resources = sorted(parse_list_commands(cli_dir).keys())

    existing_show = {
        path.stem[:-5].replace("_", "-"): path for path in cli_dir.glob("*_show.go")
    }

    cmd_vars = parse_list_commands(cli_dir)
    generated = 0
    skipped = []

    for resource in resources:
        if resource in existing_show and not args.force:
            continue
        var_name = cmd_vars.get(resource)
        if not var_name:
            skipped.append(resource)
            continue

        go_name = to_camel(resource)
        filename = resource_to_filename(resource)
        path = cli_dir / filename
        content = "\n".join(
            [
                "package cli",
                "",
                "import \"github.com/spf13/cobra\"",
                "",
                f"func new{go_name}ShowCmd() *cobra.Command {{",
                f"\treturn newGenericShowCmd(\"{resource}\")",
                "}",
                "",
                "func init() {",
                f"\t{var_name}.AddCommand(new{go_name}ShowCmd())",
                "}",
                "",
            ]
        )
        path.write_text(content, encoding="utf-8")
        generated += 1

    print(f"Generated {generated} show commands")
    if skipped:
        print(f"Skipped {len(skipped)} resources without command var")
        for resource in skipped[:20]:
            print(f"  - {resource}")


if __name__ == "__main__":
    main()
