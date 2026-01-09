Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> *Concurrency refers to a system's ability to perform multiple tasks through simultaneous execution or context-switching, sharing resources and managing interactions. Parallelism is about executing several tasks independently on multiple CPU cores.*

What is the difference between a *race condition* and a *data race*? 
> *A **race condition** is the condition of a system where the system's behavior is dependent on the sequence or timing of other uncontrollable events, leading to unexpected results. A data race, on the other hand, occurs when two instructions from different threads access the same memory location, at least one of these accesses is a write and there is no synchronization that is mandating any particular order among these accesses.* 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> *A scheduler decides which task gets to use the CPU at any given moment. It uses context switching to achieve this, which works by pausing a running task, saving its current state, and swapping in the state of a different task from a waiting list.* 


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> *We use multiple threads to avoid a program from being blocked by a single slow task. Multiple threads allow a program to do something else while waiting for slow external events. Additionally, threads allow a program to finish heavy work faster by using more hardware because we can divide the problem into smaller parts and solve them in parallel or concurrently.*

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *They are lightweight threads managed by the programming language runtime rather than the OS kernel. They avoid the high memory overhead and slow context-switching costs of OS threads, allowing a single program to run many concurrent tasks efficiently.*

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> *Both. It makes life easier by providing a good way to model real-world systems with several simultaneous interactions and preventing programs from freezing during slow tasks. However, it also makes life harder by introducing bugs like race conditions, data races and deadlocks, making the program more difficult to test and debug.*

What do you think is best - *shared variables* or *message passing*?
> *It depends on the use case. Shared variables seems better for performance critical systems with less resources like memory and processor power, but message passing seems better for the opposite where our program is more complex and we have more resources available for our program.*


