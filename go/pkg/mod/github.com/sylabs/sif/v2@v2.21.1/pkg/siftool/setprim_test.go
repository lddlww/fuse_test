// Copyright (c) 2021-2022, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package siftool

import (
	"testing"
)

func Test_command_getSetPrim(t *testing.T) {
	tests := []struct {
		name string
		opts commandOpts
	}{
		{
			name: "OK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &command{opts: tt.opts}

			cmd := c.getSetPrim()

			runCommand(t, cmd, []string{"1", makeTestSIF(t, true)}, nil)
		})
	}
}
