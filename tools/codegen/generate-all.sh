#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONFIGS_DIR="$SCRIPT_DIR/configs"

cd "$PROJECT_ROOT"

echo "Building codegen tool..."
go build -o "$SCRIPT_DIR/codegen" "$SCRIPT_DIR/main.go"

MODULES=(
    "incantesimi:3"
    "mostri:4"
    "oggetti:5"
    "maestrie:6"
    "talenti:7"
    "specie:8"
    "background:9"
    "regole:10"
    "condizioni:11"
    "linguaggi:12"
    "divinita:13"
    "bastioni:14"
)

BASE_BRANCH="feat/api-classi"
PREV_PR="3"

for module_info in "${MODULES[@]}"; do
    IFS=':' read -r module migration_num <<< "$module_info"

    echo ""
    echo "=========================================="
    echo "Processing module: $module"
    echo "=========================================="

    BRANCH_NAME="feat/api-$module"

    # Checkout base and create new branch
    git checkout "$BASE_BRANCH"
    git checkout -B "$BRANCH_NAME"

    # Generate the module
    "$SCRIPT_DIR/codegen" \
        -config "$CONFIGS_DIR/$module.json" \
        -output "$PROJECT_ROOT/internal" \
        -migrations "$PROJECT_ROOT/migrations" \
        -migration-num "$migration_num"

    echo "Generated module: $module"

    # Update BASE_BRANCH for next iteration
    BASE_BRANCH="$BRANCH_NAME"
done

echo ""
echo "=========================================="
echo "All modules generated!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Review generated code in each branch"
echo "2. Update models.go with proper fields from swagger"
echo "3. Update app.go to register routes"
echo "4. Run: ./tools/codegen/commit-and-pr.sh"
