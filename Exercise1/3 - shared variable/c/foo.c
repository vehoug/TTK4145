// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>

int i = 0;

// Note the return type: void*
// We want to protext the shared variable i, so mutexes are the best choice.
// We only want one thread to operate on the variable i at a time.

pthread_mutex_t lock;

// Lock the mutex before performing operations on i, and unlock after.
void* incrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++){
        pthread_mutex_lock(&lock);
        i++;
        pthread_mutex_unlock(&lock);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    for (int j = 0; j < 1000000; j++){
        pthread_mutex_lock(&lock);
        i--;
        pthread_mutex_unlock(&lock);
    }
    return NULL;
}


int main(){
    if (pthread_mutex_init(&lock, NULL) != 0){
        printf("\nMutex initialization failed\n");
        return EXIT_FAILURE;
    }

    pthread_t* thread1;
    pthread_create(&thread1, NULL, incrementingThreadFunction, NULL);

    pthread_t* thread2;
    pthread_create(&thread2, NULL, decrementingThreadFunction, NULL);
    
    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);
    
    printf("The magic number is: %d\n", i);

    pthread_mutex_destroy(&lock);

    return 0;
}
