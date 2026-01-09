Pseudocode
==========

UDP   
---

UDP uses datagrams, so receiveFrom will return whenever it receives anything. The buffer size is just the maximum size of the message, it doesn't have to be "filled". This example is for broadcasting.

### Receiver
```C
// the address we are listening for messages on
// we have no choice in IP, so use 0.0.0.0, INADDR_ANY, or leave the IP field empty
// the port should be whatever the sender sends to
// alternate names: sockaddr, resolve(udp)addr, 
InternetAddress addr

// a socket that plugs our program to the network. This is the "portal" to the outside world
// alternate names: conn
// UDP is sometimes called SOCK_DGRAM. You will sometimes also find UDPSocket or UDPConn as separate types
recvSock = new Socket(udp)

// bind the address we want to use to the socket
recvSock.bind(addr)


// a buffer where the received network data is stored
byte[1024] buffer  

// an empty address that will be filled with info about who sent the data
InternetAddress fromWho 

loop {
    // clear buffer (or just create a new one)
    
    // receive data on the socket
    // fromWho will be modified by ref here. Or it's a return value. Depends.
    // receive-like functions return the number of bytes received
    // alternate names: read, readFrom
    numBytesReceived = recvSock.receiveFrom(buffer, ref fromWho)
    
    // the buffer just contains a bunch of bytes, so you may have to explicitly convert it to a string
    
    // optional: filter out messages from ourselves
    if(fromWho.IP != localIP){
        // do stuff with buffer
    }
}
```

### Sender
```C

// if sending directly to a single remote machine:
    addr = new Address(remoteIP, remotePort)
    sock = new Socket(udp)
    
    // either: set up the socket to use a single remote address
        sock.connect(addr)
        sock.send(message)
    // or: set up the remote address when sending
        sock.sendTo(message, addr)
        
// if sending on broadcast:
// you have to set up the BROADCAST socket option before calling connect / sendTo
    broadcastIP = #.#.#.255 // First three bytes are from the local IP, or just use 255.255.255.255
    addr = new InternetAddress(broadcastIP, port)
    sendSock = new Socket(udp) // UDP, aka SOCK_DGRAM
    sendSock.setOption(broadcast, true)
    sendSock.sendTo(message, addr)
```


TCP
---

For TCP sockets, you may find that a call to recv() will block until the entire buffer has been filled. Either accept fixed-size messages of size 1024 (which is what the server sends), or find some functionality that avoids this.

A handy diagram describing [Berkeley Sockets](http://en.wikipedia.org/wiki/Berkeley_sockets) on Wikipedia

### Client
```C
addr = new InternetAddress(serverIP, serverPort) 
sock = new Socket(tcp) // TCP, aka SOCK_STREAM
sock.connect(addr)
// use sock.recv() and sock.send(), just like with UDP
```

### Server
```C
// Send a message to the server:  "Connect to: " <your IP> ":" <your port> "\0"

// do not need IP, because we will set it to listening state
addr = new InternetAddress(localPort)
acceptSock = new Socket(tcp)

// You may not be able to use the same port twice when you restart the program, unless you set this option
acceptSock.setOption(REUSEADDR, true)
acceptSock.bind(addr)

loop {
    // backlog = Max number of pending connections waiting to connect()
    newSock = acceptSock.listen(backlog)

    // Spawn new thread to handle recv()/send() on newSock
}
```
   

    
Shutting down sockets
=====================
Use SocketOption REUSEADDRESS, so you can use the same address when the program restarts. This way you can afford to be lazy, and not use the proper shutdown()/close() calls.


Non-blocking sockets and select()
=================================
### Aka avoiding the use of a new thread for each connection

[From the Python Sockets HowTo](http://docs.python.org/2/howto/sockets.html#non-blocking-sockets), but the concept is the same in any language.
