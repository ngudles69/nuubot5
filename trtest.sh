#!/usr/bin/env bash

set -u -o pipefail

runs="${1:-1}"
sweep_id="${2:-9}"
bot_id="${3:-13}"
for value in "$runs" "$sweep_id" "$bot_id"; do
    if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
        echo "usage: bash trtest.sh [runs] [sweep_id] [bot_id]" >&2
        exit 2
    fi
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if ! bash "$repo_root/build.sh"; then
    exit 1
fi
suffix=""
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) suffix=".exe" ;;
esac
binary="$repo_root/bin/nuubot-btrunner${suffix}"
if [[ ! -x "$binary" ]]; then
    echo "Go binary not found: $binary" >&2
    exit 2
fi
config="$repo_root/workspace/config/tradeexecutor.toml"
log_dir="$repo_root/workspace/logs"
mkdir -p "$log_dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
result_log="$log_dir/nuubot5-trtest-s${sweep_id}-b${bot_id}-${runs}-${stamp}.log"
exec > >(tee -a "$result_log") 2>&1

passed=0
process_total_ms=0
replay_total_ms=0
suite_started_ms="$(date +%s%3N)"
bot_log="$log_dir/bot_${sweep_id}_${bot_id}.log"
result_db="$repo_root/workspace/db/sweeps/sweep_${sweep_id}/bot_${bot_id}.db"

for ((run = 1; run <= runs; run++)); do
    before_lines=0
    if [[ -f "$bot_log" ]]; then
        before_lines="$(wc -l < "$bot_log")"
    fi
    started_ms="$(date +%s%3N)"
    output="$(
        cd "$repo_root" &&
        NUUBOT_CONFIG="$config" timeout 120s "$binary" "$sweep_id" "$bot_id" 2>&1
    )"
    status=$?
    process_ms=$(( $(date +%s%3N) - started_ms ))
    if [[ -f "$bot_log" ]]; then
        output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
    fi

    btrunner_line="$(printf '%s\n' "$output" | grep '] btrunner stopped ' | tail -n 1)"
    runtime_line="$(printf '%s\n' "$output" | grep '] runtime stopped ' | tail -n 1)"
    replay_ms="$(printf '%s\n' "$btrunner_line" | sed -n 's/.*replay_ms=\([0-9][0-9]*\).*/\1/p')"
    ticks="$(printf '%s\n' "$runtime_line" | sed -n 's/.*ticks_accepted=\([0-9][0-9]*\).*/\1/p')"
    runtime_runs="$(printf '%s\n' "$runtime_line" | sed -n 's/.*runs=\([0-9][0-9]*\).*/\1/p')"
    cycles="$(printf '%s\n' "$runtime_line" | sed -n 's/.*cycles_closed=\([0-9][0-9]*\).*/\1/p')"
    trade_lines="$(printf '%s\n' "$output" | grep '] executor stopped .*kind=trade ' || true)"
    trade_cycles="$(printf '%s\n' "$trade_lines" | grep -c .)"
    trades="$(printf '%s\n' "$trade_lines" | awk '
        { for (i = 1; i <= NF; i++) if ($i ~ /^trades=/) {
            split($i, value, "="); total += value[2]
        }}
        END { print total + 0 }
    ')"
    fills="$(printf '%s\n' "$trade_lines" | awk '
        { for (i = 1; i <= NF; i++) if ($i ~ /^fills=/) {
            split($i, value, "="); total += value[2]
        }}
        END { print total + 0 }
    ')"
    account_lines="$(printf '%s\n' "$output" | grep '] account stopped ' || true)"
    orders="$(printf '%s\n' "$account_lines" | awk '
        { for (i = 1; i <= NF; i++) if ($i ~ /^orders=/) {
            split($i, value, "="); total += value[2]
        }}
        END { print total + 0 }
    ')"

    if [[ $status -ne 0 || -z "$replay_ms" || "$ticks" != "7948800" ||
          "$runtime_runs" != "794880" || -z "$cycles" ||
          "$trade_cycles" != "$cycles" || "$trades" != "$cycles" ||
          "$fills" != "$((cycles * 2))" || "$orders" != "$((cycles * 3 + 1))" ||
          ! -f "$result_db" || -f "${result_db}.partial" ]]; then
        printf '%s\n' "$output"
        printf 'run=%d result=FAIL exit=%d process_ms=%d replay_ms=%s cycles=%s trades=%s orders=%s fills=%s\n' \
            "$run" "$status" "$process_ms" "${replay_ms:-missing}" "${cycles:-missing}" \
            "$trades" "$orders" "$fills"
        exit 1
    fi

    process_total_ms=$((process_total_ms + process_ms))
    replay_total_ms=$((replay_total_ms + replay_ms))
    ((passed += 1))
    printf '%s\n' "$runtime_line"
    printf '%s\n' "$btrunner_line"
    printf 'run=%d result=PASS process_ms=%d replay_ms=%d ticks=%s runs=%s cycles=%s trades=%s orders=%s fills=%s result_db=%s\n' \
        "$run" "$process_ms" "$replay_ms" "$ticks" "$runtime_runs" "$cycles" \
        "$trades" "$orders" "$fills" "$result_db"
done

printf 'requested=%d attempted=%d passed=%d failed=0 suite_ms=%d process_total_ms=%d process_average_ms=%d replay_total_ms=%d replay_average_ms=%d log=%s\n' \
    "$runs" "$runs" "$passed" "$(( $(date +%s%3N) - suite_started_ms ))" \
    "$process_total_ms" "$((process_total_ms / runs))" \
    "$replay_total_ms" "$((replay_total_ms / runs))" "$result_log"
