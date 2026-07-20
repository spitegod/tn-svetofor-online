package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpdelivery "tn/backend/internal/delivery/http"

	"tn/backend/internal/config"
	"tn/backend/internal/db"
	"tn/backend/internal/repository"
	"tn/backend/internal/service"
)

func Run(cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := db.Migrate(ctx, database); err != nil {
		return err
	}

	classificationRepo := repository.NewClassificationRepository(database)
	classificationService := service.NewClassificationService(classificationRepo)
	systemCatalogRepo := repository.NewSystemCatalogRepository(database)
	systemCatalogService := service.NewSystemCatalogService(systemCatalogRepo)
	systemDocumentRepo := repository.NewSystemDocumentRepository(database)
	systemDocumentService := service.NewSystemDocumentService(systemDocumentRepo)
	navParserService := service.NewNavParserService(systemCatalogRepo)
	schedulerCtx, stopScheduler := context.WithCancel(context.Background())
	defer stopScheduler()
	go navParserService.RunScheduler(schedulerCtx)
	orderRepo := repository.NewOrderRepository(database)
	orderService := service.NewOrderService(orderRepo, classificationService, systemCatalogService)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpdelivery.NewRouter(classificationService, systemCatalogService, systemDocumentService, orderService, navParserService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("backend started on %s", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(ctx)
	}
}
