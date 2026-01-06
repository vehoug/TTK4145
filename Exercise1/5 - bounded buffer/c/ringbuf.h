#pragma once

struct RingBuffer {
    int* buffer;
    int capacity;
    
    int insertIdx;
    int removeIdx;
    int length;
};


struct  RingBuffer* rb_new(int size);
void    rb_destroy(struct RingBuffer* rb);

void    rb_push(struct RingBuffer* rb, int val);
int     rb_pop(struct RingBuffer* rb);

void    rb_print(const struct RingBuffer* const rb);


