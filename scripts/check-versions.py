"""Check the release files without requiring third-party Python packages."""

from pathlib import Path
import re


version = Path("version").read_text()
assert re.fullmatch(r"\d+\.\d+\.\d+", version), "version must contain only X.Y.Z (no newline)"
values = Path("helm/values.yaml").read_text()
chart = Path("helm/Chart.yaml").read_text()
tag = re.search(r"^  tag:\s*[\"']?([^\s\"']+)", values, re.MULTILINE)
app = re.search(r"^appVersion:\s*[\"']?([^\s\"']+)", chart, re.MULTILINE)
assert tag and tag.group(1) == version, "Helm image.tag differs from version"
assert app and app.group(1) == version, "Helm appVersion differs from version"
print(f"Release versions match: {version}")
