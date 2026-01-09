The network server
------------------

The source code for a modified version of the network server (for single-machine use) is found in this repository. To run it, you must either download [a D compiler](https://dlang.org/download#dmd), or download one of the pre-compiled binaries from the [releases tab](../../releases/latest).

### Compiling from source

Run `rdmd server_wfh.d` from the command line.

Port numbers
------------

Normally, it is only possible to bind one connection to one program, so if the server binds to UDP port `20000`, you will get an error if you do the same. There are two workarounds to this:

 - Use a set of ports, one for each "direction" of messages. For the UDP part, the work-from-home server receives messages on port `20000`, and replies back to you on port `20001`. For the TCP part, a connection is identified by the IP and ports of *both* endpoints (instead of just the IP and port of the local endpoint), so there is no change for this part.
 - Use the socket option `SO_REUSEADDR` *before* binding the connection to the program. This lets multiple programs (or threads in the same program) all receive the messages to the bound address. If you want to use the same UDP port for sending and receiving (`20000`, in this case), start the server with the `-s` option (or `rdmd server_wfh.d -s`).
 
A note about Go: There is no elegant way to set socket options before binding in Go. You will have to create a PacketConn, and on Windows you will have to do everything from scratch. See [these files](https://github.com/TTK4145/Network-go/tree/master/network/conn) for examples and more information about why. I would encourage sticking to asymmetric ports (one for each direction) for this exercise.




