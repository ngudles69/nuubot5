#!/usr/bin/env bash

set -u -o pipefail

runs="${1:-1}"
sweep_id="${2:-10}"
bot_id="${3:-14}"
for value in "$runs" "$sweep_id" "$bot_id"; do
    if [[ ! "$value" =~ ^[1-9][0-9]*$ ]]; then
        echo "usage: bash grtest.sh [runs] [sweep_id] [bot_id]" >&2
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
if ! command -v sqlite3 >/dev/null 2>&1; then
    echo "sqlite3 is required" >&2
    exit 2
fi

log_dir="$repo_root/workspace/logs"
mkdir -p "$log_dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
result_log="$log_dir/nuubot5-grtest-s${sweep_id}-b${bot_id}-${runs}-${stamp}.log"
suite_json="${result_log%.log}.json"
result_db="$repo_root/workspace/db/sweeps/sweep_${sweep_id}/bot_${bot_id}.db"
source_db="$repo_root/workspace/db/nuubot.db"
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) source_db="$(cygpath -m "$source_db")" ;;
esac
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
    run_json="$(cd "$repo_root" && timeout 180s "$btrunner" "$sweep_id" "$bot_id")"
    status=$?
    elapsed_ms=$(( $(date +%s%3N) - started_ms ))
    output=""
    if [[ -f "$bot_log" ]]; then
        output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
    fi

    integrity=""
    foreign_keys="missing"
    result_spec=""
    completed=""
    ticks=""
    controller_runs=""
    cycles=""
    trades=""
    orders=""
    fills=""
    cancellations=""
    stop_order_count=""
    retries=""
    telemetry_rows=""
    report_samples=""
    result_config_match=""
    result_cycles=""
    result_executors=""
    result_levels=""
    result_boundaries=""
    result_initial_levels=""
    active_orders=""
    nonflat_accounts=""
    false_equity_samples=""
    declining_max_drawdown=""
    close_orders=""
    stop_orders=""
    if [[ -f "$result_db" ]]; then
        integrity="$(sqlite3 "$result_db" 'PRAGMA integrity_check;')"
        foreign_keys="$(sqlite3 "$result_db" 'PRAGMA foreign_key_check;')"
        IFS='|' read -r result_spec completed ticks controller_runs cycles trades orders fills cancellations stop_order_count retries telemetry_rows report_samples <<< "$(
            sqlite3 -separator '|' "$result_db" "
                SELECT b.bot_spec_id,b.completed,b.ticks_served,b.runs_triggered,
                       r.cycles_closed,r.trades,r.orders,r.fills,r.cancellations,
                       r.stop_orders,r.retries,
                       (SELECT COUNT(*) FROM telemetry_sample),r.telemetry_samples
                FROM backtest_result b CROSS JOIN run_report r;
            "
        )"
        result_cycles="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM botcycle_result;')"
        result_executors="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM executor_result;')"
        result_levels="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM grid_level_result;')"
        result_boundaries="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM grid_level_result WHERE boundary=1;')"
        result_initial_levels="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM grid_level_result WHERE initial_submission_completed=1;')"
        active_orders="$(sqlite3 "$result_db" "SELECT COUNT(*) FROM account_order WHERE status IN ('created','submitted','open','partially_filled');")"
        false_equity_samples="$(sqlite3 "$result_db" "SELECT COUNT(*) FROM telemetry_sample WHERE active_cycle>0 AND bot_equity='0';")"
        declining_max_drawdown="$(
            sqlite3 "$result_db" "
                SELECT COUNT(*) FROM telemetry_sample AS later
                JOIN telemetry_sample AS earlier
                  ON earlier.sequence = later.sequence - 1
                WHERE CAST(later.max_drawdown AS REAL)
                    < CAST(earlier.max_drawdown AS REAL);
            "
        )"
        close_orders="$(sqlite3 "$result_db" "SELECT COUNT(*) FROM account_order WHERE order_role='close';")"
        stop_orders="$(sqlite3 "$result_db" "SELECT COUNT(*) FROM account_order WHERE order_role='stop';")"
        nonflat_accounts="$(
            sqlite3 "$result_db" "
                SELECT COUNT(*) FROM account_ledger
                WHERE CAST(json_extract(account_state_json, '$.PositionSize') AS REAL) != 0;
            "
        )"
        result_config_match="$(
            sqlite3 "$result_db" "
                ATTACH DATABASE '$source_db' AS source;
                SELECT CASE WHEN result.config_toml = source.config_toml
                    AND result.config_hash = source.config_hash THEN 1 ELSE 0 END
                FROM backtest_result AS result
                JOIN source.bot AS source
                  ON source.sweep_id = result.sweep_id
                 AND source.bot_id = result.bot_id;
            "
        )"
    fi

    valid=1
    if [[ $status -ne 0 || -z "$run_json" ||
          "$integrity" != "ok" || -n "$foreign_keys" ||
          "$result_spec" != "macross_grid_bot" || "$completed" != "1" ||
          "$ticks" != "7948800" || "$controller_runs" != "794880" ||
          -z "$cycles" || "$cycles" -le 0 ||
          "$trades" -lt "$((cycles * 28))" ||
          "$orders" != "$((trades * 2 + stop_order_count))" ||
          "$telemetry_rows" != "$((controller_runs + 1))" ||
          "$report_samples" != "$telemetry_rows" ||
          "$result_cycles" != "$cycles" || "$result_executors" != "$cycles" ||
          "$result_levels" != "$((cycles * 30))" ||
          "$result_boundaries" != "$((cycles * 2))" ||
          "$result_initial_levels" != "$((cycles * 28))" ||
          "$active_orders" != "0" || "$nonflat_accounts" != "0" ||
          "$false_equity_samples" != "0" || "$declining_max_drawdown" != "0" ||
          "$close_orders" != "0" || "$stop_orders" != "$stop_order_count" ||
          "$result_config_match" != "1" || -f "${result_db}.partial" ]]; then
        valid=0
    fi
    if [[ $valid -eq 1 ]]; then
        attempt_records+="$(printf \
            '{"run":%d,"exit":0,"btrunner_elapsed_ms":%d,"report":%s}' \
            "$run" "$elapsed_ms" "$run_json")"$'\n'
        printf '%s\n' "$output" | grep -E '] (controller stopped|btrunner stopped)'
        printf 'run=%d result=PASS btrunner_elapsed_ms=%d result_db=%s\n' \
            "$run" "$elapsed_ms" "$result_db"
        continue
    fi

    if [[ $status -eq 0 ]]; then
        status=1
    fi
    attempt_records+="$(printf \
        '{"run":%d,"exit":%d,"btrunner_elapsed_ms":%d,"error":"validation failed"}' \
        "$run" "$status" "$elapsed_ms")"$'\n'
    printf '%s\n' "$output"
    printf 'run=%d result=FAIL exit=%d cycles=%s trades=%s orders=%s fills=%s cancellations=%s stop_orders=%s retries=%s\n' \
        "$run" "$status" "${cycles:-missing}" "${trades:-missing}" \
        "${orders:-missing}" "${fills:-missing}" "${cancellations:-missing}" \
        "${stop_order_count:-missing}" "${retries:-missing}"
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
