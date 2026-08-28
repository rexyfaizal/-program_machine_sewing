package utils

// ProcRuntimeToleranceSec = buffer mesin nyala di atas proses (5 menit).
// Hanya post-processing metrik; query dan aturan shift tidak diubah.
const ProcRuntimeToleranceSec int64 = 300

// ApplyProcRuntimeTolerance menaikkan mesin nyala jika proses lebih besar,
// sehingga mesin nyala = proses + toleransi (selalu lebih besar dari proses).
func ApplyProcRuntimeTolerance(runtimeSec, procSec int64) (adjustedRuntime, adjustedProc int64, applied bool) {
	adjustedRuntime = runtimeSec
	adjustedProc = procSec

	if procSec <= runtimeSec {
		return adjustedRuntime, adjustedProc, false
	}

	adjustedRuntime = procSec + ProcRuntimeToleranceSec
	return adjustedRuntime, adjustedProc, true
}
