#!/bin/bash

ID=$1
PORT=$2
PRIMARY_PID=$3

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$ID" ] || [ -z "$PORT" ]; then
echo "Usage: ./launch.sh <ID> <port_number>"
exit 1
fi

cd "$SCRIPT_DIR"

go build -o elevator_bin main.go || { echo "Build failed."; exit 1; }

if [ -n "$PRIMARY_PID" ]; then
echo "[$(date +%H:%M:%S)][Launcher]: Backup Mode."

while kill -0 "$PRIMARY_PID" 2>/dev/null; do
sleep 0.5
done

sleep 0.5
fi

echo "[$(date +%H:%M:%S)][Launcher]: Primary Mode."

gnome-terminal -- bash -c "cd $(pwd); ./launch.sh $ID $PORT $$"

./elevator_bin -port="$PORT" -id="$ID"

sleep 1
