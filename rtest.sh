#!/usr/bin/env bash

set -u -o pipefail

runs="${1:-5}"
sweep_id="${2:-6}"
bot_id="${3:-9}"
for value in "$runs" "$sweep_id" "$bot_id"; do
    if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
        echo "usage: bash rtest.sh [runs] [sweep_id] [bot_id]" >&2
        exit 2
    fi
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$repo_root/build.sh"
suffix=""
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) suffix=".exe" ;;
esac
binary="$repo_root/bin/nuubot-btrunner${suffix}"
if [[ ! -x "$binary" ]]; then
    echo "Go binary not found: $binary" >&2
    exit 2
fi

log_dir="$repo_root/workspace/logs"
mkdir -p "$log_dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
result_log="$log_dir/nuubot5-rtest-s${sweep_id}-b${bot_id}-${runs}-${stamp}.log"
exec > >(tee -a "$result_log") 2>&1

passed=0
process_total_ms=0
process_minimum_ms=0
process_maximum_ms=0
replay_total_ms=0
replay_minimum_ms=0
replay_maximum_ms=0
suite_started_ms="$(date +%s%3N)"
bot_log="$log_dir/bot_${sweep_id}_${bot_id}.log"

for ((run = 1; run <= runs; run++)); do
    before_lines=0
    if [[ -f "$bot_log" ]]; then
        before_lines="$(wc -l < "$bot_log")"
    fi
    started_ms="$(date +%s%3N)"
    output="$(cd "$repo_root" && timeout 120s "$binary" "$sweep_id" "$bot_id" 2>&1)"
    status=$?
    elapsed_ms=$(( $(date +%s%3N) - started_ms ))
    if [[ -f "$bot_log" ]]; then
        output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
    fi
    process_total_ms=$((process_total_ms + elapsed_ms))
    if [[ $process_minimum_ms -eq 0 || $elapsed_ms -lt $process_minimum_ms ]]; then
        process_minimum_ms=$elapsed_ms
    fi
    if [[ $elapsed_ms -gt $process_maximum_ms ]]; then
        process_maximum_ms=$elapsed_ms
    fi

    replay_ms=""
    if [[ "$output" =~ replay_completed=true[[:space:]]replay_ms=([0-9]+) ]]; then
        replay_ms="${BASH_REMATCH[1]}"
    fi
    runtime_line="$(printf '%s\n' "$output" | grep 'msg="runtime stopped".*component=runtime.*event=stop' | tail -n 1)"
    runtime_ok=0
    if [[ "$runtime_line" =~ status=success.*ticks_accepted=([0-9]+).*passes=([0-9]+).*signals_received=([0-9]+).*cycles_started=([0-9]+).*cycles_closed=([0-9]+).*stop_loss_exits=([0-9]+).*end_date_exits=([0-9]+).*stop_reason=end_date ]]; then
        ticks="${BASH_REMATCH[1]}"
        passes="${BASH_REMATCH[2]}"
        signals="${BASH_REMATCH[3]}"
        cycles_started="${BASH_REMATCH[4]}"
        cycles_closed="${BASH_REMATCH[5]}"
        stop_loss_exits="${BASH_REMATCH[6]}"
        end_date_exits="${BASH_REMATCH[7]}"
        if [[ $ticks -eq 7948800 && $passes -eq 794880 && $signals -eq 55 &&
              $cycles_started -eq 18 && $cycles_closed -eq 18 &&
              $stop_loss_exits -eq 17 && $end_date_exits -eq 1 ]]; then
            runtime_ok=1
        fi
    fi
    if [[ $status -ne 0 || -z "$replay_ms" || $runtime_ok -ne 1 ]]; then
        printf '%s\n' "$output"
        if [[ $status -eq 0 ]]; then
            status=1
            printf 'run=%d incomplete_replay=missing_completion_timing_or_runtime_stats\n' "$run"
        fi
        printf 'run=%d result=FAIL exit=%d elapsed_ms=%d\n' "$run" "$status" "$elapsed_ms"
        printf 'requested=%d attempted=%d passed=%d failed=1 suite_ms=%d process_total_ms=%d process_average_ms=%d process_min_ms=%d process_max_ms=%d replay_total_ms=%d replay_min_ms=%d replay_max_ms=%d log=%s\n' \
            "$runs" "$run" "$passed" "$(( $(date +%s%3N) - suite_started_ms ))" \
            "$process_total_ms" "$((process_total_ms / run))" "$process_minimum_ms" "$process_maximum_ms" \
            "$replay_total_ms" "$replay_minimum_ms" "$replay_maximum_ms" "$result_log"
        exit "$status"
    fi

    printf '%s\n' "$output" | grep -E 'msg="(signaler prepared|runtime stopped|tick reader stopped|btrunner stopped)"'

    replay_total_ms=$((replay_total_ms + replay_ms))
    if [[ $replay_minimum_ms -eq 0 || $replay_ms -lt $replay_minimum_ms ]]; then
        replay_minimum_ms=$replay_ms
    fi
    if [[ $replay_ms -gt $replay_maximum_ms ]]; then
        replay_maximum_ms=$replay_ms
    fi
    ((passed += 1))
    printf 'run=%d result=PASS exit=0 process_ms=%d replay_ms=%d ticks=%d passes=%d signals=%d cycles=%d stop_loss=%d end_date=%d\n' \
        "$run" "$elapsed_ms" "$replay_ms" "$ticks" "$passes" "$signals" "$cycles_closed" "$stop_loss_exits" "$end_date_exits"
done

printf 'requested=%d attempted=%d passed=%d failed=0 suite_ms=%d process_total_ms=%d process_average_ms=%d process_min_ms=%d process_max_ms=%d replay_total_ms=%d replay_average_ms=%d replay_min_ms=%d replay_max_ms=%d log=%s\n' \
    "$runs" "$runs" "$passed" \
    "$(( $(date +%s%3N) - suite_started_ms ))" \
    "$process_total_ms" "$((process_total_ms / runs))" "$process_minimum_ms" "$process_maximum_ms" \
    "$replay_total_ms" "$((replay_total_ms / runs))" "$replay_minimum_ms" "$replay_maximum_ms" "$result_log"
