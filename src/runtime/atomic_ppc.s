// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// The 750CL has no lwsync, so use a full sync.
TEXT ·publicationBarrier(SB),NOSPLIT|NOFRAME,$0-0
	SYNC
	RET
