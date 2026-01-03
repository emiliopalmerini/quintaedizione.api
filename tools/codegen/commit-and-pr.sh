#!/bin/bash

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

MODULES=(
    "incantesimi"
    "mostri"
    "oggetti"
    "maestrie"
    "talenti"
    "specie"
    "background"
    "regole"
    "condizioni"
    "linguaggi"
    "divinita"
    "bastioni"
)

PREV_PR="3"

for module in "${MODULES[@]}"; do
    BRANCH_NAME="feat/api-$module"

    echo ""
    echo "=========================================="
    echo "Processing: $BRANCH_NAME"
    echo "=========================================="

    git checkout "$BRANCH_NAME"

    # Check if there are changes to commit
    if [[ -n $(git status --porcelain) ]]; then
        git add .
        git commit -m "feat($module): implement /v1/$module API endpoint

- Add $module module with service, repository, and HTTP handler
- Add database migration for $module table
- Follow vertical slice architecture pattern"

        git push -u origin "$BRANCH_NAME"

        # Create PR
        PR_URL=$(gh pr create \
            --title "feat($module): implement /v1/$module API" \
            --body "## Summary

- Implement \`/v1/$module\` endpoint with pagination and filtering
- Add database migration for \`$module\` table
- Follow the same architecture pattern as \`classi\` module

Follows #$PREV_PR

## Test plan
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual API testing" \
            --base main 2>&1 || echo "PR may already exist")

        echo "Created/Updated PR for $module"

        # Extract PR number for next iteration
        if [[ "$PR_URL" =~ pull/([0-9]+) ]]; then
            PREV_PR="${BASH_REMATCH[1]}"
        fi
    else
        echo "No changes to commit for $module"
    fi
done

echo ""
echo "=========================================="
echo "All PRs created!"
echo "=========================================="
