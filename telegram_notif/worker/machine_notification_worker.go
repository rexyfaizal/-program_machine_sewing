package worker

import (
	"context"
	"log"
	"time"
)

type MachineNotificationProcessor interface {
	ProcessPending(
		ctx context.Context,
	) error
}

type MachineNotificationWorker struct {
	processor MachineNotificationProcessor
	interval  time.Duration
}

func NewMachineNotificationWorker(
	processor MachineNotificationProcessor,
	interval time.Duration,
) *MachineNotificationWorker {
	if interval <= 0 {
		interval = 5 * time.Second
	}

	return &MachineNotificationWorker{
		processor: processor,
		interval:  interval,
	}
}

func (w *MachineNotificationWorker) Start(
	ctx context.Context,
) {
	log.Printf(
		"Worker notifikasi operator aktif. Interval=%s",
		w.interval,
	)

	log.Println(
		"MACHINE_BROKEN → seluruh mekanik aktif.",
	)

	log.Println(
		"WAIT_HANCA → SPV berdasarkan branch dan line.",
	)

	w.process(ctx)

	ticker := time.NewTicker(
		w.interval,
	)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(
				"Worker notifikasi operator berhenti.",
			)
			return

		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *MachineNotificationWorker) process(
	ctx context.Context,
) {
	requestCtx, cancel := context.WithTimeout(
		ctx,
		45*time.Second,
	)
	defer cancel()

	if err := w.processor.ProcessPending(
		requestCtx,
	); err != nil {
		if ctx.Err() != nil {
			return
		}

		log.Printf(
			"Gagal memproses notifikasi operator: %v",
			err,
		)
	}
}
