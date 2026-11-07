Read `version` and `.claude/skills/version-management/SKILL.md`.
For a release, bump the requested semantic version and synchronize Helm image.tag/appVersion.
If the chart changed, bump its version too. Run checks before committing.
Documentation-only changes do not require a version bump. Tag publication and cluster deployment are separate actions.
