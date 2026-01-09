#include <stdio.h>
#include <stdlib.h>
#include <assert.h>


struct RingBuffer {
    int* buffer;
    int capacity;
    
    int insertIdx;
    int removeIdx;
    int length;
    
};


struct RingBuffer* rb_new(int size){
    struct RingBuffer* rb = malloc(sizeof(struct RingBuffer));
    
    rb->buffer       = malloc(sizeof(int) * size);
    rb->capacity     = size;
    rb->insertIdx    = 0;
    rb->removeIdx    = 0;    
    rb->length       = 0;
    
    return rb;    
}

void rb_print(const struct RingBuffer* const rb){
    //printf("b:%p c:%d, l:%d, i:%d, r:%d  ", rb->buffer, rb->capacity, rb->length, rb->insertIdx, rb->removeIdx);
    
    printf("[");
    int idx, elem;
    for(elem = 0, idx = rb->removeIdx;  elem < rb->length;  elem++, idx = (idx+1) % rb->capacity){
        printf("%d, ", rb->buffer[idx]);
    }
    
    printf("]\n");
}

void rb_push(struct RingBuffer* rb, int val){
    assert(rb->length < rb->capacity && "Bounds error: Attempted to push an element into a full buffer");
    rb->buffer[rb->insertIdx] = val;
    rb->insertIdx = (rb->insertIdx + 1) % rb->capacity;
    rb->length++;
}

int rb_pop(struct RingBuffer* rb){    
    assert(rb->length > 0 && "Bounds error: Attempted to pop an element from an empty buffer");
    int val = rb->buffer[rb->removeIdx];
    rb->removeIdx = (rb->removeIdx + 1) % rb->capacity;
    rb->length--;
    
    return val;
}

void rb_destroy(struct RingBuffer* rb){
    free(rb->buffer);
    free(rb);
}



