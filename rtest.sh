#!/usr/bin/env bash

set -u -o pipefail

runs="${1:-5}"
sweep_id="${2:-6}"
bot_id="${3:-9}"
if [[ $# -gt 3 ]]; then
    echo "usage: bash rtest.sh [runs] [sweep_id] [bot_id]" >&2
    exit 2
fi
for value in "$runs" "$sweep_id" "$bot_id"; do
    if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
        echo "usage: bash rtest.sh [runs] [sweep_id] [bot_id]" >&2
        exit 2
    fi
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$repo_root/build.sh" || exit 1
suffix=""
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) suffix=".exe" ;;
esac
btrunner="$repo_root/bin/nuubot-btrunner${suffix}"
reporter="$repo_root/bin/nuubot-report${suffix}"
if [[ ! -x "$btrunner" || ! -x "$reporter" ]]; then
    echo "required binary is missing" >&2
    exit 2
fi

log_dir="$repo_root/workspace/logs"
mkdir -p "$log_dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
result_log="$log_dir/nuubot5-rtest-s${sweep_id}-b${bot_id}-${runs}-${stamp}.log"
suite_json="${result_log%.log}.json"
exec > >(tee -a "$result_log") 2>&1

suite_started_ms="$(date +%s%3N)"
bot_log="$log_dir/bot_${sweep_id}_${bot_id}.log"
attempt_records=""
suite_failed=0

for ((run = 1; run <= runs; run++)); do
    before_lines=0
    if [[ -f "$bot_log" ]]; then
        before_lines="$(wc -l < "$bot_log")"
    fi
    started_ms="$(date +%s%3N)"
    run_json="$(cd "$repo_root" && timeout 120s "$btrunner" "$sweep_id" "$bot_id")"
    status=$?
    elapsed_ms=$(( $(date +%s%3N) - started_ms ))
    output=""
    if [[ -f "$bot_log" ]]; then
        output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
    fi

    controller_line="$(printf '%s\n' "$output" | grep '] controller stopped ' | tail -n 1)"
    valid=0
    if [[ "$controller_line" =~ ticks_accepted=7948800.*runs=794880.*signal_packages_read=2208.*start_actions_skipped=724.*cycles_started=64.*cycles_rejected=0.*cycles_closed=64.*stop_loss_exits=17.*stop_reason=parent_stop ]]; then
        valid=1
    fi
    if [[ $status -eq 0 && $valid -eq 1 && -n "$run_json" ]]; then
        attempt_records+="$(printf \
            '{"run":%d,"exit":0,"btrunner_elapsed_ms":%d,"report":%s}' \
            "$run" "$elapsed_ms" "$run_json")"$'\n'
        printf '%s\n' "$output" |
            grep -E '] (signaler initialized|controller stopped|tick reader stopped|btrunner stopped)'
        printf 'run=%d result=PASS btrunner_elapsed_ms=%d\n' "$run" "$elapsed_ms"
        continue
    fi

    if [[ $status -eq 0 ]]; then
        status=1
    fi
    attempt_records+="$(printf \
        '{"run":%d,"exit":%d,"btrunner_elapsed_ms":%d,"error":"validation failed"}' \
        "$run" "$status" "$elapsed_ms")"$'\n'
    printf '%s\n' "$output"
    printf 'run=%d result=FAIL exit=%d btrunner_elapsed_ms=%d\n' \
        "$run" "$status" "$elapsed_ms"
    suite_failed=1
    break
done

suite_elapsed_ms=$(( $(date +%s%3N) - suite_started_ms ))
printf '%s' "$attempt_records" |
    "$reporter" "$runs" "$sweep_id" "$bot_id" "$suite_elapsed_ms" "$suite_json"
report_status=$?
printf 'suite_report=%s result_log=%s\n' "$suite_json" "$result_log"
if [[ $suite_failed -ne 0 || $report_status -ne 0 ]]; then
    exit 1
fi
