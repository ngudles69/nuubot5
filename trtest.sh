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
source_db="$repo_root/workspace/db/nuubot.db"
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) source_db="$(cygpath -m "$source_db")" ;;
esac

for ((run = 1; run <= runs; run++)); do
    before_lines=0
    if [[ -f "$bot_log" ]]; then
        before_lines="$(wc -l < "$bot_log")"
    fi
    started_ms="$(date +%s%3N)"
    output="$(
        cd "$repo_root" &&
        timeout 120s "$binary" "$sweep_id" "$bot_id" 2>&1
    )"
    status=$?
    process_ms=$(( $(date +%s%3N) - started_ms ))
    if [[ -f "$bot_log" ]]; then
        output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
    fi

    btrunner_line="$(printf '%s\n' "$output" | grep '] btrunner stopped ' | tail -n 1)"
    controller_line="$(printf '%s\n' "$output" | grep '] controller stopped ' | tail -n 1)"
    replay_ms="$(printf '%s\n' "$btrunner_line" | sed -n 's/.*replay_ms=\([0-9][0-9]*\).*/\1/p')"
    ticks="$(printf '%s\n' "$controller_line" | sed -n 's/.*ticks_accepted=\([0-9][0-9]*\).*/\1/p')"
    controller_runs="$(printf '%s\n' "$controller_line" | sed -n 's/.*runs=\([0-9][0-9]*\).*/\1/p')"
    cycles="$(printf '%s\n' "$controller_line" | sed -n 's/.*cycles_closed=\([0-9][0-9]*\).*/\1/p')"
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
    forced_exits="$(printf '%s\n' "$trade_lines" | awk '
        /stop_reason=/ && $0 !~ /stop_reason=completed$/ { total++ }
        END { print total + 0 }
    ')"
    account_lines="$(printf '%s\n' "$output" | grep '] account stopped ' || true)"
    orders="$(printf '%s\n' "$account_lines" | awk '
        { for (i = 1; i <= NF; i++) if ($i ~ /^orders=/) {
            split($i, value, "="); total += value[2]
        }}
        END { print total + 0 }
    ')"
    integrity=""
    foreign_keys=""
    db_trades=""
    db_orders=""
    db_fills=""
    db_states=""
    result_spec=""
    result_config_match=""
    equity_carry=""
    result_equity_match=""
    result_completed=""
    result_ticks=""
    result_runs=""
    result_cycles=""
    result_executors=""
    result_pnl=""
    result_equity=""
    if command -v sqlite3 >/dev/null 2>&1 && [[ -f "$result_db" ]]; then
        integrity="$(sqlite3 "$result_db" 'PRAGMA integrity_check;')"
        foreign_keys="$(sqlite3 "$result_db" 'PRAGMA foreign_key_check;')"
        db_trades="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM account_trade;')"
        db_orders="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM account_order;')"
        db_fills="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM account_fill;')"
        db_states="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM simulator_state;')"
        IFS='|' read -r result_spec result_completed result_ticks result_runs result_pnl result_equity <<< "$(
            sqlite3 -separator '|' "$result_db" \
                'SELECT bot_spec_id,completed,ticks_served,runs_triggered,net_pnl,bot_equity FROM backtest_result;'
        )"
        result_cycles="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM botcycle_result;')"
        result_executors="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM executor_result;')"
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
        equity_carry="$(
            sqlite3 "$result_db" "
                SELECT CASE WHEN
                    (SELECT json_extract(payload_json, '$.Equity')
                     FROM simulator_state
                     WHERE json_extract(payload_json, '$.CycleNumber') = 2)
                    =
                    (SELECT json_extract(account_state_json, '$.AccountValue')
                     FROM account_ledger
                     WHERE cycle_no = 1)
                THEN 1 ELSE 0 END;
            "
        )"
        result_equity_match="$(
            sqlite3 "$result_db" "
                SELECT CASE WHEN result.bot_equity =
                    (SELECT json_extract(account_state_json, '$.AccountValue')
                     FROM account_ledger
                     ORDER BY cycle_no DESC LIMIT 1)
                THEN 1 ELSE 0 END
                FROM backtest_result AS result;
            "
        )"
    fi

    if [[ $status -ne 0 || -z "$replay_ms" || "$ticks" != "7948800" ||
          "$controller_runs" != "794880" || -z "$cycles" ||
          "$trade_cycles" != "$cycles" || "$trades" != "$cycles" ||
          "$fills" != "$((cycles * 2))" ||
          "$orders" != "$((cycles * 3 + forced_exits))" ||
          "$integrity" != "ok" || -n "$foreign_keys" ||
          "$db_trades" != "$trades" || "$db_orders" != "$orders" ||
          "$db_fills" != "$fills" || "$db_states" != "$cycles" ||
          "$result_spec" != "macross_trade_bot" || "$result_completed" != "1" ||
          "$result_ticks" != "$ticks" || "$result_runs" != "$controller_runs" ||
          "$result_cycles" != "$cycles" || "$result_executors" != "$cycles" ||
          "$result_config_match" != "1" || "$equity_carry" != "1" ||
          "$result_equity_match" != "1" ||
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
    printf '%s\n' "$controller_line"
    printf '%s\n' "$btrunner_line"
    printf 'run=%d result=PASS process_ms=%d replay_ms=%d ticks=%s runs=%s cycles=%s trades=%s orders=%s fills=%s result_db=%s\n' \
        "$run" "$process_ms" "$replay_ms" "$ticks" "$controller_runs" "$cycles" \
        "$trades" "$orders" "$fills" "$result_db"
    printf 'run=%d result_summary bot_spec=%s net_pnl=%s bot_equity=%s forced_exits=%s integrity=%s\n' \
        "$run" "$result_spec" "$result_pnl" "$result_equity" "$forced_exits" "$integrity"
done

printf 'requested=%d attempted=%d passed=%d failed=0 suite_ms=%d process_total_ms=%d process_average_ms=%d replay_total_ms=%d replay_average_ms=%d log=%s\n' \
    "$runs" "$runs" "$passed" "$(( $(date +%s%3N) - suite_started_ms ))" \
    "$process_total_ms" "$((process_total_ms / runs))" \
    "$replay_total_ms" "$((replay_total_ms / runs))" "$result_log"
