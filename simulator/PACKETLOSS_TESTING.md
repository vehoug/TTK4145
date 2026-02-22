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

- place a cab order on simulator port `15657` so node 0 starts moving,
- then press Enter in the terminal running the script.
