package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/KazikovAP/backend_course_start/hw2/internal/coingecko"
	"github.com/KazikovAP/backend_course_start/hw2/internal/handler"
	"github.com/KazikovAP/backend_course_start/hw2/internal/repository"
	"github.com/KazikovAP/backend_course_start/hw2/internal/service"
)

func main() {
	cgClient := coingecko.NewClient()
	log.Println("Loading CoinGecko coin list...")
	if err := cgClient.WarmCache(); err != nil {
		log.Printf("Warning: %v — will use offline fallback", err)
	}

	userRepo := repository.NewUserMemoryRepository()
	cryptoRepo := repository.NewCryptoMemoryRepository()

	jwtSecret := []byte("secret_key")
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	cryptoSvc := service.NewCryptoService(cryptoRepo, cgClient)
	schedulerSvc := service.NewSchedulerService(cryptoRepo, cgClient)

	authH := handler.NewAuthHandler(authSvc)
	cryptoH := handler.NewCryptoHandler(cryptoSvc)
	scheduleH := handler.NewScheduleHandler(schedulerSvc)

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", authH.Register)
	mux.HandleFunc("/auth/login", authH.Login)

	auth := func(fn http.HandlerFunc) http.HandlerFunc {
		return handler.AuthMiddleware(authSvc, fn)
	}

	mux.HandleFunc("/crypto", auth(cryptoH.Collection))
	mux.HandleFunc("/crypto/", auth(cryptoH.Item))
	mux.HandleFunc("/schedule", auth(scheduleH.Schedule))
	mux.HandleFunc("/schedule/trigger", auth(scheduleH.Trigger))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go schedulerSvc.Run(ctx)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
