package rsync

import "testing"

func TestBuildArgs(t *testing.T) {
	args := BuildArgs(RunRequest{
		Source:      "src/",
		Destination: "dest/",
		Archive:     true,
		Delete:      true,
		Partial:     true,
		Includes:    []string{"*.go"},
		Excludes:    []string{"tmp"},
	})

	want := []string{"-a", "--delete", "--partial", "--include=*.go", "--exclude=tmp", "src/", "dest/"}
	if len(args) != len(want) {
		t.Fatalf("len(args)=%d want=%d args=%v", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d]=%q want %q", i, args[i], want[i])
		}
	}
}
