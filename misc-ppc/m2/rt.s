#include "textflag.h"

// M2 用の偽 runtime: R0=0、g(R30) を偽の g に向けて main.main を呼び、main.out を書き出して終了
TEXT _rt0_ppc_linux(SB), NOSPLIT|NOFRAME, $0
	MOVW	$0, R0
	MOVD	$runtime·fakeg(SB), g
	ADD	$-64, R1
	BL	main·main(SB)
	MOVW	$4, R0			// SYS_write
	MOVW	$1, R3
	MOVD	$main·out(SB), R4
	MOVD	$main·n(SB), R5
	MOVWZ	0(R5), R5
	SYSCALL
	MOVW	$234, R0		// SYS_exit_group
	MOVW	$0, R3
	SYSCALL
	RET

// stackguard0(+8) = 0 なので Go のスタックチェックは常に通る
GLOBL runtime·fakeg(SB), NOPTR, $64

TEXT runtime·mapinitnoop(SB), NOSPLIT|NOFRAME, $0
	RET

TEXT runtime·morestack_noctxt(SB), NOSPLIT|NOFRAME, $0
	RET

TEXT runtime·panicdivide(SB), NOSPLIT|NOFRAME, $0
	RET

TEXT runtime·panicBounds(SB), NOSPLIT|NOFRAME, $0
	RET
