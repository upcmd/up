// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import (
	"fmt"
	"time"
)

var (
	P   = fmt.Println
	Pf  = fmt.Printf
	Sp  = fmt.Sprint
	Spf = fmt.Sprintf
)

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Sleep(mscnt int) {
	PfHiColor("sleeping %d milli seconds", mscnt)
	total := 0
	for i := 0; i < mscnt; i += 100 {
		Pf("%s", ".")
		total += 100
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(time.Duration(mscnt-total) * time.Millisecond)
	P()
}

