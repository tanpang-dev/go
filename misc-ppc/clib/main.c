/* c-archive / c-shared の呼び出し側。C のスレッドからも Go を呼ぶ。 */
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "lib.h"

static void *worker(void *arg) {
	long *r = arg;
	*r = GoWork(20);
	return NULL;
}

int main(void) {
	int ok = 1;
	ok &= GoAdd(40, 2) == 42;
	char *u = GoUpper("wii");
	ok &= strcmp(u, "WII") == 0;
	free(u);
	pthread_t t[4];
	long r[4];
	for (int i = 0; i < 4; i++) pthread_create(&t[i], NULL, worker, &r[i]);
	for (int i = 0; i < 4; i++) { pthread_join(t[i], NULL); ok &= r[i] == 20000; }
	printf("%s\n", ok ? "PASS" : "FAIL");
	return !ok;
}
