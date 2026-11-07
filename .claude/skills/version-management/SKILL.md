# Version management

- Source: `version` (single value, no trailing newline).
- Keep `helm/values.yaml` image.tag and `helm/Chart.yaml` appVersion identical.
- Increment chart version when chart configuration changes.
- Release code/image changes with a new version; documentation-only changes do not require a version bump.
- Run local checks, commit, push main, then push the matching `vX.Y.Z` tag.
- GitHub Actions validates the tag and publishes `ghcr.io/icurfer/kube-node-autodelete:X.Y.Z` and a commit tag.
- Do not reuse a published version or mutate source files from CI.
- Deploy manually with Helm after checking the published image digest.
