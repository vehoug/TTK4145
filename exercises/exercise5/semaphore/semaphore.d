
import std.algorithm, std.concurrency, std.format, std.range, std.stdio, std.traits;
import core.thread, core.sync.semaphore, core.sync.mutex, core.sync.condition;

immutable Duration tick = 33.msecs;

// --- RESOURCE CLASS --- //
/* 
You will implement the functionality for `allocate` and `deallocate`.

Documentation:
Semaphore:
    https://dlang.org/phobos/core_sync_semaphore.html
    You only need to know:
        wait()
        notify()    (same as "signal")
    (Note that there is no `getValue()` function!)
*/
class Resource(T) {
    private {
        T               value;
        Semaphore       mtx;
        Semaphore[2]    sems;
        int[2]          numWaiting;
        bool            busy;
    }
    
    this(){
        mtx = new Semaphore(1);
        foreach(ref sem; sems){
            sem = new Semaphore(0);
        }
    }
    
    T allocate(int priority){
        return value;
    }
    
    void deallocate(T v){
        value = v;
    }
}





void main(){

    // Resource type is `int[]`. Each user appends its own id to the back of the list.
    auto resource = new Resource!(int[])();

    executionStates = new ExecutionState[](10);
    
    auto cfgs = [
        ResourceUserConfig(0, 0, 1, 1),
        ResourceUserConfig(1, 0, 3, 1),
        ResourceUserConfig(2, 1, 5, 1),
        
        ResourceUserConfig(0, 1, 10, 2),
        ResourceUserConfig(1, 0, 11, 1),
        ResourceUserConfig(2, 1, 11, 1),
        ResourceUserConfig(3, 0, 11, 1),
        ResourceUserConfig(4, 1, 11, 1),
        ResourceUserConfig(5, 0, 11, 1),
        ResourceUserConfig(6, 1, 11, 1),
        ResourceUserConfig(7, 0, 11, 1),
        ResourceUserConfig(8, 1, 11, 1),
        
        ResourceUserConfig(0, 1, 25, 3),
        ResourceUserConfig(6, 0, 26, 2),
        ResourceUserConfig(7, 0, 26, 2),
        ResourceUserConfig(1, 1, 26, 2),
        ResourceUserConfig(2, 1, 27, 2),
        ResourceUserConfig(3, 1, 28, 2),
        ResourceUserConfig(4, 1, 29, 2),
        ResourceUserConfig(5, 1, 30, 2),
    ];
    
    spawn(&executionLogger);
    foreach(cfg; cfgs){
        spawnLinked(&resourceUser, cfg, cast(shared)resource);
    }
    foreach(_; 0..cfgs.length){
        receive(
            (LinkTerminated lt){
            }
        );
    }
    Thread.sleep(tick*2);
    
    
    auto val = resource.allocate(0);
    
    assert(val.length == cfgs.length,
        "Test failed: Did not run all users once");
    assert(val[0..3] == [0, 1, 2],
        format("Test 1 failed: Did not run users in ascending order, instead ran %s", val[0..3]));
    
    assert(val[3] == 0,
        format("Test 2 failed: Did not run initial (high priority) user, instead ran %s", val[3]));
    assert(val[4..8].all!("(a & 1) == 0"),
        format("Test 2 failed: Did not run high priority (even id) users first, instead ran %s", val[4..8]));
    assert(val[8..12].all!("a & 1"),
        format("Test 2 failed: Did not run low priority (odd id) users last, instead ran %s", val[8..12]));
    
    assert(val[12] == 0,
        format("Test 3 failed: Did not run initial (high priority) user, instead ran %s", val[12]));
    assert(val[13..18].all!("a >= 1") && val[13..18].all!("a <= 5"),
        format("Test 3 failed: Did not run high priority users first, instead ran %s", val[13..18]));
    assert(val[18..20].all!("a >= 6") && val[18..20].all!("a <= 7"),
        format("Test 3 failed: Did not run low priority users last, instead ran %s", val[18..20]));
    writeln("All tests pass");
}



// --- RESOURCE USERS -- //

struct ResourceUserConfig {
    int     id;
    int     priority;
    int     release;
    int     execute;
}

void resourceUser(ResourceUserConfig cfg, shared Resource!(int[]) r){
    Thread.getThis.isDaemon = true;    
    auto resource = cast(Resource!(int[]))r;
    
    Thread.sleep(cfg.release * tick);
    
    executionStates[cfg.id] = ExecutionState.waiting;
    auto val = resource.allocate(cfg.priority);
    
    executionStates[cfg.id] = ExecutionState.executing;
    
    Thread.sleep(cfg.execute * tick);
    val ~= cfg.id;
    resource.deallocate(val);
    
    executionStates[cfg.id] = ExecutionState.done;
}




// --- EXECUTION LOGGING --- //

version(Windows){
    enum ExecutionState : char {
        none        = ' ',
        waiting     = cast(char)177,
        executing   = cast(char)178,
        done        = cast(char)223,
    }
    enum Grid : char {
        none        = ' ',
        horizontal  = cast(char)196,
    }
} else {
    enum ExecutionState : wchar {
        none        = ' ',
        waiting     = '\u2592',
        executing   = '\u2593',
        done        = '\u2580',
    }
    enum Grid : wchar {
        none        = ' ',
        horizontal  = '\u2500',
    }
}

__gshared ExecutionState[] executionStates;

void executionLogger(){
    Thread.getThis.isDaemon = true;
    Thread.sleep(tick/2);
    
    writefln("  id:%(%3d%)", iota(0, executionStates.length));
    
    auto t = 0;
    while(true){
        writef("%04d : " , t);
        foreach(id, ref state; executionStates){
            auto grid = (t % 5 == 0) ? Grid.horizontal : Grid.none;
            writef("%c%c%c", cast(OriginalType!ExecutionState)state, grid, grid);
            if(state == ExecutionState.done){
                state = ExecutionState.none;
            }
        }
        writeln;
        t++;
        Thread.sleep(tick);
    }
}