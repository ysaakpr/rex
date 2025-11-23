package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/vyshakhp/utm-backend/internal/config"
	"github.com/vyshakhp/utm-backend/internal/jobs/tasks"
)

type Worker struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	scheduler *asynq.Scheduler
}

func NewWorker(cfg *config.Config, db *gorm.DB, logger *zap.Logger) (*Worker, error) {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: cfg.Asynq.Concurrency,
			Queues:      cfg.Asynq.Queues,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Error processing task %s: %v", task.Type(), err)
			}),
		},
	)

	mux := asynq.NewServeMux()

	// Register task handlers
	tenantInitHandler := tasks.NewTenantInitHandler(db, cfg)
	mux.HandleFunc(TypeTenantInitialization, tenantInitHandler.HandleTenantInitialization)

	invitationHandler := tasks.NewInvitationHandler(db, cfg)
	mux.HandleFunc(TypeUserInvitation, invitationHandler.HandleUserInvitation)

	// Initialize system user expiry task
	systemUserExpiryTask := tasks.NewSystemUserExpiryTask(db, logger)
	mux.HandleFunc(TypeSystemUserExpiry, systemUserExpiryTask.HandleSystemUserExpiry)

	// Initialize scheduler for periodic tasks
	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Logger: logger.Sugar(),
	})

	// Schedule system user expiry check (runs every hour)
	_, err := scheduler.Register(
		"@hourly",
		asynq.NewTask(TypeSystemUserExpiry, nil),
		asynq.Queue(QueueLow),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register periodic task: %w", err)
	}

	logger.Info("Scheduled periodic task: system user expiry check (hourly)")

	return &Worker{
		server:    server,
		mux:       mux,
		scheduler: scheduler,
	}, nil
}

func (w *Worker) Start() error {
	fmt.Println("Starting background worker...")

	// Start scheduler in a goroutine
	if w.scheduler != nil {
		go func() {
			if err := w.scheduler.Run(); err != nil {
				log.Printf("Scheduler error: %v", err)
			}
		}()
		fmt.Println("Scheduler started for periodic tasks")
	}

	// Start the worker server (blocking)
	if err := w.server.Run(w.mux); err != nil {
		return fmt.Errorf("could not start worker: %w", err)
	}
	return nil
}

func (w *Worker) Shutdown() {
	fmt.Println("Shutting down background worker...")

	// Shutdown scheduler first
	if w.scheduler != nil {
		w.scheduler.Shutdown()
		fmt.Println("Scheduler shut down")
	}

	// Shutdown server
	w.server.Shutdown()
	fmt.Println("Worker shut down")
}
