#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN="${RUN:-$ROOT_DIR/artifacts/packetloss_$(date +%Y%m%d_%H%M%S)}"
SIM_BIN="$ROOT_DIR/simulator/SimElevatorServer"
PACKETLOSS_SRC="$ROOT_DIR/simulator/packetloss.d"
PACKETLOSS_BIN="$ROOT_DIR/simulator/packetloss"
ELEVATOR_DIR="$ROOT_DIR/elevator"
CONTROLLER_LOG="$RUN/controller.log"

iso_now() {
    date +"%Y-%m-%dT%H:%M:%S%z"
}

START_TS="$(iso_now)"
declare -a PIDS=()
MANUAL_WAIT_SECONDS="${MANUAL_WAIT_SECONDS:-10}"
SIM0_INTERACTIVE="${SIM0_INTERACTIVE:-1}"
SIM0_TERMINAL="${SIM0_TERMINAL:-auto}"
SIM0_STARTED_INTERACTIVE="false"

BASELINE_SECONDS="${BASELINE_SECONDS:-20}"
MODERATE_SECONDS="${MODERATE_SECONDS:-20}"
PARTITION_SECONDS="${PARTITION_SECONDS:-10}"
RECOVERY_SECONDS="${RECOVERY_SECONDS:-15}"
WATCHDOG_BLOCK_SECONDS="${WATCHDOG_BLOCK_SECONDS:-8}"
WATCHDOG_RECOVERY_SECONDS="${WATCHDOG_RECOVERY_SECONDS:-15}"

log() {
    local msg="$1"
    echo "[$(iso_now)] $msg" | tee -a "$CONTROLLER_LOG"
}

require_cmd() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Missing required command: $cmd" >&2
        exit 1
    fi
}

start_sim0_interactive() {
    if [[ "$SIM0_INTERACTIVE" != "1" ]]; then
        return 1
    fi

    if [[ -z "${DISPLAY:-}" && -z "${WAYLAND_DISPLAY:-}" ]]; then
        log "No GUI display detected; cannot open interactive simulator window for port 15657."
        return 1
    fi

    local preferred=""
    case "$SIM0_TERMINAL" in
        auto)
            if command -v gnome-terminal >/dev/null 2>&1; then
                preferred="gnome-terminal"
            elif command -v xterm >/dev/null 2>&1; then
                preferred="xterm"
            fi
            ;;
        gnome-terminal|xterm)
            preferred="$SIM0_TERMINAL"
            ;;
        none)
            return 1
            ;;
        *)
            log "Unknown SIM0_TERMINAL='$SIM0_TERMINAL'; falling back to non-interactive mode."
            return 1
            ;;
    esac

    if [[ -z "$preferred" ]]; then
        log "No supported terminal emulator found (gnome-terminal/xterm); falling back to non-interactive mode."
        return 1
    fi

    local sim_cmd
    sim_cmd="stdbuf -oL -eL '$SIM_BIN' --port 15657 2>&1 | tee -a '$RUN/sim0.log'"

    if [[ "$preferred" == "gnome-terminal" ]]; then
        gnome-terminal --title="SimElevatorServer 15657" -- bash -lc "$sim_cmd"
    else
        xterm -T "SimElevatorServer 15657" -e bash -lc "$sim_cmd" &
    fi

    SIM0_STARTED_INTERACTIVE="true"
    log "Started interactive simulator on port 15657 in $preferred."
    sleep 1
    return 0
}

stop_sim0_interactive() {
    if [[ "$SIM0_STARTED_INTERACTIVE" != "true" ]]; then
        return
    fi

    local sim0_pids
    sim0_pids="$(pgrep -f "$SIM_BIN --port 15657" || true)"
    if [[ -n "$sim0_pids" ]]; then
        kill $sim0_pids >/dev/null 2>&1 || true
    fi
}

cleanup() {
    local exit_code="$1"
    set +e

    if [[ -x "$PACKETLOSS_BIN" ]]; then
        sudo "$PACKETLOSS_BIN" -f >/dev/null 2>&1 || true
    fi

    for pid in "${PIDS[@]:-}"; do
        kill "$pid" >/dev/null 2>&1 || true
    done
    for pid in "${PIDS[@]:-}"; do
        wait "$pid" >/dev/null 2>&1 || true
    done

    stop_sim0_interactive

    log "Cleanup finished. Artifacts in: $RUN"
    exit "$exit_code"
}

trap 'cleanup $?' EXIT
trap 'exit 130' INT TERM

mkdir -p "$RUN"
: > "$CONTROLLER_LOG"

if [[ "$(uname -s)" != "Linux" ]]; then
    echo "This script must run on Linux (found: $(uname -s))." >&2
    exit 1
fi

if [[ "$(uname -m)" != "x86_64" && "$(uname -m)" != "amd64" ]]; then
    echo "This script expects x86_64/amd64 (found: $(uname -m))." >&2
    exit 1
fi

require_cmd sudo
require_cmd go
require_cmd rg
require_cmd iptables
require_cmd stdbuf
require_cmd git

if [[ ! -x "$SIM_BIN" ]]; then
    echo "Simulator binary not found or not executable: $SIM_BIN" >&2
    exit 1
fi

if [[ ! -f "$PACKETLOSS_SRC" && ! -x "$PACKETLOSS_BIN" ]]; then
    echo "Missing packetloss source and binary. Expected one of:" >&2
    echo "  $PACKETLOSS_SRC" >&2
    echo "  $PACKETLOSS_BIN" >&2
    exit 1
fi

if [[ -f "$PACKETLOSS_SRC" ]]; then
    if [[ ! -x "$PACKETLOSS_BIN" || "$PACKETLOSS_SRC" -nt "$PACKETLOSS_BIN" ]]; then
        require_cmd ldc2
        log "Compiling packetloss helper from $PACKETLOSS_SRC"
        ldc2 -O2 -of "$PACKETLOSS_BIN" "$PACKETLOSS_SRC"
        chmod +x "$PACKETLOSS_BIN"
    fi
fi

packetloss_flush() {
    log "Flushing packet loss rules"
    sudo "$PACKETLOSS_BIN" -f >> "$CONTROLLER_LOG" 2>&1
}

packetloss_apply() {
    local ports="$1"
    local rate="$2"
    log "Applying packet loss: ports=$ports rate=$rate"
    sudo "$PACKETLOSS_BIN" -p "$ports" -r "$rate" >> "$CONTROLLER_LOG" 2>&1
}

start_processes() {
    log "Starting simulator processes"
    if ! start_sim0_interactive; then
        log "Starting simulator on port 15657 in background (non-interactive)."
        stdbuf -oL -eL "$SIM_BIN" --port 15657 > "$RUN/sim0.log" 2>&1 & PIDS+=("$!")
    fi
    stdbuf -oL -eL "$SIM_BIN" --port 15658 > "$RUN/sim1.log" 2>&1 & PIDS+=("$!")
    stdbuf -oL -eL "$SIM_BIN" --port 15659 > "$RUN/sim2.log" 2>&1 & PIDS+=("$!")

    log "Starting elevator nodes"
    stdbuf -oL -eL bash -lc "cd '$ELEVATOR_DIR' && go run . -id=0 -port=15657" > "$RUN/node0.log" 2>&1 & PIDS+=("$!")
    stdbuf -oL -eL bash -lc "cd '$ELEVATOR_DIR' && go run . -id=1 -port=15658" > "$RUN/node1.log" 2>&1 & PIDS+=("$!")
    stdbuf -oL -eL bash -lc "cd '$ELEVATOR_DIR' && go run . -id=2 -port=15659" > "$RUN/node2.log" 2>&1 & PIDS+=("$!")
}

extract_results() {
    log "Extracting key events"
    rg -n "lost network connection|reconnected to network|ActiveStatus:false|ActiveStatus:true|panic|Lost connection to Elevator Server" \
        "$RUN"/node*.log > "$RUN/events.txt" || true
    rg -n "Common state:" "$RUN/node0.log" > "$RUN/node0_commonstate.txt" || true
}

first_match() {
    local pattern="$1"
    rg -n -m1 "$pattern" "$RUN"/node*.log 2>/dev/null | head -n1 || true
}

evaluate_checks() {
    CHECK1="FAIL"
    CHECK2="FAIL"
    CHECK3="FAIL"
    CHECK4="FAIL"
    CHECK5="FAIL"

    EVIDENCE1="$(first_match "lost network connection")"
    EVIDENCE2A="$(first_match "lost network connection")"
    EVIDENCE2B="$(first_match "reconnected to network")"
    EVIDENCE3="$(first_match "ActiveStatus:false")"
    EVIDENCE4A="$(first_match "ActiveStatus:false")"
    EVIDENCE4B="$(first_match "ActiveStatus:true")"
    EVIDENCE5="$(first_match "panic|fatal error|Lost connection to Elevator Server")"

    if rg -q "lost network connection" "$RUN"/node*.log; then
        CHECK1="PASS"
    fi

    if rg -q "lost network connection" "$RUN"/node*.log && rg -q "reconnected to network" "$RUN"/node*.log; then
        CHECK2="PASS"
    fi

    if rg -q "ActiveStatus:false" "$RUN"/node0.log; then
        CHECK3="PASS"
    fi

    local false_line true_line
    false_line="$(rg -n "ActiveStatus:false" "$RUN/node0.log" | head -n1 | cut -d: -f1 || true)"
    if [[ -n "$false_line" ]]; then
        true_line="$(rg -n "ActiveStatus:true" "$RUN/node0.log" | awk -F: -v start="$false_line" '$1 > start {print $1; exit}' || true)"
        if [[ -n "$true_line" ]]; then
            CHECK4="PASS"
        fi
    fi

    if ! rg -q "panic|fatal error|Lost connection to Elevator Server" "$RUN"/node*.log; then
        CHECK5="PASS"
    fi
}

generate_report() {
    local end_ts os_info go_info commit
    end_ts="$(iso_now)"
    os_info="$(uname -a)"
    go_info="$(go version)"
    commit="$(git -C "$ROOT_DIR" rev-parse HEAD 2>/dev/null || echo "unknown")"

    evaluate_checks

    cat > "$RUN/report.md" <<EOF
# Packet Loss Test Report

## Environment
- OS/Kernel: \`$os_info\`
- Go: \`$go_info\`
- Commit: \`$commit\`
- Test start: \`$START_TS\`
- Test end: \`$end_ts\`
- Artifact dir: \`$RUN\`

## Scenario Table
| Scenario | Stimulus | Expected response | Observation | Result |
|---|---|---|---|---|
| Baseline | Flush rules, ${BASELINE_SECONDS}s | Stable common state updates | See \`$RUN/node*.log\` | N/A |
| Moderate loss | ports 58735,58750 rate 0.25, ${MODERATE_SECONDS}s | Degraded but functioning synchronization | See \`$RUN/events.txt\` | N/A |
| Full partition | ports 58735,58750 rate 1.0, ${PARTITION_SECONDS}s | At least one node logs network loss | ${EVIDENCE1:-No match} | $CHECK1 |
| Recovery | Flush rules, ${RECOVERY_SECONDS}s | Reconnect + state sync resumes | loss=${EVIDENCE2A:-No match}; reconnect=${EVIDENCE2B:-No match} | $CHECK2 |
| Watchdog block | ports 15657 rate 1.0, ${WATCHDOG_BLOCK_SECONDS}s | Node0 ActiveStatus becomes false while moving | ${EVIDENCE3:-No match} | $CHECK3 |
| Watchdog recovery | Flush rules, ${WATCHDOG_RECOVERY_SECONDS}s | Node0 ActiveStatus returns true after progress | false=${EVIDENCE4A:-No match}; true=${EVIDENCE4B:-No match} | $CHECK4 |
| Crash check | Full test window | No uncontrolled crash/panic in node logs | ${EVIDENCE5:-No panic/fatal/lost-server match} | $CHECK5 |

## Key Findings
- Network loss signal: \`${EVIDENCE1:-not observed}\`
- Reconnect signal: \`${EVIDENCE2B:-not observed}\`
- Watchdog false signal: \`${EVIDENCE3:-not observed}\`
- Watchdog true-after-false signal: \`${EVIDENCE4B:-not observed}\`
- Panic/fatal signal: \`${EVIDENCE5:-none observed}\`

## Conclusion
- System reaction under packet loss is **$([[ "$CHECK1" == "PASS" && "$CHECK2" == "PASS" && "$CHECK3" == "PASS" && "$CHECK4" == "PASS" && "$CHECK5" == "PASS" ]] && echo "acceptable (all checks passed)" || echo "needs attention (one or more checks failed)")**.
- Detailed evidence is available in:
  - \`$RUN/events.txt\`
  - \`$RUN/node0_commonstate.txt\`
  - \`$RUN/node0.log\`, \`$RUN/node1.log\`, \`$RUN/node2.log\`
  - \`$RUN/sim0.log\`, \`$RUN/sim1.log\`, \`$RUN/sim2.log\`
EOF

    log "Report written to $RUN/report.md"
}

log "Packet-loss integration test started"
start_processes
packetloss_flush

log "Settling period before scenarios (8s)"
sleep 8

log "Scenario 1/5: Baseline (${BASELINE_SECONDS}s)"
packetloss_flush
sleep "$BASELINE_SECONDS"

log "Scenario 2/5: Moderate network loss (${MODERATE_SECONDS}s)"
packetloss_apply "58735,58750" "0.25"
sleep "$MODERATE_SECONDS"

log "Scenario 3/5: Full network partition (${PARTITION_SECONDS}s)"
packetloss_apply "58735,58750" "1.0"
sleep "$PARTITION_SECONDS"

log "Scenario 4/5: Recovery (${RECOVERY_SECONDS}s)"
packetloss_flush
sleep "$RECOVERY_SECONDS"

log "Scenario 5/5: Watchdog on node0 simulator connection"
if [[ "$SIM0_STARTED_INTERACTIVE" == "true" ]]; then
    log "Manual step required: in the simulator window for port 15657, place a cab order so node0 is moving."
else
    log "Manual step required: no interactive simulator window was opened."
    log "Run with SIM0_INTERACTIVE=1 and a GUI terminal (gnome-terminal/xterm) for easier manual watchdog stimulus."
fi
if [[ -t 0 ]]; then
    read -r -p "Press Enter when node0 is moving (or Ctrl+C to abort): " _
else
    log "Non-interactive shell detected; waiting ${MANUAL_WAIT_SECONDS}s for manual stimulus."
    sleep "$MANUAL_WAIT_SECONDS"
fi

packetloss_apply "15657" "1.0"
sleep "$WATCHDOG_BLOCK_SECONDS"
packetloss_flush
sleep "$WATCHDOG_RECOVERY_SECONDS"

extract_results
generate_report
log "Test run complete"
