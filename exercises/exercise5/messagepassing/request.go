package main

import "fmt"
import "time"
import "sort"


const tick = time.Millisecond * 33


// --- RESOURCE ROUTINE --- //
/*
You will modify the resourceManager routine so that it... works...
(You know the drill by now...)

Note: The code in `main` does not contain checks that the final execution order is correct.

See how the request/reply/giveBack operations work in the `resourceUser` routine.

Hints:
 - Remember that you can do things outside the `select` too:
    ```Go
    for {
        select {
            cases...
        }
        other stuff...
    }
    ```
 - Use the priority queue (scroll down below `main()`. And remember the type-assertion thing:  
    ```Go
    request := queue.Front().(ResourceRequest)
    ```
*/

type Resource struct {
    value []int // Resource type is []int. Each user appends its own id when executing.
}

type ResourceRequest struct {
    id          int
    priority    int
    channel     chan Resource
}

func resourceManager(askFor chan ResourceRequest, giveBack chan Resource){

    res     := Resource{}
    //busy    := false
    //queue   := PriorityQueue{}

    for {
        select {
        case request := <-askFor:
            //fmt.Printf("[resource manager]: received request: %+v\n", request)
            request.channel <- res
        case res = <-giveBack:
            //fmt.Printf("[resource manager]: resource returned\n")
        }
    }
}


// --- RESOURCE USERS -- //

type ResourceUserConfig struct {
    id          int
    priority    int
    release     int
    execution   int
}

func resourceUser(cfg ResourceUserConfig, askFor chan ResourceRequest, giveBack chan Resource){

    replyChan := make(chan Resource)

    time.Sleep(time.Duration(cfg.release) * tick)

    executionStates[cfg.id] = waiting
    askFor <- ResourceRequest{cfg.id, cfg.priority, replyChan}
    res := <- replyChan

    executionStates[cfg.id] = executing

    time.Sleep(time.Duration(cfg.execution) * tick)
    res.value = append(res.value, cfg.id)
    giveBack <- res

    executionStates[cfg.id] = done
}


func main(){
    askFor      := make(chan ResourceRequest, 10)
    giveBack    := make(chan Resource)
    go resourceManager(askFor, giveBack)

    executionStates = make([]ExecutionState, 10)

    cfgs := []ResourceUserConfig{
        {0, 0, 1, 1},
        {1, 0, 3, 1},
        {2, 1, 5, 1},

        {0, 1, 10, 2},
        {1, 0, 11, 1},
        {2, 1, 11, 1},
        {3, 0, 11, 1},
        {4, 1, 11, 1},
        {5, 0, 11, 1},
        {6, 1, 11, 1},
        {7, 0, 11, 1},
        {8, 1, 11, 1},

        {0, 1, 25, 3},
        {6, 0, 26, 2},
        {7, 0, 26, 2},
        {1, 1, 26, 2},
        {2, 1, 27, 2},
        {3, 1, 28, 2},
        {4, 1, 29, 2},
        {5, 1, 30, 2},
    }

    go executionLogger()
    for _, cfg := range cfgs {
        go resourceUser(cfg, askFor, giveBack)
    }

    // (no way to join goroutines, hacking it with sleep)
    time.Sleep(time.Duration(45) * tick)

    resourceCh := make(chan Resource)
    askFor <- ResourceRequest{0, 1, resourceCh}
    executionOrder := <-resourceCh
    fmt.Println("Execution order:", executionOrder)
}



// --- PRIORITY QUEUE --- //
/*
Who needs type-safety anyway? Can take elements of any type, and also mix them...

Can take multiple elements for each priority (also of different types).
Ordering of same-priority elements is first-come-first-served.

You will have to use a "type assertion" to cast the interface{} to the correct type:
`first := queue.Front().(YourCustomType)`
*/
type PriorityQueue struct {
    queue []struct {
        val         interface{}
        priority    int
    }
}
func (pq *PriorityQueue) Insert(value interface{}, priority int){
    pq.queue = append(pq.queue, struct{val interface{}; priority int}{value, priority})
    sort.SliceStable(pq.queue, func(i, j int) bool {
        return pq.queue[i].priority > pq.queue[j].priority
    })
}
func (pq *PriorityQueue) Front() interface{} {
    return pq.queue[0].val
}
func (pq *PriorityQueue) PopFront(){
    pq.queue = pq.queue[1:]
}
func(pq *PriorityQueue) Empty() bool {
    return len(pq.queue) == 0
}



// --- EXECUTION LOGGING --- //

type ExecutionState rune
const (
    none    ExecutionState  = '\u0020'
    waiting                 = '\u2592'
    executing               = '\u2593'
    done                    = '\u2580'
)

var executionStates []ExecutionState

func executionLogger(){
    time.Sleep(tick/2)
    t := 0


    fmt.Printf("  id:")
    for id, _ := range executionStates {
        fmt.Printf("%3d", id)
    }
    fmt.Printf("\n")

    for {
        grid := ' ';
        if t % 5 == 0 {
            grid = '\u2500'
        }
        fmt.Printf("%04d : ", t)
        for id, state := range executionStates {
            fmt.Printf("%c%c%c", state, grid, grid)
            if state == done {
                executionStates[id] = none
            }
        }
        fmt.Printf("\n")
        t++
        time.Sleep(tick)
    }
}
