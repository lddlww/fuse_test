package chown

import (
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDangerousHostPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Current paths are supported only by Linux")
	}

	tests := []struct {
		Path             string
		Expected         bool
		ExpectError      bool
		ExpectedErrorMsg string
	}{
		{
			"/tmp",
			true,
			false,
			"",
		},
		{
			t.TempDir(), // Create a temp dir that is not dangerous
			false,
			false,
			"",
		},
		{
			"/doesnotexist",
			false,
			true,
			"no such file or directory",
		},
	}

	for _, test := range tests {
		result, err := DangerousHostPath(test.Path)
		if test.ExpectError {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.ExpectedErrorMsg)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		}
	}
}

func TestChangeHostPathOwnership(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Current paths are supported only by Linux")
	}

	// Create a temp dir that is not dangerous
	td := t.TempDir()

	// Get host path info
	f, err := os.Lstat(td)
	if err != nil {
		t.Fatal(err)
	}

	sys, ok := f.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("failed to cast stat to *syscall.Stat_t")
	}
	// Get current ownership
	currentUID := int(sys.Uid)
	currentGID := int(sys.Gid)

	tests := []struct {
		Path             string
		Recursive        bool
		UID              int
		GID              int
		ExpectError      bool
		ExpectedErrorMsg string
	}{
		{
			"/doesnotexist",
			false,
			0,
			0,
			true,
			"no such file or directory",
		},
		{
			"/tmp",
			false,
			0,
			0,
			true,
			"is not allowed",
		},
		{
			td,
			false,
			currentUID,
			currentGID,
			false,
			"",
		},
		{
			td,
			true,
			currentUID,
			currentGID,
			false,
			"",
		},
	}

	for _, test := range tests {
		err := ChangeHostPathOwnership(test.Path, test.Recursive, test.UID, test.GID)
		if test.ExpectError {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), test.ExpectedErrorMsg)
		} else {
			assert.NoError(t, err)
		}
	}
}
