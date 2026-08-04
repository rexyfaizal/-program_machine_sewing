package repository

import (
	"context"
	"testing"
	"time"

	"backend_machine/config"
)

func TestGetRuntimeSecMatchesNativeSample(t *testing.T) {
	db, err := config.ConnectDB()
	if err != nil {
		t.Skipf("skip: tidak bisa konek DB: %v", err)
	}
	defer db.Close()

	repo := New(db)

	// Samakan dengan GetMachineProductivity: time.Parse → UTC midnight.
	day, err := time.Parse("2006-01-02", "2026-08-04")
	if err != nil {
		t.Fatal(err)
	}

	sec, err := repo.GetRuntimeSec(
		context.Background(),
		"21211Jt6sbk819cP",
		"0",
		day,
		day.AddDate(0, 0, 1),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Power On = %dh %dm %ds (total=%d)", sec/3600, (sec%3600)/60, sec%60, sec)

	// Native app: 01:19:43 = 4783 detik. Izinkan toleransi kecil.
	expected := int64(4783)
	diff := sec - expected
	if diff < 0 {
		diff = -diff
	}

	if diff > 5 {
		t.Fatalf("Power On jauh dari native: got=%d expected~=%d diff=%d", sec, expected, diff)
	}
}
