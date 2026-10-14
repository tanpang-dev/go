/* dlopen で c-shared の Go ライブラリを後から読み込む。 */
#include <dlfcn.h>
#include <stdio.h>

int main(int argc, char **argv) {
	void *h = dlopen(argc > 1 ? argv[1] : "./libgo.so", RTLD_NOW);
	if (!h) { printf("FAIL dlopen: %s\n", dlerror()); return 1; }
	int (*add)(int, int) = (int (*)(int, int))dlsym(h, "GoAdd");
	long (*work)(int) = (long (*)(int))dlsym(h, "GoWork");
	int ok = add && work && add(1, 2) == 3 && work(10) == 10000;
	printf("%s\n", ok ? "PASS" : "FAIL");
	return !ok;
}
