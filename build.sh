#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
go_bin="${GO_BIN:-}"
if [[ -z "$go_bin" ]]; then
    if command -v go >/dev/null 2>&1; then
        go_bin="$(command -v go)"
    elif [[ -x /c/Users/PC/.local/go1.26.5/go/bin/go.exe ]]; then
        go_bin=/c/Users/PC/.local/go1.26.5/go/bin/go.exe
    else
        echo "Go not found; set GO_BIN" >&2
        exit 2
    fi
fi

suffix=""
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) suffix=".exe" ;;
esac

mkdir -p "$repo_root/bin"
"$go_bin" build -buildvcs=false -tags noasm \
    -o "$repo_root/bin/nuubot-btrunner${suffix}" \
    "$repo_root/cmd/nuubot-btrunner"
