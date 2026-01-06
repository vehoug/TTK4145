// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;

// Note the return type: void*
void* incrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++){
        i++;
    }
    return NULL;
}

void* decrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++){
        i--;
    }
    return NULL;
}


int main(){
    pthread_t* thread1;
    pthread_create(&thread1, NULL, incrementingThreadFunction, NULL);

    pthread_t* thread2;
    pthread_create(&thread2, NULL, decrementingThreadFunction, NULL);
    
    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);
    
    printf("The magic number is: %d\n", i);
    return 0;
}
