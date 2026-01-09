Exercise 2 : Networking Essentials
==================================

This exercise serves three goals:
 - *Understanding networking:*  
   The main goal is to find or create a networking module that can be used in your project, and understand both what it does and how it does it such that you are able to make proper use of it.
 - *Experience with concurrency:*  
   It is very likely that you will be using some form of concurrent execution in the project, and you will find good use for this already in this exercise.
 - *Experience with a new programming language:*  
   It is also very likely that you will be using a new programming language for the project, so you should use this exercise as a case study to try, learn - and possibly even reject - whatever language(s) you want.

This exercise does not make any explicit requirements or recommendations for programming language, concurrency usage, code quality, or even general "feature-completeness" of whatever code you produce. Instead, the goal is to develop your understanding of networking, and expand your toolbelt for creating and testing software with a netorking component. You should keep the end goal in sight (the project, and what it needs), and not get (too) bogged down by the details. But the details are still important, as they are what give you the understanding needed.

From here, the exercise is roughly divided into three parts:
 1. The first part is to make you more familiar with using TCP and UDP for communication between processes running on different machines. 
 3. The second part will have you consider the things you have learned about these two protocols, and either create or find (or modify) a network module that *you* can use in *your* project.   
 2. The third part is to get introduced to some useful tools when testing networked code: remote logins, file transfers, and packet loss testing.

 
*Only the first part should be handed in for approval, though you should make use of the student assistants to ask whatever questions you have of the latter two parts too.*

Exactly how you do communication for the project is up to you, so if you want to venture out into the land of libraries you should make sure that the library satisfies all your requirements, and knowing how TCP and UDP work will help you make an informed decision. You should also check the license.

---

*If you are doing this exercise in Erlang or Elixir, you will want to skip the part about TCP, and learn how to use Nodes instead. But UDP might still be useful, in order to make something that can auto-detect the IP addresses of the other nodes on the network. Some links to get you started:*
 - [*Distributed Erlang*](http://erlang.org/doc/reference_manual/distributed.html)
 - [*Elixir School on OTP Distribution*](https://elixirschool.com/en/lessons/advanced/otp_distribution)
 - [*Elixir documentation on `Node`*](https://hexdocs.pm/elixir/Node.html). 

---


1.1: UDP
--------

Since network programming requires that both sending and receiving works at the same time, it becomes quite difficult to get things right when starting out. This exercise comes with a pre-made network server, so that you can incrementally create one new thing at a time. On the lab, this networking server is run on the machine near the student assistants (so you do not need to do anything to set it up), and also prints out a full log of everything. 

If you are working from home, you may want to run such a server yourself, and [instructions to do so are found in this document](./working-from-home.md).

Practical tips:
 - Sharing a socket between threads should not be a problem, although reading from a socket in two threads will probably mean that only one of the threads gets the message. If you are using blocking sockets, you could create a "receiving"-thread for each socket. 
   - Alternatively, you can use socket sets and the [`select()`](http://en.wikipedia.org/wiki/Select_%28Unix%29) function (or its equivalent). Note that this is not the same "select" as in message passing - although its functionality is the same: it gives you the ability wait for activity on several connections simultaneously.
 - Be nice to the network: Put some amount of `sleep()` or equivalent in the loops that send messages. The network at the lab will be shut off if IT finds a DDOS-esque level of traffic. Yes, this has happened before. Several times.
 - You can find [some pseudocode here](resources.md) to get you started.


### Receiving UDP packets, and finding the server IP:
The server broadcasts its own IP address on UDP port `30000`. Listen for messages on this port to find it. You should write down the IP address, as you will need it for again later in the exercise.

### Sending UDP packets:
The server will respond to any message you send to it. Try sending a message to the server IP on port `20000 + n` where `n` is the number of the workspace you're sitting at. Listen to messages from the server and print them to a terminal to confirm that the messages are in fact being responded to.

- The server will act the same way if you send a broadcast (`#.#.#.255` or `255.255.255.255`) instead of sending directly to the server.
  - If you use broadcast, the messages will loop back to you! The server prefixes its reply with "You said: ", so you can tell if you are getting a reply from the server or if you are just listening to yourself.
- You are free to mess with the people around you by using the same port as them, but they may not appreciate it.
- It may be easier to use two sockets: one for sending and one for receiving. You might also find it is easier if these two are separated into their own threads.


1.2: TCP
--------

There are three common ways to format a message: (Though there are probably others)
 - 1: Always send fixed-sized messages
 - 2: Send the message size with each message (as part of a fixed-size header)
 - 3: Use some kind of marker to signify the end of a message

The TCP server can send you two of these:
 - Fixed size messages of size `1024`, if you connect to port `34933`
 - Delimited messages that use `\0` as the marker, if you connect to port `33546`

The server will read until it encounters the first `\0`, regardless. Strings in most programming languages are already null-terminated, but in case they aren't you will have to append your own null character.

*TCP guarantees that packets arrive in the order they are sent. But this does not mean that it guarantees that these packets are delivered individually (or that they are delivered at all, since you could always apply scissors to the network cable...). If you send several packets with no delay between them, they may be coalesced into a larger packet. The networking server is too simple to handle this (and fixing it is a very low priority), but you can disable the coalescing behavior on the sender-side by setting the socket option `TCP_NODELAY`.*

### Connecting:
The IP address of the TCP server will be the same as the address the UDP server as spammed out on port 30000. Connect to the TCP server, using port `34933` for fixed-size messages, or port `33546` for 0-terminated messages. 

The server will send you a welcome-message when you connect, and after that it will echo anything you say back to you (as long as your message ends with `'\0'`). Try sending and receiving a few messages.

### Accepting connections:
Tell the server to connect back to you, by sending a message of the form `Connect to: #.#.#.#:#\0` (IP of your machine and port you are listening to). You can find your own address by running `ifconfig` in the terminal - the first three bytes should be the same as the server's IP.
 
This new connection will behave the same way on the server-side, so you can send messages and receive echoes in the same way as before. You can even have it "Connect to" recursively (but please be nice... And no, the server will refuse requests to connect to itself).


2.1: Network design
-------------------

Before proceeding with any project-related networking code, think about how you would solve these problems, and what you need in order to solve them.

 - Guarantees about elevators:
   - What should happen if one of the nodes loses its network connection?
   - What should happen if one of the nodes loses power for a brief moment?
   - What should happen if some unforeseen event causes the elevator to never reach its destination, but communication remains intact?
   
 - Guarantees about orders:
   - Do all your nodes need to "agree" on a call for it to be accepted? In that case, how is a faulty node handled? 
   - How can you be sure that a remote node "agrees" on an call?
   - How do you handle losing packets between the nodes?
   - Do you share the entire state of the current calls, or just the changes as they occur?
     - For either one: What should happen when an elevator re-joins after having been offline?

*Pencil and paper is encouraged! Drawing a diagram/graph of the message pathways between nodes (elevators) will aid in visualizing complexity. Drawing the order of messages through time will let you more easily see what happens when communication fails.*
     
 - Topology:
   - What kind of network topology do you want to implement? Peer to peer? Master slave? Circle? Something else?
   - In the case of a master-slave configuration: Do you have only one program, or two (a "master" executable and a "slave")?
     - How do you handle a master node disconnecting?
     - Is a slave becoming a master a part of the network module?
   - In the case of a peer-to-peer configuration:
     - Who decides the order assignment?
     - What happens if someone presses the same button on two panels at once? Is this even a problem?
   - Remember that you only have three machines available, no outside always-online fourth machine is permitted.
     
 - Technical implementation and module boundary:
   - Protocols: TCP, UDP, or something else?
      - If you are using TCP: How do you know who connects to who?
        - Do you need an initialization phase to set up all the connections?
      - If you are using UDP broadcast: How do you differentiate between messages from different nodes?
      - If you are using a library or language feature to do the heavy lifting - what is it, and does it satisfy your needs?
   - Do you want to build the necessary reliability into the module, or handle that at a higher level?
   - Is detection (and handling) of things like lost messages or lost nodes a part of the network module?
   - How will you pack and unpack (serialize) data?
     - Do you use structs, classes, tuples, lists, ...?
     - JSON, XML, plain strings, or just plain memcpy?
     - Is serialization a part of the network module?



2.2 Network modules
-------------------

Pre-made network modules for a couple programming languages [can be found among the TTK4145 respositories](https://github.com/TTK4145?q=network). These are typically UDP broadcast only. Use whatever you find useful - extend, modify, or delete as you see fit.
For Erlang or Elixir, look at [Distributed Erlang](http://erlang.org/doc/reference_manual/distributed.html)


3.1 SSH and SCP
---------------

In order to test networking (and the project) on the lab, you will find it useful to run your code from multiple machines at once. The best way to do this is to work on one machine, transfer your files, and be logged in remotely in other terminal sessions. Remember to be nice the people sitting at that computer (don't mess with their files, and so on), and try to avoid using the same ports as them.

 - Logging in:
   - `ssh username@#.#.#.#` where #.#.#.# is the remote IP
   - You can combine this with `sshpass` to avoid typing the password each time:  
     `sshpass -p password ssh student@10.100.23.#`
 - Copying files between machines:
   - `scp source destination`, with optional flag `-r` for recursive copy (folders)
   - Examples:
     - Copying files *to* remote: `scp -r fileOrFolderAtThisMachine username@#.#.#.#:fileOrFolderAtOtherMachine`
     - Copying files *from* remote: `scp -r username@#.#.#.#:fileOrFolderAtOtherMachine fileOrFolderAtThisMachine`
   - Again, this can be combined with `sshpass`.
    
*If you are scripting something to automate any part of this process, remember to **not** include the login password in any files you upload to GitHub (or anywhere else for that matter). For example, you can use `sshpass -f passwordfile` and add passwordfile to your `.gitignore`, or you could just pass the password as a command-line option to your custom build-and-deploy script*.

You can also log in to these machines from outside the lab, though you will have to [install NTNU's VPN](https://i.ntnu.no/wiki/-/wiki/norsk/installere+vpn).

By combining remote logins with [the simulator](https://github.com/TTK4145/Simulator-v2) (which is pre-installed on the lab machines), you could feasibly conduct all project development and testing from outside the lab. 

For this reason we also encourage you to not turn off the machines after a lab session, as someone else may be logged in remotely.

3.2 Network impairment testing
------------------------------

A custom network impairment program `netimpair` is installed on the machines on the lab.

*[The program can also be downloaded from the Project-resources respository](https://github.com/TTK4145/Project-resources/releases/latest)*

*If you are on Windows, you can instead use [Clumsy](https://github.com/jagt/clumsy)*

Try running `netimpair` alongside either your networking test code from earlier in the exercise, or any demo code from another networking module. For example, running your UDP receiver from the very first section alongside `sudo netimpair -p 30000 --loss 50`, you should see that instead of receiving an IP announcement exactly once a second, half of them will be dropped.

Note: you *can* emulate disconnections with `--loss 100`, but you will experience somewhat different error behaviors in networking code from this and physically disconnecting the network cable.



