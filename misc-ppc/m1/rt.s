#include "textflag.h"

// M1 用の偽 runtime: write(1, msg, len) して exit_group(0)
TEXT _rt0_ppc_linux(SB), NOSPLIT|NOFRAME, $0
	MOVW	$4, R0			// SYS_write
	MOVW	$1, R3
	MOVD	$runtime·msg(SB), R4
	MOVW	$25, R5
	SYSCALL
	MOVW	$234, R0		// SYS_exit_group
	MOVW	$0, R3
	SYSCALL
	RET

DATA runtime·msg+0(SB)/8, $"hello fr"
DATA runtime·msg+8(SB)/8, $"om Go/pp"
DATA runtime·msg+16(SB)/8, $"c on Wii"
DATA runtime·msg+24(SB)/1, $"\n"
GLOBL runtime·msg(SB), RODATA, $25

TEXT runtime·mapinitnoop(SB), NOSPLIT|NOFRAME, $0
	RET
