package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/config"
	infraRedis "github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/cache/redis"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/client"
	infraPostgres "github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/database/postgres"
	infraHttp "github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/http"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/http/handler"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/infra/pubsub"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/usecase/cart"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/usecase/order"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("[ORDER-SERVICE] Iniciando na porta %s...", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Conexão PostgreSQL
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Não foi possível conectar ao Postgres agora: %v (suba o docker-compose)", err)
	} else {
		defer dbPool.Close()
		log.Println("[POSTGRES] Conectado ao schema orders com sucesso.")
	}

	// 2. Conexão Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()
	log.Println("[REDIS] Conectado com sucesso.")

	// 3. Repositórios e Clientes
	orderRepo := infraPostgres.NewPostgresOrderRepository(dbPool)
	cartRepo := infraRedis.NewRedisCartRepository(rdb)
	publisher := pubsub.NewRedisPublisher(rdb)
	catalogClient := client.NewCatalogHTTPClient(cfg.CatalogBaseURL)

	// 4. Inicia ouvinte de carrinho abandonado em segundo plano
	keyspaceListener := infraRedis.NewKeyspaceListener(rdb, cartRepo)
	go keyspaceListener.StartListening(ctx)

	// 5. Casos de Uso
	cartUC := cart.NewCartUseCase(cartRepo, cfg.CartTTL)
	checkoutUC := order.NewCheckoutUseCase(orderRepo, cartRepo, catalogClient, publisher)

	// 6. Handlers e Router
	orderHandler := handler.NewOrderHandler(cartUC, checkoutUC, orderRepo)
	router := infraHttp.NewRouter(orderHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Início do Servidor com Graceful Shutdown
	go func() {
		log.Printf("[ORDER-SERVICE] Servidor pronto e escutando na porta :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no servidor HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[ORDER-SERVICE] Encerrando conexões com segurança...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Falha no shutdown: %v", err)
	}

	log.Println("[ORDER-SERVICE] Servidor finalizado.")
}

