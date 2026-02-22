# TTK4145 Real-time Programming

## Distributed Elevator Control System
>*Software for controlling $n$ elevators working in parallel across $m$ floors.*

### How to launch elevators:
1. Navigate to `elevator` directory.
2. Launch elevator nodes from the command line in separate terminals using the shell script (IDs must be unique, and starting from zero), and remember to use different port numbers if simulating the system with several elevators on the same computer:
```bash
# Launch elevator node with process pairs
./launch.sh <ID> <PORT>

# Alternatively (without process pairs):
go run main.go -port=<PORT> -id=<ID>
```


