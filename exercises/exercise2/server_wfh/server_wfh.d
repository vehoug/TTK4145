#!/usr/bin/env rdmd
import core.sync.mutex;
import core.thread;
import std.algorithm;
import std.concurrency;
import std.conv;
import std.datetime;
import std.getopt;
import std.process;
import std.range;
import std.regex;
import std.socket;
import std.stdio;
import std.string;

__gshared uint      bufferSize      = 1024;
__gshared string    localIP;
__gshared Mutex     consoleMtx;

__gshared ushort    udpAnnouncePort = 30000;
__gshared ushort    udpRecvPort     = 20000;
__gshared ushort    udpSendPort     = 20001;
__gshared bool      symmetric       = false;




void main(string[] args){
    consoleMtx = new Mutex;
    localIP = new TcpSocket(new InternetAddress("www.google.com", 80)).localAddress.toAddrString;
    string bcastIP = localIP[0 .. localIP.lastIndexOf(".")+1] ~ "255";

    getopt(args,
        std.getopt.config.passThrough,
        "b|bufsize",    &bufferSize,
        "s|symmetric",  &symmetric,
    );
    if(symmetric){
        udpSendPort = udpRecvPort;
    }



    spawn(&TCPAccepter, 34933.to!ushort, MessageFormat.fixed);
    spawn(&TCPAccepter, 33546.to!ushort, MessageFormat.delim);
    spawn(&UDPEchoer, localIP, bcastIP, udpRecvPort, udpSendPort);
    spawn(&UDPAnnouncer, localIP, bcastIP, udpAnnouncePort);

    writeln("Server started. Local IP: ", localIP);

    while(true){ Thread.sleep(1.hours); }
}




void TCPAccepter(ushort port, MessageFormat msgFmt){
    Socket acceptSock  = new TcpSocket();
    Socket newSock;

    acceptSock.setOption(SocketOptionLevel.SOCKET, SocketOption.REUSEADDR, 1);
    acceptSock.bind(new InternetAddress(port));
    acceptSock.listen(10);

    while(true){
        newSock = acceptSock.accept();
        writefcln(Color.tcp_info, "\nTCP\t%s\nAccepted socket %s", Clock.currTime, newSock.remoteAddress);
        spawn(&TCPReceiver, cast(shared)newSock, msgFmt);
    }
}

void TCPReceiver(shared Socket ssocket, MessageFormat msgFmt){
    auto    sock    = cast(Socket)ssocket;
    char[]  buf     = new char[](bufferSize);
    string  port    = sock.localAddress.toPortString;

    writefcln(Color.tcp_info, "\nTCP\t%s\nReceive thread started for\n    :%s <-> %s\n", Clock.currTime, port, sock.remoteAddress);

    sock.sendFmt(msgFmt, format("Hello from TCP server! (server-side port: %s)", port));

    while(sock.isAlive && sock.receive(buf) > 0){
        auto n = buf.countUntil('\0');
        string s = buf[0..n].idup;

        writefcln(Color.tcp_data, "\nTCP   %s\nFrom: %s (to :%s):\n    %s",
            Clock.currTime, sock.remoteAddress, sock.localAddress.toPortString, s);

        if(s.skipOver("Connect to:")){
            auto capture = s.matchFirst(r"(?P<addr>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(?P<port>\d{1,5})");
            if(capture.length != 3){
                sock.sendFmt(msgFmt, "The TCP server was unable to properly read the \"Connect to:\" command");
            } else {
                Socket newSock;
                try {
                    newSock = new TcpSocket(new InternetAddress(capture["addr"], capture["port"].to!ushort));
                    writefcln(Color.tcp_info, "\nTCP\t%s\nConnected to\n    :%s <-> %s:%s\n", Clock.currTime, port, capture["addr"], capture["port"]);
                    sock.sendFmt(msgFmt, format("The TCP server connected to %s:%s", capture["addr"], capture["port"]));
                    spawn(&TCPReceiver, cast(shared)newSock, msgFmt);
                } catch(Exception e){
                    writefcln(Color.tcp_info, "\nTCP\t%s\nUnable to connect to\n    :%s <-> %s:%s\n    %s\n", Clock.currTime, port, capture["addr"], capture["port"], e.msg);
                    sock.sendFmt(msgFmt, format("The TCP server was unable to connect to %s:%s: \"%s\"", capture["addr"], capture["port"], e.msg));
                }
            }
        } else {
            sock.sendFmt(msgFmt, format("You said: \"%s\" (server-side port: %s)", s, port));
        }
        buf[] = '\0';
    }

    writefcln(Color.tcp_info, "\nTCP\t%s\nDisconnected\n    :%s <|> %s\n", Clock.currTime, port, sock.remoteAddress);
    sock.shutdown(SocketShutdown.BOTH);
    sock.close();
}


enum MessageFormat {
    fixed,
    delim
}

void sendFmt(Socket sock, MessageFormat msgFmt, const char[] str){
    final switch(msgFmt) with(MessageFormat){
    case fixed:
        if(str.length > bufferSize){
            sock.send(str[0..bufferSize]);
            sock.sendFmt(fixed, str[bufferSize..$]);
        } else {
            sock.send(str ~ ['\0'].cycle.take(bufferSize-str.length).to!(char[]));
        }
        break;
    case delim:
        sock.send(str ~ "\0");
        break;
    }
}




void UDPEchoer(string localIP, string bcastIP, const ushort recvPort, const ushort sendPort){
    auto sendSock = new UdpSocket();
    sendSock.setOption(SocketOptionLevel.SOCKET, SocketOption.BROADCAST, 1);
    sendSock.setOption(SocketOptionLevel.SOCKET, SocketOption.REUSEADDR, 1);


    auto recvSock = new UdpSocket();
    recvSock.setOption(SocketOptionLevel.SOCKET, SocketOption.BROADCAST, 1);
    recvSock.setOption(SocketOptionLevel.SOCKET, SocketOption.REUSEADDR, 1);
    recvSock.bind(new InternetAddress(recvPort));

    char[]  buf         = new char[](bufferSize);
    Address remoteAddr  = new UnknownAddress;

    while(1){
        buf[] = '\0';
        auto r = recvSock.receiveFrom(buf, remoteAddr);
        if(r <= 0){
            writefcln(Color.udp_info, "\nUDP   %s\nReceive error: \"%s\" (%s)", Clock.currTime, recvSock.getErrorText, r);
        } else {
            if(!symmetric || !buf.startsWith("You said")){
                writefcln(Color.udp_data, "\nUDP   %s\nFrom: %s (to :%s):\n    %s\n",
                    Clock.currTime, remoteAddr, recvPort, (cast(string)buf).strip('\0'));
                sendSock.sendTo(
                    cast(ubyte[])format("You said: %s", buf[0..r]),
                    new InternetAddress(remoteAddr.toAddrString, sendPort)
                );
            }
        }
    }
}


void UDPAnnouncer(string localIP, string bcastIP, const ushort port){
    auto bcastAddr = new InternetAddress(bcastIP, port);

    auto sendSock = new UdpSocket();
    sendSock.setOption(SocketOptionLevel.SOCKET, SocketOption.BROADCAST, 1);
    sendSock.setOption(SocketOptionLevel.SOCKET, SocketOption.REUSEADDR, 1);

    while(1){
        auto r = sendSock.sendTo(cast(ubyte[])format("Hello from UDP server at %s!", localIP), bcastAddr);
        Thread.sleep(1.seconds);
    }
}


version(Windows){
    import core.sys.windows.windef;
    import core.sys.windows.winbase;
    import core.sys.windows.wincon;

    enum Color : ushort {
        tcp_info    = FOREGROUND_RED | FOREGROUND_BLUE,
        tcp_data    = FOREGROUND_RED | FOREGROUND_GREEN,
        udp_info    = FOREGROUND_GREEN,
        udp_data    = FOREGROUND_GREEN | FOREGROUND_BLUE,
    }
    enum BACKGROUND_MASK = BACKGROUND_BLUE | BACKGROUND_GREEN | BACKGROUND_RED | BACKGROUND_INTENSITY;

    void writefcln(Char, A...)(Color c, in Char[] fmt, A args){
        synchronized(consoleMtx){
            HANDLE hConsole = GetStdHandle(STD_OUTPUT_HANDLE);
            CONSOLE_SCREEN_BUFFER_INFO consoleInfo;

            GetConsoleScreenBufferInfo(hConsole, &consoleInfo);
            scope(exit) SetConsoleTextAttribute(hConsole, consoleInfo.wAttributes);

            SetConsoleTextAttribute(hConsole, 
                c | 
                (consoleInfo.wAttributes & BACKGROUND_MASK) | 
                ((consoleInfo.wAttributes & BACKGROUND_INTENSITY) ? 0 : FOREGROUND_INTENSITY));
            writefln(fmt, args);
        }
    }
} else {
    enum Color {
        tcp_info    = "\x1b[38;5;220m",
        tcp_data    = "\x1b[38;5;226m",
        udp_info    = "\x1b[38;5;44m",
        udp_data    = "\x1b[38;5;123m",
        reset       = "\x1b[0m",
    }

    void writefcln(Char, A...)(Color c, in Char[] fmt, A args){
        synchronized(consoleMtx){
            writefln(c ~ fmt ~ Color.reset, args);
        }
    }
}










