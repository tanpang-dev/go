#include "textflag.h"

// 32bit 化の確認用。ppc64 の書き方で書いて、32bit の命令になるかを見る
TEXT ·leaf(SB), NOSPLIT|NOFRAME, $0-8
	MOVD	x+0(FP), R3
	MOVD	$0x12345678, R4
	MOVD	$-1, R5
	ADD	R4, R3, R3
	MOVW	R3, R6
	MOVD	R3, ret+4(FP)
	CMP	R3, R4
	CMPU	R3, R4
	RET

TEXT ·nonleaf(SB), $16-0
	MOVD	$·leaf(SB), R3
	BL	·leaf(SB)
	MOVD	16(g), R4
	RET
