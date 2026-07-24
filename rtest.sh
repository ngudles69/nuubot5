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
heap_values=()
total_alloc_values=()
gc_run_values=()
gc_pause_values=()
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
    heap_mb=""
    total_alloc_mb=""
    gc_runs=""
    gc_pause_ms=""
    btrunner_line="$(printf '%s\n' "$output" | grep '] btrunner stopped ' | tail -n 1)"
    for field in $btrunner_line; do
        case "$field" in
            heap_mb=*) heap_mb="${field#heap_mb=}" ;;
            total_alloc_mb=*) total_alloc_mb="${field#total_alloc_mb=}" ;;
            gc_runs=*) gc_runs="${field#gc_runs=}" ;;
            gc_pause_ms=*) gc_pause_ms="${field#gc_pause_ms=}" ;;
        esac
    done
    runtime_line="$(printf '%s\n' "$output" | grep '] runtime stopped ' | tail -n 1)"
    runtime_ok=0
    if [[ "$runtime_line" =~ ticks_accepted=([0-9]+).*runs=([0-9]+).*signals_received=([0-9]+).*cycles_started=([0-9]+).*cycles_closed=([0-9]+).*stop_loss_exits=([0-9]+).*stop_reason=parent_stop ]]; then
        ticks="${BASH_REMATCH[1]}"
        runtime_runs="${BASH_REMATCH[2]}"
        signals="${BASH_REMATCH[3]}"
        cycles_started="${BASH_REMATCH[4]}"
        cycles_closed="${BASH_REMATCH[5]}"
        stop_loss_exits="${BASH_REMATCH[6]}"
        if [[ $ticks -eq 7948800 && $runtime_runs -eq 794880 && $signals -eq 55 &&
              $cycles_started -eq 18 && $cycles_closed -eq 18 &&
              $stop_loss_exits -eq 17 ]]; then
            runtime_ok=1
        fi
    fi
    if [[ $status -ne 0 || -z "$replay_ms" || -z "$heap_mb" || -z "$total_alloc_mb" ||
          -z "$gc_runs" || -z "$gc_pause_ms" || $runtime_ok -ne 1 ]]; then
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

    printf '%s\n' "$output" | grep -E '] (signaler prepared|runtime stopped|tick reader stopped|btrunner stopped)'

    replay_total_ms=$((replay_total_ms + replay_ms))
    if [[ $replay_minimum_ms -eq 0 || $replay_ms -lt $replay_minimum_ms ]]; then
        replay_minimum_ms=$replay_ms
    fi
    if [[ $replay_ms -gt $replay_maximum_ms ]]; then
        replay_maximum_ms=$replay_ms
    fi
    heap_values+=("$heap_mb")
    total_alloc_values+=("$total_alloc_mb")
    gc_run_values+=("$gc_runs")
    gc_pause_values+=("$gc_pause_ms")
    ((passed += 1))
    printf 'run=%d result=PASS exit=0 process_ms=%d replay_ms=%d heap_mb=%s total_alloc_mb=%s gc_runs=%s gc_pause_ms=%s ticks=%d runs=%d signals=%d cycles=%d stop_loss=%d\n' \
        "$run" "$elapsed_ms" "$replay_ms" "$heap_mb" "$total_alloc_mb" "$gc_runs" "$gc_pause_ms" \
        "$ticks" "$runtime_runs" "$signals" "$cycles_closed" "$stop_loss_exits"
done

printf 'requested=%d attempted=%d passed=%d failed=0 suite_ms=%d process_total_ms=%d process_average_ms=%d process_min_ms=%d process_max_ms=%d replay_total_ms=%d replay_average_ms=%d replay_min_ms=%d replay_max_ms=%d log=%s\n' \
    "$runs" "$runs" "$passed" \
    "$(( $(date +%s%3N) - suite_started_ms ))" \
    "$process_total_ms" "$((process_total_ms / runs))" "$process_minimum_ms" "$process_maximum_ms" \
    "$replay_total_ms" "$((replay_total_ms / runs))" "$replay_minimum_ms" "$replay_maximum_ms" "$result_log"

metric_stats() {
    printf '%s\n' "$@" | awk '
        NR == 1 { minimum = maximum = $1 }
        { total += $1; if ($1 < minimum) minimum = $1; if ($1 > maximum) maximum = $1 }
        END { printf "%.3f %.3f %.3f", total / NR, minimum, maximum }
    '
}

read -r heap_average heap_minimum heap_maximum <<< "$(metric_stats "${heap_values[@]}")"
read -r total_alloc_average total_alloc_minimum total_alloc_maximum <<< "$(metric_stats "${total_alloc_values[@]}")"
read -r gc_run_average gc_run_minimum gc_run_maximum <<< "$(metric_stats "${gc_run_values[@]}")"
read -r gc_pause_average gc_pause_minimum gc_pause_maximum <<< "$(metric_stats "${gc_pause_values[@]}")"

printf 'heap_mb_average=%s heap_mb_min=%s heap_mb_max=%s total_alloc_mb_average=%s total_alloc_mb_min=%s total_alloc_mb_max=%s gc_runs_average=%s gc_runs_min=%s gc_runs_max=%s gc_pause_ms_average=%s gc_pause_ms_min=%s gc_pause_ms_max=%s\n' \
    "$heap_average" "$heap_minimum" "$heap_maximum" \
    "$total_alloc_average" "$total_alloc_minimum" "$total_alloc_maximum" \
    "$gc_run_average" "$gc_run_minimum" "$gc_run_maximum" \
    "$gc_pause_average" "$gc_pause_minimum" "$gc_pause_maximum"
