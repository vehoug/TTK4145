
with Ada.Text_IO; use Ada.Text_IO;
with Ada.Containers; use Ada.Containers;
with Ada.Containers.Vectors;


procedure protectobj is

    tick : Float := 0.033;
    package IntVec is new Ada.Containers.Vectors
        (Index_Type  => Natural,
        Element_Type => Integer);
    use IntVec;
    package Integer_IO is new Ada.Text_IO.Integer_IO (Integer);

    -- --- RESOURCE OBJECT --- --
    -- You will finish implementing allocateLow, allocateHigh, and deallocate.
    -- Note: There are no checks that the final execution order is correct. You will have to check this yourself.
    -- Hints:
    --  - Use `entryName'Count` to get the number of tasks waiting on an entry. This can be used in the guard of another entry.
    --  - Equality checks are done with `=`, instead of `==`. Assignment is always done with `:=`.
    -----------------------
    protected type Resource is
        entry allocateHigh(val: out IntVec.Vector);
        entry allocateLow(val: out IntVec.Vector);
        procedure deallocate(val: IntVec.Vector);
    private
        value: IntVec.Vector;
        busy: Boolean := False;
    end Resource;
    protected body Resource is
    
        entry allocateLow(val: out IntVec.Vector) when True is
        begin
            --Put_Line("allocateLow");
            val := value;
        end allocateLow;
    
        entry allocateHigh(val: out IntVec.Vector) when True is
        begin
            --Put_Line("allocateHigh");
            val := value;
        end allocateHigh;

        procedure deallocate(val: IntVec.Vector) is
        begin
            --Put_Line("deallocate");
            value := val;
        end deallocate;

    end Resource;



    type ExecutionState is (none, waiting, executing, done);
    type ExecutionStateArrT is array (0..9) of ExecutionState;
    executionStates: ExecutionStateArrT := (others => none);


    task type resourceUser(
        id:         Integer;
        priority:   Integer; 
        release:    Integer; 
        execute:    Integer; 
        r:          access Resource
    );
        value: IntVec.Vector;
    task body resourceUser is
    begin
        delay Duration(tick * Float(release));
        
        executionStates(id) := waiting;
        if priority = 1 then
            r.allocateHigh(value);
        else
            r.allocateLow(value);
        end if;
        
        executionStates(id) := executing;
        
        delay Duration(tick * Float(execute));
        value.Append(id);
        r.deallocate(value);
        
        executionStates(id) := done;
    end resourceUser;





    task type executionLogger;
        t : Integer := 0;
    task body executionLogger is
    begin
        delay Duration(tick/2.0);

        Put("  id:");
        for i in 0..executionStates'length-1 loop
            Integer_IO.Put(i, Width => 3);
        end loop;
        Put_Line("");

        loop
            Integer_IO.Put(t, Width => 4);
            Put(" : ");
            for state of executionStates loop
                case state is
                    when none =>
                        Put(" ");
                    when waiting =>
                        Put("|");
                    when executing =>
                        Put("#");
                    when done =>
                        Put("^");
                        state := none;
                end case;
                if t rem 5 = 0 then
                    Put("--");
                else
                    Put("  ");
                end if;
            end loop;
            Put_Line("");
            t := t+1;
            delay Duration(tick);
        end loop;
    end executionLogger;

    
    r:              aliased Resource;
    logger:         executionLogger;
    executionOrder: IntVec.Vector;
    
begin
    Put_Line("started");
    
    declare
    
        user00: resourceUser(0, 0, 1,  1, r'Access);
        user01: resourceUser(1, 0, 3,  1, r'Access);
        user02: resourceUser(2, 1, 5,  1, r'Access);
        
        user03: resourceUser(0, 1, 10, 2, r'Access);
        user04: resourceUser(1, 0, 11, 1, r'Access);
        user05: resourceUser(2, 1, 11, 1, r'Access);
        user06: resourceUser(3, 0, 11, 1, r'Access);
        user07: resourceUser(4, 1, 11, 1, r'Access);
        user08: resourceUser(5, 0, 11, 1, r'Access);
        user09: resourceUser(6, 1, 11, 1, r'Access);
        user10: resourceUser(7, 0, 11, 1, r'Access);
        user11: resourceUser(8, 1, 11, 1, r'Access);
        
        user12: resourceUser(0, 1, 25, 3, r'Access);
        user13: resourceUser(6, 0, 26, 2, r'Access);
        user14: resourceUser(7, 0, 26, 2, r'Access);
        user15: resourceUser(1, 1, 26, 2, r'Access);
        user16: resourceUser(2, 1, 27, 2, r'Access);
        user17: resourceUser(3, 1, 28, 2, r'Access);
        user18: resourceUser(4, 1, 29, 2, r'Access);
        user19: resourceUser(5, 1, 30, 2, r'Access);
    begin
        null;
    end;
    
    delay Duration(tick * 2.0);    
    abort logger;
    
    r.allocateHigh(executionOrder);
    
    Put_Line("Execution order: ");
    for idx in executionOrder.Iterate loop
        Put(Integer'Image(executionOrder(idx)));
    end loop;
    Put_Line("");

end protectobj;