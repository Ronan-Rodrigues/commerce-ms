package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/config"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/client"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/crypto"
	infraPostgres "github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/database/postgres"
	infraHttp "github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/http"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/http/handler"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/infra/pubsub"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/usecase/financeiro"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/usecase/webhook"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("[PAYMENT-SERVICE] Iniciando na porta %s...", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Conexão PostgreSQL
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Não foi possível conectar ao Postgres agora: %v (suba o docker-compose)", err)
	} else {
		defer dbPool.Close()
		log.Println("[POSTGRES] Conectado ao schema payments com sucesso.")
	}

	// 2. Conexão Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()
	log.Println("[REDIS] Conectado com sucesso.")

	// 3. Repositórios e Clientes
	txRepo := infraPostgres.NewPostgresTransactionRepository(dbPool)
	publisher := pubsub.NewRedisPublisher(rdb)
	orderClient := client.NewOrderHTTPClient(cfg.OrderBaseURL)
	hmacValidator := crypto.NewSHA256HMACValidator()

	// 4. Casos de Uso
	webhookUC := webhook.NewProcessWebhookUseCase(txRepo, orderClient, publisher, hmacValidator, cfg.WebhookSecret)
	summaryUC := financeiro.NewFinancialSummaryUseCase(txRepo)

	// 5. Handlers e Router
	paymentHandler := handler.NewPaymentHandler(webhookUC, summaryUC, txRepo)
	router := infraHttp.NewRouter(paymentHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Início do Servidor com Graceful Shutdown
	go func() {
		log.Printf("[PAYMENT-SERVICE] Servidor pronto e escutando na porta :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no servidor HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[PAYMENT-SERVICE] Encerrando conexões com segurança...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Falha no shutdown: %v", err)
	}

	log.Println("[PAYMENT-SERVICE] Servidor finalizado.")
}
