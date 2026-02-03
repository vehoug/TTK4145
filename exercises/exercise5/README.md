Exercise 5 : A Synchronization Problem
======================================

In this exercise, we will solve the same synchronization problem in five different ways, using:

- Semaphores
- Condition Variables
- Protected Objects
- Message passing (2 ways)

The goal is not just to get better at solving problems with these primitives, but also to gain a deeper understanding of their relationships (in what ways the solutions with different primitives are similar or different) and their applicability (which ones are best suited for what purpose).

The first goal of this exercise is to get some hands-on experience with solving problems with each of these mechanisms. We often talk about these as the "C/Java/Ada/Go way of doing things", and there are different ways of thinking enabled (or required) by each of these. Translating a single problem between all of these will more clearly highlight the similarities and differences between them.

There is not always a specific link between the programming language and the synchronization mechanism. The only one of these that is bound to a language is Protected Objects with Ada (and to some extent also Channels with Go). But remember that all of these solutions run on the same operating system, on the same hardware, and with the same instruction set. The only thing a language or mechanism can provide is *sugar*, making some things *easier* at the cost of making other things harder (or sometimes impossible). Understanding this tradeoff is the second goal of this exercise.

For this exercise, we will in fact not use C or Java, but instead, use D for both semaphores and condition variables. It's one less language so support, unlike C it is portable across operating systems, and unlike Java, you already know how to use it (assuming you know C and a tiny bit of C++).

### The problem

The problem we will be solving is a single resource with multiple priority levels. Using mutexes we can protect a single shared resource, but we have no good way of deciding who gets the resource when it becomes available. In our case, we want to make sure it is always given to the task with the highest priority, so a simple mutex is not enough.

For this specific example, we will only use two priority levels (*high* and *low*). But keep in mind how the different solutions generalize to any number of priorities, as this might reflect on the power of the different synchronization primitives.

The pseudocode below implements a buggy solution to this problem, using semaphores. This code is adapted from Burns & Wellings Chapter 5.4 *Semaphores* (Specifically 5.4.7 *Semaphore programming using C/Real-time POSIX*). In the lectures, you will find it referred to as "A hard-to-find bug".

```C
// Initial values:
// M     = 1
// PS[2] = [0, 0]
// busy  = false

// priority: 1=high, 0=low
void allocate(int priority){
    Wait(M);
    if(busy){
        Signal(M);
        Wait(PS[priority]);
    }
    busy = true;
    Signal(M);
}

void deallocate(){
    Wait(M);
    busy = false;
    if(GetValue(PS[1]) < 0){
        Signal(PS[1]);
    } else if(GetValue(PS[0]) < 0){
        Signal(PS[0]);
    } else {
        Signal(M);
    }
}
```

To briefly narrativize this (buggy) solution: When allocating, the resource is either in use (`busy`) or not. If it is in use, we wait at the appropriate priority semaphore (`PS`), and the job of handing over the resource falls to whoever deallocates it. When deallocating, we first check for high-priority waiters, then low-priority, and then finally do nothing if there are neither. The operations manipulating the priority queues are protected with an outer mutex (`M`).

Before continuing - try to find the bug yourself.

### The bug

The bug comes from two interacting sources:

 1. In `allocate`, the outer mutex must be unlocked *before* we can wait in our queue.  
    Otherwise calling `wait` would block, and the outer mutex `M` would never be unlocked (which would prevent anyone else from running either function, blocking all execution).
 2. In `deallocate`, we want to check that there is someone waiting in the priority queues (with `GetValue`) before we `Signal` it.  
    Otherwise, we would be signaling queues with no one waiting, which would let multiple users allocate the resource at a later time.
 3. Combined, it is now possible that the value for one of the `PS` semaphores changes between `GetValue` and `Signal`, because **1.** allows for one task to run `deallocate` while another is between `Signal(M)` and `Wait(PS[...])` in `allocate`.

A proper (non-buggy) solution would have to fix either **1.** or **2.** (or both): The outer lock must not be released before modifications to internals (like the priority queues) are finished, or the getvalue and signal operations must be replaced with an atomic check-and-modify operation.

### The test

In order to check that your various solutions are correct, each starter code comes with a set of tasks that allocate and deallocate the resource in a specific order, and a logging mechanism to show the execution state of the tasks. The test consists of three sequences:

- **Releasing single users:**  
  *Test: Users take the resource, and give it back before anyone else shows up.*  
  Expected: Users execute in the order they are released (id 0, 1, then 2), and use the resource immediately.
- **Releasing multiple users:**  
  *Test: Multiple users of different priorities show up at the same time.*  
  Expected 1: High-priority users (even id number) execute before low-priority users (odd id).  
  Expected 2: No users execute with the resource at the same time.  
- **Releasing high-priority users while low-priority users are waiting:**  
  *Test: Multiple high-priority users show up over time.*  
  Expected: Low-priority users (6 & 7) wait until all high-priority users (1 through 5) are done.

The output consists of two parts: First, a (vertical) Gantt chart that shows which tasks are doing nothing (blank), waiting for the resource (light shade), executing with the resource (dark shade), or have just finished (filled upper-half square). Second, it displays the order in which the tasks used the resource. The internal data structure of the resource is an array of integers, and each resource user appends its own identifier to this list. Example output:

```text
  id:  0  1  2  3  4  5  6  7  8  9
0000 :  ── ── ── ── ── ── ── ── ── ──
0001 : ▓                             
0002 : ▀                             
0003 :    ▓                          
0004 :    ▀                          
0005 :  ── ──▓── ── ── ── ── ── ── ──
0006 :       ▀                       
0007 :                               
0008 :                               
0009 :                               
0010 : ▓── ── ── ── ── ── ── ── ── ──
0011 : ▓  ▒  ▒  ▒  ▒  ▒  ▒  ▒  ▒     
0012 : ▀  ▒  ▓  ▒  ▒  ▒  ▒  ▒  ▒     
0013 :    ▒  ▀  ▒  ▓  ▒  ▒  ▒  ▒     
0014 :    ▒     ▒  ▀  ▒  ▒  ▒  ▓     
0015 :  ──▒── ──▒── ──▒──▓──▒──▀── ──
0016 :    ▒     ▒     ▒  ▀  ▓        
0017 :    ▒     ▓     ▒     ▀        
0018 :    ▒     ▀     ▓              
0019 :    ▓           ▀              
0020 :  ──▀── ── ── ── ── ── ── ── ──
0021 :                               
0022 :                               
0023 :                               
0024 : ▓                             
0025 : ▓──▒── ── ── ── ──▒──▒── ── ──
0026 : ▓  ▒  ▒           ▒  ▒        
0027 : ▀  ▓  ▒  ▒        ▒  ▒        
0028 :    ▓  ▒  ▒  ▒     ▒  ▒        
0029 :    ▀  ▓  ▒  ▒  ▒  ▒  ▒        
0030 :  ── ──▓──▒──▒──▒──▒──▒── ── ──
0031 :       ▀  ▓  ▒  ▒  ▒  ▒        
0032 :          ▓  ▒  ▒  ▒  ▒        
0033 :          ▀  ▓  ▒  ▒  ▒        
0034 :             ▓  ▒  ▒  ▒        
0035 :  ── ── ── ──▀──▓──▒──▒── ── ──
0036 :                ▓  ▒  ▒        
0037 :                ▀  ▒  ▓        
0038 :                   ▒  ▓        
0039 :                   ▓  ▀        
0040 :  ── ── ── ── ── ──▓── ── ── ──
0041 :                   ▀           
0042 :                               
Execution order: [0, 1, 2, 0, 2, 4, 8, 6, 7, 3, 5, 1, 0, 1, 2, 3, 4, 5, 7, 6]
All tests pass
```

You should only have to modify the resource class/object/thread in the starter codes, the resource users and logging mechanisms are already completed.

*Note that the execution logging and the task release and execution times are triggered by sleeps, tied to the tick rate. Setting the tick rate too fast can make the logger output strange things, scramble the order in which tasks are released and make the tasks execute so fast that they don't get to block for each other.*

The part where you do the thing
-------------------------------

### Part 1: Semaphores

The buggy solution is very close to something that works. Instead of using `getValue` on the semaphore to get the amount of users waiting for that priority level, use a separate `int[2] numWaiting` variable. This will let you make sure all modifications to this variable are complete before you release the outer mutex `M` in allocate (since modifying the numbers in the queues is no longer tied to modifying the semaphore itself).

- You will find starter code in [the `semaphore` folder](/semaphore).
- You will need [a D compiler](https://dlang.org/download#dmd) (which should be installed on the lab PCs). Run the code with `rdmd semaphore.d`.
- Alternatively, you can use the [online editor and compiler](https://run.dlang.io/) and copy and paste the code back and forth.

### Part 2: Condition Variables

A condition variable is an extension to a mutex, that allows you to both wait like a semaphore *and* temporarily release the mutex – all in one operation. This gives the combined condition–mutex-pair four operations:

- `lock` and `unlock` for the mutex, which work like normal.
- `wait`, which blocks until notified, and also temporarily releases the mutex (the mutex is automatically re-locked by the time `wait` returns).
- `notify` and `notifyAll`, which releases one or all of the tasks `wait`ing (but since they have a mutex locked, they run *in turn*, not at the same time).

The standard pattern for using a condition variable is:

```text
functionA(){
    lock
    while(not our turn){
        wait
    }
    unlock
}

functionB(){
    lock
    make one available
    notifyAll
    unlock
}
```

The ability to temporarily release the mutex while waiting, lets another task do whatever is necessary to make the waiting task runnable again. The hard part is finding out what "not our turn" and "make (only?) one available" entails. This "condition" part of the condition variable is entirely up to you, the programmer.

In our case, a priority queue makes the most sense: `allocate` adds ourselves to the queue, and we wait while we are not the first one in the queue. `deallocate` pops us off the queue, making someone else the first element.

- You will find starter code in [the `conditionvariable` folder](/conditionvariable).
- Run the code with `rdmd condvar.d`.
- Again, you can use the [online editor and compiler](https://run.dlang.io/) and copy and paste the code back and forth.

---

Why is it called a "condition variable"? This appears to be mostly an artifact of history. C. A. R. Hoare proposed the mechanism in 1974, and in his paper [*Monitors: An Operating System Structuring Concept*](https://dl.acm.org/doi/pdf/10.1145/355620.361161), he writes: 

*Note that a condition "variable" is neither true nor false; indeed, it does not have any stored value accessible to the program. In practice, a condition variable will be represented by an (initially empty) queue of processes which are currently waiting on the condition; but this queue is invisible both to waiters and signallers.*

This means that the "condition" part is just "whatever the reason for waiting is" (which is entirely user-defined), and the "variable" part is the data structure for the hidden/backend/internal queuing mechanism. The user-facing reference to this variable (like the `core.sync.Condition` type) never got a better name.

Just to make this clear: Any variables the user checks when evaluating the condition are just *regular variables*, and should not be confused with the Condition Variable itself - despite being both variables, and part of a condition...

As for the relation to monitors: A monitor is the mutex plus condition variable combo package, where the locking (and subsequent unlocking) of the mutex is done automatically when entering (and exiting) the synchronized method call. In this way, the "standard pattern" mentioned above is enforced automatically.

### Part 3: Protected Objects

In Ada's tasking model, a *protected object* is an object that disallows simultaneous access, thus preventing data races. The main way of providing *conditional* access is with [entries and guards](https://learn.adacore.com/courses/intro-to-ada/chapters/tasking.html#entries). With conditional access, we can program things like `allocate(value: out Resource) when not busy`.

Note that the guards on entries cannot use the parameters in their conditions. This means we have to have separate entries for each priority, since having just a single entry `allocate(priority: Integer; value: out Resource)` would not be able to distinguish between high and low priority.

- You will find starter code in [the `protectedobject` folder](/protectedobject).
  - Note that this code does *not* have any built-in tests for the execution order! You will have to check this manually.
- You will need [an Ada compiler](https://www.adacore.com/download). Compile the code with `gnatmake protectobj.adb`.
- Alternatively, try an online compiler. In order of most to least promising:  
  - [OneCompiler](https://onecompiler.com/ada/3wtpw5fr4) (pre-loaded)  
  - [JDoodle](https://www.jdoodle.com/execute-ada-online/)
  - [CodingGround](https://www.tutorialspoint.com/compile_ada_online.php)  

### Part 4: Message Passing

With message passing, the resource must be sent to and from *something*, which means the resource "manager" would have to be a thread of its own. From here we have two main categories of solutions:

- Send a message to request the resource, containing a way for the resource manager to "call back" and give us the resource.  
  *This method does not rely on channels specifically and can be done with any kind of message passing.*
- Try to take the resource directly from a channel, where the resource manager uses `select`'s ability for send-cases to conditionally refuse to send.  
  *This method requires channels, select, and both receive and send cases.*

The "request" solution is in principle very similar to the one with condition variables. The resource manager takes all the requests it has received, sorts them in a priority queue, and then hands out the resource whenever it is available (and there is someone who wants it). The big difference is that instead of calling `allocate` as a function we send a resource-request-message on a channel, and instead of getting the resource as a return value we get it sent on a channel.

The "send case" solution is... different. Since we need multiple priority levels, we can't use just one channel, as there would be no way for the resource manager to know what priority we are. With multiple channels we will have to use select:

```Go
resource := Resource{}
for {
    select {
    case takeHigh<- resource:
        fmt.Printf("[resource manager]: resource taken (high)\n")
    case takeLow<- resource:
        fmt.Printf("[resource manager]: resource taken (low)\n")
    case resource = <-giveBack:
        fmt.Printf("[resource manager]: resource returned\n")
    }
}
```

The problem is that there is no way to prioritize the cases, as Go will [choose a random case](https://golang.org/ref/spec#Select_statements) if multiple are available. Since we need a priority select mechanism, we will have to hack it with the parts that are available, specifically `default`: Try sending to a high-priority user, then default back to waiting for either one.

- You will find starter code in [the `messagepassing` folder](/messagepassing). You should complete both variants.
- You will need [a Go compiler](https://golang.org/dl/). Run the code with `go run request.go` and `go run priorityselect.go`.
- Alternatively, you can use the online editor and compiler (pre-loaded with the starter code):  
  - [The request version](https://play.golang.org/p/Up7SvdkoHSE)  
  - [The priority select version](https://play.golang.org/p/Io7StFfUTk5)

*Note: "Resource Manager" is a highly mediocre name (suggestions for alternatives are welcome). Be sure to not confuse this with the "resource manager" from transactions. Naming things is one of the top two hardest things in programming, along with cache invalidation and off-by-one errors.*

### Part 5: Reflecting

*You do not have to answer every question in turn, as long as you address the contents somewhere.*

- Condition variables, Java monitors, and Ada protected objects are quite similar in what they do (temporarily yield execution so some other task can unblock us).
  - But in what ways do these mechanisms differ?

- Bugs in this kind of low-level synchronization can be hard to spot.
  - Which solutions are you most confident are correct?
  - Why, and what does this say about code quality?

- We operated only with two priority levels here, but it makes sense for this "kind" of priority resource to support more priorities.
  - How would you extend these solutions to N priorities? Is it even possible to do this elegantly?
  - What (if anything) does that say about code quality?

- In D's standard library, `getValue` for semaphores is not even exposed (probably because it is not portable – Windows semaphores don't have `getValue`, though you could hack it together with `ReleaseSemaphore()` and `WaitForSingleObject()`).
  - A leading question: Is using `getValue` *ever* appropriate?
  - Explain your intuition: What is it that makes `getValue` so dubious?

- Which one(s) of these different mechanisms do you prefer, both for this specific task and in general? (This is a matter of taste – there are no "right" answers here)
