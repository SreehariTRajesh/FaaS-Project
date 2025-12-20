#define _POSIX_C_SOURCE 199309L
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdint.h>
#include <time.h>
#include <unistd.h>

// Configuration
#define MATRIX_SIZE 256
#define PRIME_LIMIT 100000
#define PI_ITERATIONS 10000000
#define HASH_ITERATIONS 1000000

void multiply(double **A, double **B, double **C, int size)
{
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    for (int i = 0; i < size; i++)
    {
        for (int j = 0; j < size; j++)
        {
            for (int k = 0; k < size; k++)
            {
                C[i][j] += A[i][k] * B[k][j];
            }
        }
    }
    clock_gettime(CLOCK_MONOTONIC, &end);

    return;
}

int main(int argc, char *argv[])
{
    double **A, **B, **C;
    int size = 256;
    A = malloc(size * sizeof(double *));
    B = malloc(size * sizeof(double *));
    C = malloc(size * sizeof(double *));

    for (int i = 0; i < size; ++i)
    {
        A[i] = malloc(size * sizeof(double));
        B[i] = malloc(size * sizeof(double));
        C[i] = malloc(size * sizeof(double));
        for (int j = 0; j < size; ++j)
        {
            A[i][j] = (double)rand() / RAND_MAX;
            B[i][j] = (double)rand() / RAND_MAX;
            C[i][j] = 0.0;
        }
    }
    multiply(A, B, C, 256);
    for (int i = 0; i < size; i++)
    {
        free(A[i]);
        free(B[i]);
        free(C[i]);
    }
    free(A);
    free(B);
    free(C);
    return 0;
}
