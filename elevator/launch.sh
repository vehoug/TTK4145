#!/bin/bash

ID=$1
PORT=$2
PRIMARY_PID=$3

trap "echo -e '\nInterrupted. Closing terminal...'; exit 0" SIGINT SIGTERM

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$ID" ] || [ -z "$PORT" ]; then
    echo "Usage: ./launch.sh <ID> <port_number> [PRIMARY_PID]"
    exit 1
fi

cd "$SCRIPT_DIR"

echo "Building project..."
go build -o elevator_bin main.go || { echo "Build failed."; exit 1; }

# Backup Phase
if [ -n "$PRIMARY_PID" ]; then
    echo "--------------------------------------"
    echo " Backup Mode (Node ID: $ID)"
    echo " Monitoring Primary PID: $PRIMARY_PID"
    echo "--------------------------------------"

    while kill -0 "$PRIMARY_PID" 2>/dev/null; do
        sleep 0.5
    done

    echo "Primary crashed or closed. Taking over as Primary..."
    sleep 0.5
fi

# Primary Phase
echo "--------------------------------------"
echo " Primary Mode (Node ID: $ID)"
echo "--------------------------------------"


echo "Spawning Backup terminal..."
gnome-terminal -- bash -c "cd $(pwd); ./launch.sh $ID $PORT $$"

./elevator_bin -port="$PORT" -id="$ID"

echo "Elevator $ELEVATOR_PID exited. Closing terminal..."
sleep 1
