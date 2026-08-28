package utils

import "testing"

func TestApplyProcRuntimeTolerance(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		runtime       int64
		proc          int64
		wantRuntime   int64
		wantProc      int64
		wantApplied   bool
	}{
		{"proc sama runtime", 3600, 3600, 3600, 3600, false},
		{"proc lebih kecil", 3600, 3000, 3600, 3000, false},
		{"proc lebih besar 30 menit", 3600, 5400, 5700, 5400, true},
		{"proc lebih besar 35 menit", 3600, 5700, 6000, 5700, true},
		{"runtime nol proc ada", 0, 5400, 5700, 5400, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotRuntime, gotProc, gotApplied := ApplyProcRuntimeTolerance(tc.runtime, tc.proc)
			if gotRuntime != tc.wantRuntime || gotProc != tc.wantProc || gotApplied != tc.wantApplied {
				t.Fatalf("ApplyProcRuntimeTolerance(%d,%d) = (%d,%d,%v), want (%d,%d,%v)",
					tc.runtime, tc.proc, gotRuntime, gotProc, gotApplied,
					tc.wantRuntime, tc.wantProc, tc.wantApplied)
			}
		})
	}
}
