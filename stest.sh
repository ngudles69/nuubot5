#!/usr/bin/env bash

set -u -o pipefail

usage() {
    echo "usage: ./stest.sh (-sweep ID | -bot ID) [-runs N] [-pp]" >&2
    exit 2
}

validation_check() {
    local result_db="$1"
    local check="$2"
    local mode="$3"
    local expected="$4"
    local actual="$5"
    local passed=0
    case "$mode" in
        equal) [[ "$actual" == "$expected" ]] && passed=1 ;;
        nonempty) [[ -n "$actual" ]] && passed=1 ;;
        empty) [[ -z "$actual" ]] && passed=1 ;;
        greater) [[ -n "$actual" && "$actual" -gt "$expected" ]] && passed=1 ;;
        at_least) [[ -n "$actual" && "$actual" -ge "$expected" ]] && passed=1 ;;
        absent) [[ ! -e "$actual" ]] && passed=1 ;;
        true) [[ "$actual" == "1" ]] && passed=1 ;;
        *) printf 'unknown validation mode: %s\n' "$mode" >&2; return 2 ;;
    esac
    if [[ $passed -eq 1 ]]; then
        return 0
    fi
    printf 'validation failed check=%s expected=%s actual=%s result_db=%s\n' \
        "$check" "$expected" "${actual:-missing}" "$result_db" >&2
    return 1
}

validate_observer() {
    local output="$1"
    local result_db="$2"
    local completed ticks_expected ticks_served runs_expected runs_triggered
    local telemetry_rows report_samples controller_line
    local failed=0
    IFS='|' read -r \
        completed ticks_expected ticks_served runs_expected runs_triggered \
        telemetry_rows report_samples <<< "$(
            sqlite3 -separator '|' "$result_db" "
                SELECT b.completed,b.ticks_expected,b.ticks_served,
                       b.runs_expected,b.runs_triggered,
                       (SELECT COUNT(*) FROM telemetry_sample),r.telemetry_samples
                FROM backtest_result b CROSS JOIN run_report r;
            "
        )"
    controller_line="$(printf '%s\n' "$output" | grep '] controller stopped ' | tail -n 1)"

    validation_check "$result_db" completed equal 1 "$completed" || failed=1
    validation_check "$result_db" ticks_expected equal "$ticks_expected" "$ticks_served" || failed=1
    validation_check "$result_db" runs_expected equal "$runs_expected" "$runs_triggered" || failed=1
    validation_check "$result_db" telemetry_rows greater 0 "$telemetry_rows" || failed=1
    validation_check "$result_db" report_samples equal "$telemetry_rows" "$report_samples" || failed=1
    local semantic_result=0
    if [[ "$controller_line" =~ ticks_accepted=7948800.*signal_packages_read=2208.*start_actions_skipped=724.*cycles_started=63.*cycles_rejected=0.*cycles_closed=63.*stop_loss_exits=16.*stop_reason=parent_stop ]]; then
        semantic_result=1
    fi
    validation_check "$result_db" observer_result true 'accepted semantic result' "$semantic_result" || failed=1
    return "$failed"
}

validate_trade() {
    local result_db="$1"
    local source_db_sql="$2"
    local integrity foreign_keys
    local completed ticks_expected ticks runs_expected controller_runs cycles trades orders fills
    local stop_order_count telemetry_rows report_samples
    local result_cycles result_executors db_states result_config_match
    local equity_carry result_equity_match false_equity_samples
    local declining_max_drawdown close_orders stop_orders domain_persisted expected_states

    integrity="$(sqlite3 "$result_db" 'PRAGMA integrity_check;')"
    foreign_keys="$(sqlite3 "$result_db" 'PRAGMA foreign_key_check;')"
    IFS='|' read -r \
        completed ticks_expected ticks runs_expected controller_runs cycles trades orders fills \
        stop_order_count telemetry_rows report_samples <<< "$(
            sqlite3 -separator '|' "$result_db" "
                SELECT b.completed,b.ticks_expected,b.ticks_served,
                       b.runs_expected,b.runs_triggered,
                       r.cycles_closed,r.trades,r.orders,r.fills,r.stop_orders,
                       (SELECT COUNT(*) FROM telemetry_sample),r.telemetry_samples
                FROM backtest_result b CROSS JOIN run_report r;
            "
        )"
    result_cycles="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM botcycle_result;')"
    result_executors="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM executor_result;')"
    domain_persisted="$(
        sqlite3 "$result_db" \
            "SELECT CASE WHEN instr(config_toml, 'persist_mode = \"max\"') > 0 THEN 1 ELSE 0 END FROM backtest_result;"
    )"
    db_states=0
    expected_states=0
    if [[ "$domain_persisted" == "1" ]]; then
        db_states="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM simulator_state;')"
        expected_states="$cycles"
    fi
    false_equity_samples="$(
        sqlite3 "$result_db" "
            SELECT COUNT(*) FROM telemetry_sample
            WHERE active_cycle>0 AND bot_equity='0';
        "
    )"
    declining_max_drawdown="$(
        sqlite3 "$result_db" "
            SELECT COUNT(*) FROM telemetry_sample later
            JOIN telemetry_sample earlier
              ON earlier.sequence=later.sequence-1
            WHERE CAST(later.max_drawdown AS REAL)
                < CAST(earlier.max_drawdown AS REAL);
        "
    )"
    close_orders=0
    stop_orders="$stop_order_count"
    equity_carry=1
    result_equity_match=1
    result_config_match="$(
        sqlite3 "$result_db" "
            ATTACH DATABASE '$source_db_sql' AS source;
            SELECT CASE WHEN result.config_toml=source.config_toml
                AND result.config_hash=source.config_hash THEN 1 ELSE 0 END
            FROM backtest_result result
            JOIN source.bot source ON source.bot_id=result.bot_id
            WHERE source.sweep_id=result.sweep_id;
        "
    )"
    if [[ "$domain_persisted" == "1" ]]; then
        close_orders="$(
        sqlite3 "$result_db" \
            "SELECT COUNT(*) FROM account_order WHERE order_role='close';"
    )"
    stop_orders="$(
        sqlite3 "$result_db" \
            "SELECT COUNT(*) FROM account_order WHERE order_role='stop';"
    )"
    equity_carry="$(
        sqlite3 "$result_db" "
            SELECT CASE WHEN
                (SELECT json_extract(payload_json,'$.Equity')
                 FROM simulator_state
                 WHERE json_extract(payload_json,'$.CycleNumber')=2)
                =
                (SELECT json_extract(account_state_json,'$.AccountValue')
                 FROM account_ledger WHERE cycle_no=1)
            THEN 1 ELSE 0 END;
        "
    )"
    result_equity_match="$(
        sqlite3 "$result_db" "
            SELECT CASE WHEN result.bot_equity=
                (SELECT json_extract(account_state_json,'$.AccountValue')
                 FROM account_ledger ORDER BY cycle_no DESC LIMIT 1)
            THEN 1 ELSE 0 END
            FROM backtest_result result;
        "
    )"
    fi

    local failed=0
    validation_check "$result_db" integrity equal ok "$integrity" || failed=1
    validation_check "$result_db" foreign_keys empty empty "$foreign_keys" || failed=1
    validation_check "$result_db" completed equal 1 "$completed" || failed=1
    validation_check "$result_db" ticks_expected equal "$ticks_expected" "$ticks" || failed=1
    validation_check "$result_db" runs_expected equal "$runs_expected" "$controller_runs" || failed=1
    validation_check "$result_db" cycles greater 0 "$cycles" || failed=1
    validation_check "$result_db" trades equal "$cycles" "$trades" || failed=1
    validation_check "$result_db" fills equal "$((cycles * 2))" "$fills" || failed=1
    validation_check "$result_db" orders equal "$((cycles * 3 + stop_order_count))" "$orders" || failed=1
    validation_check "$result_db" telemetry_rows greater 0 "$telemetry_rows" || failed=1
    validation_check "$result_db" report_samples equal "$telemetry_rows" "$report_samples" || failed=1
    validation_check "$result_db" result_cycles equal "$cycles" "$result_cycles" || failed=1
    validation_check "$result_db" result_executors equal "$cycles" "$result_executors" || failed=1
    validation_check "$result_db" simulator_states equal "$expected_states" "$db_states" || failed=1
    validation_check "$result_db" result_config_match equal 1 "$result_config_match" || failed=1
    validation_check "$result_db" equity_carry equal 1 "$equity_carry" || failed=1
    validation_check "$result_db" result_equity_match equal 1 "$result_equity_match" || failed=1
    validation_check "$result_db" false_equity_samples equal 0 "$false_equity_samples" || failed=1
    validation_check "$result_db" declining_max_drawdown equal 0 "$declining_max_drawdown" || failed=1
    validation_check "$result_db" close_orders equal 0 "$close_orders" || failed=1
    validation_check "$result_db" stop_orders equal "$stop_order_count" "$stop_orders" || failed=1
    validation_check "$result_db" partial_result absent absent "${result_db}.partial" || failed=1
    return "$failed"
}

write_result_log() {
    local path="$1"
    tee -a "$path"
}

validate_grid() {
    local result_db="$1"
    local source_db_sql="$2"
    local integrity foreign_keys
    local completed ticks_expected ticks runs_expected controller_runs cycles trades orders fills
    local cancellations stop_order_count retries telemetry_rows report_samples
    local result_cycles result_executors result_levels result_boundaries
    local result_initial_levels active_orders nonflat_accounts
    local false_equity_samples declining_max_drawdown close_orders stop_orders
    local result_config_match domain_persisted

    integrity="$(sqlite3 "$result_db" 'PRAGMA integrity_check;')"
    foreign_keys="$(sqlite3 "$result_db" 'PRAGMA foreign_key_check;')"
    IFS='|' read -r \
        completed ticks_expected ticks runs_expected controller_runs cycles trades orders fills cancellations \
        stop_order_count retries telemetry_rows report_samples <<< "$(
            sqlite3 -separator '|' "$result_db" "
                SELECT b.completed,b.ticks_expected,b.ticks_served,
                       b.runs_expected,b.runs_triggered,
                       r.cycles_closed,r.trades,r.orders,r.fills,r.cancellations,
                       r.stop_orders,r.retries,
                       (SELECT COUNT(*) FROM telemetry_sample),r.telemetry_samples
                FROM backtest_result b CROSS JOIN run_report r;
            "
        )"
    result_cycles="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM botcycle_result;')"
    result_executors="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM executor_result;')"
    domain_persisted="$(
        sqlite3 "$result_db" \
            "SELECT CASE WHEN instr(config_toml, 'persist_mode = \"max\"') > 0 THEN 1 ELSE 0 END FROM backtest_result;"
    )"
    result_levels="$(sqlite3 "$result_db" 'SELECT COUNT(*) FROM grid_level_result;')"
    result_boundaries="$(
        sqlite3 "$result_db" \
            'SELECT COUNT(*) FROM grid_level_result WHERE boundary=1;'
    )"
    result_initial_levels="$(
        sqlite3 "$result_db" \
            'SELECT COUNT(*) FROM grid_level_result WHERE initial_submission_completed=1;'
    )"
    active_orders=0
    nonflat_accounts=0
    if [[ "$domain_persisted" == "1" ]]; then
        active_orders="$(
            sqlite3 "$result_db" "
                SELECT COUNT(*) FROM account_order
                WHERE status IN ('created','submitted','open','partially_filled');
            "
        )"
        nonflat_accounts="$(
            sqlite3 "$result_db" "
                SELECT COUNT(*) FROM account_ledger
                WHERE CAST(json_extract(account_state_json,'$.PositionSize') AS REAL)!=0;
            "
        )"
    fi
    false_equity_samples="$(
        sqlite3 "$result_db" "
            SELECT COUNT(*) FROM telemetry_sample
            WHERE active_cycle>0 AND bot_equity='0';
        "
    )"
    declining_max_drawdown="$(
        sqlite3 "$result_db" "
            SELECT COUNT(*) FROM telemetry_sample later
            JOIN telemetry_sample earlier
              ON earlier.sequence=later.sequence-1
            WHERE CAST(later.max_drawdown AS REAL)
                < CAST(earlier.max_drawdown AS REAL);
        "
    )"
    close_orders=0
    stop_orders="$stop_order_count"
    if [[ "$domain_persisted" == "1" ]]; then
        close_orders="$(
            sqlite3 "$result_db" \
                "SELECT COUNT(*) FROM account_order WHERE order_role='close';"
        )"
        stop_orders="$(
            sqlite3 "$result_db" \
                "SELECT COUNT(*) FROM account_order WHERE order_role='stop';"
        )"
    fi
    result_config_match="$(
        sqlite3 "$result_db" "
            ATTACH DATABASE '$source_db_sql' AS source;
            SELECT CASE WHEN result.config_toml=source.config_toml
                AND result.config_hash=source.config_hash THEN 1 ELSE 0 END
            FROM backtest_result result
            JOIN source.bot source ON source.bot_id=result.bot_id
            WHERE source.sweep_id=result.sweep_id;
        "
    )"

    local failed=0
    validation_check "$result_db" integrity equal ok "$integrity" || failed=1
    validation_check "$result_db" foreign_keys empty empty "$foreign_keys" || failed=1
    validation_check "$result_db" completed equal 1 "$completed" || failed=1
    validation_check "$result_db" ticks_expected equal "$ticks_expected" "$ticks" || failed=1
    validation_check "$result_db" runs_expected equal "$runs_expected" "$controller_runs" || failed=1
    validation_check "$result_db" cycles greater 0 "$cycles" || failed=1
    validation_check "$result_db" trades at_least "$((cycles * 28))" "$trades" || failed=1
    validation_check "$result_db" orders equal "$((trades * 2 + stop_order_count))" "$orders" || failed=1
    validation_check "$result_db" telemetry_rows greater 0 "$telemetry_rows" || failed=1
    validation_check "$result_db" report_samples equal "$telemetry_rows" "$report_samples" || failed=1
    validation_check "$result_db" result_cycles equal "$cycles" "$result_cycles" || failed=1
    validation_check "$result_db" result_executors equal "$cycles" "$result_executors" || failed=1
    validation_check "$result_db" result_levels equal "$((cycles * 30))" "$result_levels" || failed=1
    validation_check "$result_db" result_boundaries equal "$((cycles * 2))" "$result_boundaries" || failed=1
    validation_check "$result_db" result_initial_levels equal "$((cycles * 28))" "$result_initial_levels" || failed=1
    validation_check "$result_db" active_orders equal 0 "$active_orders" || failed=1
    validation_check "$result_db" nonflat_accounts equal 0 "$nonflat_accounts" || failed=1
    validation_check "$result_db" false_equity_samples equal 0 "$false_equity_samples" || failed=1
    validation_check "$result_db" declining_max_drawdown equal 0 "$declining_max_drawdown" || failed=1
    validation_check "$result_db" close_orders equal 0 "$close_orders" || failed=1
    validation_check "$result_db" stop_orders equal "$stop_order_count" "$stop_orders" || failed=1
    validation_check "$result_db" result_config_match equal 1 "$result_config_match" || failed=1
    validation_check "$result_db" partial_result absent absent "${result_db}.partial" || failed=1
    return "$failed"
}

if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
    return 0
fi

selector=""
selector_id=""
runs=1
runs_set=0
profile=0
profile_set=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        -sweep|-bot)
            [[ -z "$selector" && $# -ge 2 ]] || usage
            selector="$1"
            selector_id="$2"
            shift 2
            ;;
        -runs)
            [[ $runs_set -eq 0 && $# -ge 2 ]] || usage
            runs="$2"
            runs_set=1
            shift 2
            ;;
        -pp)
            [[ $profile_set -eq 0 ]] || usage
            profile=1
            profile_set=1
            shift
            ;;
        *) usage ;;
    esac
done
[[ -n "$selector" ]] || usage
for value in "$selector_id" "$runs"; do
    [[ "$value" =~ ^[1-9][0-9]*$ ]] || usage
done
[[ $profile -eq 0 || "$runs" == "1" ]] || usage

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source_db="$repo_root/workspace/db/nuubot.db"
command -v sqlite3 >/dev/null 2>&1 || {
    echo "sqlite3 is required" >&2
    exit 2
}
[[ -f "$source_db" ]] || {
    echo "source database is missing: $source_db" >&2
    exit 2
}

bot_no_exists="$(
    sqlite3 "$source_db" \
        "SELECT COUNT(*) FROM pragma_table_info('bot') WHERE name='bot_no';"
)"
case "$bot_no_exists" in
    1) order_column="bot_no" ;;
    0) order_column="bot_id" ;;
    *) echo "unable to inspect bot ordering schema" >&2; exit 2 ;;
esac

if [[ "$selector" == "-sweep" ]]; then
    if ! selected_bots="$(
        sqlite3 -separator '|' "$source_db" "
            SELECT sweep_id,bot_id,$order_column,bot_spec_id
            FROM bot
            WHERE sweep_id=$selector_id
            ORDER BY $order_column,bot_id;
        "
    )"; then
        echo "unable to select Sweep Bots" >&2
        exit 2
    fi
else
    if ! selected_bots="$(
        sqlite3 -separator '|' "$source_db" "
            SELECT sweep_id,bot_id,$order_column,bot_spec_id
            FROM bot
            WHERE bot_id=$selector_id;
        "
    )"; then
        echo "unable to select Bot" >&2
        exit 2
    fi
fi
[[ -n "$selected_bots" ]] || {
    echo "selector matched no Bots" >&2
    exit 2
}
if [[ "$selector" == "-bot" && "$selected_bots" == *$'\n'* ]]; then
    echo "BotID is not globally unique" >&2
    exit 2
fi
while IFS='|' read -r sweep_id bot_id order_value stored_spec; do
    case "$stored_spec" in
        macross_observer_bot|macross_trade_bot|macross_grid_bot) ;;
        *)
            echo "unsupported BotSpec for BotID $bot_id: ${stored_spec:-missing}" >&2
            exit 2
            ;;
    esac
done <<< "$selected_bots"

bash "$repo_root/build.sh" || exit 1
suffix=""
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) suffix=".exe" ;;
esac
backtest="$repo_root/bin/nuubot-backtest${suffix}"
stest_report="$repo_root/bin/nuubot-stest-report${suffix}"
[[ -x "$backtest" && -x "$stest_report" ]] || {
    echo "required binary is missing" >&2
    exit 2
}

source_db_attach="$source_db"
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) source_db_attach="$(cygpath -m "$source_db")" ;;
esac
source_db_sql="$(printf '%s' "$source_db_attach" | sed "s/'/''/g")"
log_dir="$repo_root/workspace/logs"
mkdir -p "$log_dir" || exit 1
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
suite_failed=0

while IFS='|' read -r sweep_id bot_id order_value stored_spec; do
    result_log="$log_dir/nuubot5-stest-s${sweep_id}-b${bot_id}-${runs}-${stamp}.log"
    suite_json="${result_log%.log}.json"
    result_db="$repo_root/workspace/db/bots/bot_${bot_id}.db"
    bot_log="$log_dir/bot_${sweep_id}_${bot_id}.log"
    profile_dir="$repo_root/workspace/perf/profiles/stest-s${sweep_id}-b${bot_id}-${stamp}"
    attempt_records=""
    bot_failed=0
    suite_started_ms="$(date +%s%3N)"

    if ! {
        printf '\nSelected Bot\n'
        printf '%-14s %s\n' "SweepID" "$sweep_id"
        printf '%-14s %s\n' "BotID" "$bot_id"
        printf '%-14s %s\n' "$order_column" "$order_value"
        printf '%-14s %s\n' "BotSpecID" "$stored_spec"
        printf '%-14s %s\n' "Runs" "$runs"
        printf '%-14s %s\n\n' "Profiling" "$profile"
    } | write_result_log "$result_log"; then
        exit 1
    fi

    if [[ $profile -eq 1 ]] && ! mkdir -p "$profile_dir"; then
        if ! printf 'result=FAIL error=unable_to_create_profile_directory\n' |
            write_result_log "$result_log"; then
            exit 1
        fi
        suite_failed=1
        continue
    fi

    for ((run = 1; run <= runs; run++)); do
        before_lines=0
        if [[ -f "$bot_log" ]]; then
            before_lines="$(wc -l < "$bot_log")"
        fi
        runner_args=("$sweep_id" "$bot_id")
        if [[ $profile -eq 1 ]]; then
            runner_args+=("-pp" "$profile_dir/run-$(printf '%03d' "$run")")
        fi
        timeout_seconds=120
        if [[ "$stored_spec" == "macross_grid_bot" ]]; then
            timeout_seconds=180
        fi

        started_ms="$(date +%s%3N)"
        run_json="$(
            cd "$repo_root" &&
                timeout "${timeout_seconds}s" "$backtest" "${runner_args[@]}"
        )"
        status=$?
        elapsed_ms=$(( $(date +%s%3N) - started_ms ))
        output=""
        if [[ -f "$bot_log" ]]; then
            output="$(tail -n "+$((before_lines + 1))" "$bot_log")"
        fi

        valid=1
        reported_spec=""
        if [[ $status -ne 0 || -z "$run_json" || ! -f "$result_db" ]]; then
            valid=0
        else
            reported_spec="$(
                sqlite3 "$result_db" \
                    'SELECT bot_spec_id FROM backtest_result;'
            )"
            [[ "$reported_spec" == "$stored_spec" ]] || valid=0
            central_bot_status="$(
                sqlite3 "$source_db" \
                    "SELECT status FROM bot WHERE bot_id=$bot_id;"
            )"
            central_process_status="$(
                sqlite3 "$source_db" "
                    SELECT status FROM process_state
                    WHERE target_kind='bot' AND target_id=$bot_id;
                "
            )"
            central_generation="$(
                sqlite3 "$source_db" "
                    SELECT generation FROM process_state
                    WHERE target_kind='bot' AND target_id=$bot_id;
                "
            )"
            validation_check "$result_db" central_bot_status equal stopped \
                "$central_bot_status" || valid=0
            validation_check "$result_db" central_process_status equal stopped \
                "$central_process_status" || valid=0
            validation_check "$result_db" central_process_generation greater 0 \
                "$central_generation" || valid=0
        fi

        if [[ $valid -eq 1 ]]; then
            case "$stored_spec" in
                macross_observer_bot)
                    validate_observer "$output" "$result_db" || valid=0
                    ;;
                macross_trade_bot)
                    validate_trade "$result_db" "$source_db_sql" || valid=0
                    ;;
                macross_grid_bot)
                    validate_grid "$result_db" "$source_db_sql" || valid=0
                    ;;
            esac
        fi

        if [[ $valid -eq 1 ]]; then
            attempt_records+="$(
                printf \
                    '{"run":%d,"exit":0,"btbot_elapsed_ms":%d,"report":%s}' \
                    "$run" "$elapsed_ms" "$run_json"
            )"$'\n'
            printf '%s\n' "$output" |
                grep -E '] (signaler initialized|controller stopped|tick reader stopped|btbot stopped)' |
                write_result_log "$result_log"
            log_statuses=("${PIPESTATUS[@]}")
            if [[ ${log_statuses[0]} -ne 0 ||
                  ${log_statuses[1]} -ne 0 ||
                  ${log_statuses[2]} -ne 0 ]]; then
                exit 1
            fi
            if ! printf 'run=%d result=PASS btbot_elapsed_ms=%d result_db=%s\n' \
                "$run" "$elapsed_ms" "$result_db" |
                write_result_log "$result_log"; then
                exit 1
            fi
            continue
        fi

        [[ $status -ne 0 ]] || status=1
        attempt_records+="$(
            printf \
                '{"run":%d,"exit":%d,"btbot_elapsed_ms":%d,"error":"validation failed"}' \
                "$run" "$status" "$elapsed_ms"
        )"$'\n'
        if ! printf '%s\n' "$output" | write_result_log "$result_log"; then
            exit 1
        fi
        if ! printf 'run=%d result=FAIL exit=%d stored_spec=%s reported_spec=%s\n' \
            "$run" "$status" "$stored_spec" "${reported_spec:-missing}" |
            write_result_log "$result_log"; then
            exit 1
        fi
        bot_failed=1
        suite_failed=1
        break
    done

    suite_elapsed_ms=$(( $(date +%s%3N) - suite_started_ms ))
    printf '%s' "$attempt_records" |
        "$stest_report" "$runs" "$sweep_id" "$bot_id" \
            "$suite_elapsed_ms" "$suite_json" |
        write_result_log "$result_log"
    report_statuses=("${PIPESTATUS[@]}")
    report_input_status=${report_statuses[0]}
    reporter_status=${report_statuses[1]}
    report_log_status=${report_statuses[2]}
    if [[ $report_log_status -ne 0 ]]; then
        exit 1
    fi
    if ! printf 'suite_report=%s result_log=%s\n' "$suite_json" "$result_log" |
        write_result_log "$result_log"; then
        exit 1
    fi
    if [[ $bot_failed -ne 0 ||
          $report_input_status -ne 0 ||
          $reporter_status -ne 0 ]]; then
        suite_failed=1
    fi
done <<< "$selected_bots"

exit "$suite_failed"
