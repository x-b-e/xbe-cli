# Releasing xbe-cli

This repo uses GoReleaser and GitHub Actions. Releases are published from tags.

## One-time setup
- Ensure GoReleaser config is present at `.goreleaser.yaml`.
- GitHub Actions workflow is at `.github/workflows/release.yml`.

## Release steps
1) Update `VERSION` (for manual builds) and ensure `main` is green with a clean working tree.
2) Create an annotated tag and push it:

```
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

3) The GitHub Action builds and publishes release assets + `checksums.txt`.
4) Verify by downloading the release and running `xbe version`.

## Local dry run
```
goreleaser release --snapshot --clean
```
