package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/config"
	infraPostgres "github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/infra/database/postgres"
	infraHttp "github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/infra/http"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/infra/http/handler"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/usecase/product"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("[CATALOG-SERVICE] Iniciando na porta %s...", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Conexão PostgreSQL
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Não foi possível conectar ao Postgres agora: %v (suba o docker-compose)", err)
	} else {
		defer dbPool.Close()
		log.Println("[POSTGRES] Conectado ao schema catalog com sucesso.")
	}

	// 2. Repositórios
	productRepo := infraPostgres.NewPostgresProductRepository(dbPool)
	categoryRepo := infraPostgres.NewPostgresCategoryRepository(dbPool)

	// 3. Casos de Uso
	createUC := product.NewCreateProductUseCase(productRepo, categoryRepo)
	listUC := product.NewListProductsUseCase(productRepo)
	deductUC := product.NewDeductStockUseCase(productRepo)

	// 4. Handlers e Router
	productHandler := handler.NewProductHandler(createUC, listUC, deductUC, productRepo)
	router := infraHttp.NewRouter(productHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Início do Servidor com Graceful Shutdown
	go func() {
		log.Printf("[CATALOG-SERVICE] Servidor pronto e escutando na porta :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no servidor HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[CATALOG-SERVICE] Encerrando conexões com segurança...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Falha no shutdown: %v", err)
	}

	log.Println("[CATALOG-SERVICE] Servidor finalizado.")
}
