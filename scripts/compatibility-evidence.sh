#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<"USAGE"
usage: scripts/compatibility-evidence.sh [--week YYYY-Www] [--out-dir path] [--command-center path]

Generate weekly compatibility evidence artifact for polymarket-go-sdk.
USAGE
}

week="$(date -u +%G-W%V)"
out_dir="${ARTIFACTS_DIR:-artifacts/compatibility}"
command_center=""

while (($#)); do
  case "$1" in
    --week)
      week="${2:-}"
      shift 2
      ;;
    --out-dir)
      out_dir="${2:-}"
      shift 2
      ;;
    --command-center)
      command_center="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ ! "$week" =~ ^[0-9]{4}-W[0-9]{2}$ ]]; then
  echo "invalid --week value: $week (expected YYYY-Www)" >&2
  exit 2
fi

if [[ -n "$command_center" && ! -f "$command_center" ]]; then
  echo "command center output not found: $command_center" >&2
  exit 2
fi

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
repo_name="$(basename "$repo_root")"
commit_sha="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

run_dir="$out_dir/$week"
mkdir -p "$run_dir"
run_dir_abs="$(cd "$run_dir" && pwd)"
gocache="${GOCACHE:-$run_dir_abs/.gocache}"
mkdir -p "$gocache"

checks_file="$run_dir/checks.tsv"
: > "$checks_file"
failed=0

run_check() {
  local check_name="$1"
  local cmd="$2"
  local log_file="$run_dir/${check_name}.log"
  local status="pass"

  if ! env GOCACHE="$gocache" bash -lc "$cmd" >"$log_file" 2>&1; then
    status="fail"
    failed=1
  fi

  local checksum
  checksum="$(shasum -a 256 "$log_file" | awk '{print $1}')"
  printf '%s\t%s\t%s\t%s\t%s\n' "$check_name" "$cmd" "$status" "$log_file" "$checksum" >> "$checks_file"
}

run_check "ws_schema_compat" "go test ./pkg/clob/ws -run '^TestProcessEvent_SchemaCompat_' -count=1"
run_check "pkg_compile_gate" "go test ./pkg/... -run '^$' -count=1"

manifest_file="$run_dir/manifest.json"
summary_file="$run_dir/summary.md"

python3 - "$repo_name" "$week" "$commit_sha" "$generated_at" "$checks_file" "$command_center" "$manifest_file" "$summary_file" "$gocache" <<'PY'
import hashlib
import json
import sys

(
    repo_name,
    week,
    commit_sha,
    generated_at,
    checks_file,
    command_center,
    manifest_file,
    summary_file,
    gocache,
) = sys.argv[1:]

checks = []
with open(checks_file, "r", encoding="utf-8") as f:
    for raw in f:
        raw = raw.rstrip("\n")
        if not raw:
            continue
        name, command, status, log_file, log_checksum = raw.split("\t", 4)
        checks.append(
            {
                "name": name,
                "command": command,
                "status": status,
                "log_file": log_file,
                "log_checksum_sha256": log_checksum,
            }
        )

command_center_meta = None
if command_center:
    with open(command_center, "rb") as f:
        command_center_meta = {
            "path": command_center,
            "checksum_sha256": hashlib.sha256(f.read()).hexdigest(),
        }

all_passed = all(item["status"] == "pass" for item in checks)
manifest = {
    "artifact_id": f"compat-{repo_name.lower()}-{week.lower()}",
    "repo": repo_name,
    "week": week,
    "generated_at_utc": generated_at,
    "commit_sha": commit_sha,
    "all_passed": all_passed,
    "go_cache": gocache,
    "checks": checks,
    "command_center": command_center_meta,
}

canonical = json.dumps(manifest, sort_keys=True, separators=(",", ":")).encode("utf-8")
manifest["manifest_checksum_sha256"] = hashlib.sha256(canonical).hexdigest()

with open(manifest_file, "w", encoding="utf-8") as f:
    json.dump(manifest, f, indent=2, sort_keys=True)
    f.write("\n")

lines = [
    "# Weekly Compatibility Evidence",
    "",
    f"- repo: {repo_name}",
    f"- week: {week}",
    f"- commit_sha: {commit_sha}",
    f"- generated_at_utc: {generated_at}",
    f"- all_passed: {'true' if all_passed else 'false'}",
    f"- go_cache: {gocache}",
    f"- manifest_checksum_sha256: {manifest['manifest_checksum_sha256']}",
]
if command_center_meta:
    lines.extend(
        [
            f"- command_center_path: {command_center_meta['path']}",
            f"- command_center_checksum_sha256: {command_center_meta['checksum_sha256']}",
        ]
    )
lines.append("")
lines.append("## Checks")
for check in checks:
    lines.append("")
    lines.append(f"- name: {check['name']}")
    lines.append(f"- status: {check['status']}")
    lines.append(f"- command: `{check['command']}`")
    lines.append(f"- log_file: `{check['log_file']}`")
    lines.append(f"- log_checksum_sha256: {check['log_checksum_sha256']}")

with open(summary_file, "w", encoding="utf-8") as f:
    f.write("\n".join(lines) + "\n")
PY

archive_file="$run_dir/${repo_name}-${week}-compatibility-artifact.tgz"
(
  cd "$run_dir"
  tar -czf "$(basename "$archive_file")" manifest.json summary.md ./*.log
)

archive_checksum="$(shasum -a 256 "$archive_file" | awk '{print $1}')"
echo "$archive_checksum  $(basename "$archive_file")" > "$run_dir/archive.sha256"

echo "manifest=$manifest_file"
echo "summary=$summary_file"
echo "archive=$archive_file"
echo "archive_checksum_sha256=$archive_checksum"

if [[ "$failed" -ne 0 ]]; then
  echo "one or more compatibility checks failed; see $summary_file" >&2
  exit 1
fi
