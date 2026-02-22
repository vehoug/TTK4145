# Packet-Loss Test Workflow

This folder contains:

- `packetloss.d`: helper that flushes and applies packet-loss rules using `iptables`.
- `run_packetloss_test.sh`: end-to-end scenario runner for simulator + 3 nodes.

## Prerequisites

- Linux `x86_64` host
- `sudo` access
- `iptables`
- `go`
- `rg` (ripgrep)
- `stdbuf`
- `ldc2` (only required if `simulator/packetloss` is not already built)

## Run

From project root:

```bash
chmod +x simulator/run_packetloss_test.sh
./simulator/run_packetloss_test.sh
```

Optional output directory:

```bash
RUN=artifacts/packetloss_custom ./simulator/run_packetloss_test.sh
```

Force interactive simulator window for node 0 (recommended for watchdog step):

```bash
SIM0_INTERACTIVE=1 SIM0_TERMINAL=auto ./simulator/run_packetloss_test.sh
```

Select terminal emulator explicitly:

```bash
SIM0_INTERACTIVE=1 SIM0_TERMINAL=gnome-terminal ./simulator/run_packetloss_test.sh
# or
SIM0_INTERACTIVE=1 SIM0_TERMINAL=xterm ./simulator/run_packetloss_test.sh
```

## Output

Each run creates:

- `artifacts/packetloss_<timestamp>/node0.log`
- `artifacts/packetloss_<timestamp>/node1.log`
- `artifacts/packetloss_<timestamp>/node2.log`
- `artifacts/packetloss_<timestamp>/sim0.log`
- `artifacts/packetloss_<timestamp>/sim1.log`
- `artifacts/packetloss_<timestamp>/sim2.log`
- `artifacts/packetloss_<timestamp>/events.txt`
- `artifacts/packetloss_<timestamp>/node0_commonstate.txt`
- `artifacts/packetloss_<timestamp>/report.md`

The script includes one manual step before the watchdog scenario:

- if an interactive window opened for port `15657`, focus that window and place a cab order so node 0 starts moving,
- cab keys in `simulator.con`: `z` (floor 0), `x` (floor 1), `c` (floor 2), `v` (floor 3),
- then press Enter in the terminal running the script.
