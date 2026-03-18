# Distributed Elevator Control System

## System Overview:
The system is based on a peer-to-peer architecture, utilizing UDP-broadcast for communication between elevators.

**Distributor (`distributor`):** The distributor is responsible for making sure the common state is synchronized/agreed upon between the elevators, and handles network loss by making the elevator operate independently and merging back when reconnected. The synchronized common state is sent to the `lights` and `assigner` module.

**Assigner (`assigner`):** The assigner determines which elevator should handle each hall request in the most efficient way based on the synchronized common state. It sends these assignments to the `elevcontrol`, which executes the decisions.

**Elevator Control (`elevcontrol`):** Elevator control handles the local control logic for a single elevator. It receives orders from `assigner` and sends the elevators local state combined with completed orders to `distributor`.

**Lights (`lights`):** Turns the elevator panel lights on or off based on the synchronized common state, ensuring that only orders that are assigned can activate a light. 

## How to launch elevators:
1. Navigate to the project root.
2. Launch elevator nodes from the command line using the shell script (IDs must be unique, and starting from zero):
```bash
# Launch elevator node with process pairs (for Linux with GNOME)
./launch.sh <ID> <PORT>

# Alternatively (without process pairs):
go run main.go -port=<PORT> -id=<ID>
```


