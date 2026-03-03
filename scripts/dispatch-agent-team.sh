#!/usr/bin/env bash
set -euo pipefail

mode="${1:-execute}"
qa_full_test="${2:-false}"

if [[ "$mode" != "plan" && "$mode" != "execute" ]]; then
  echo "mode must be 'plan' or 'execute'"
  exit 2
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required"
  exit 2
fi

if [[ -z "${GH_TOKEN:-}" ]]; then
  echo "GH_TOKEN is required (use org bot token with workflow permission)"
  exit 2
fi

# repo:company
pairs=(
  "GoPolymarket/go-builder-relayer-client:platform-core"
  "GoPolymarket/polymarket-go-sdk:platform-core"
  "GoPolymarket/pmx-alpha:company-a"
  "GoPolymarket/pmx-gateway:company-b"
  "GoPolymarket/pmx-builder-ops:company-c"
)

echo "Dispatching agent-team-delivery.yml to ${#pairs[@]} repositories"
for pair in "${pairs[@]}"; do
  target_repo="${pair%%:*}"
  company="${pair##*:}"
  echo "-> $target_repo (company=$company mode=$mode qa_full_test=$qa_full_test)"

  gh workflow run agent-team-delivery.yml \
    --repo "$target_repo" \
    -f mode="$mode" \
    -f company="$company" \
    -f qa_full_test="$qa_full_test"
done

echo "Dispatch complete."
