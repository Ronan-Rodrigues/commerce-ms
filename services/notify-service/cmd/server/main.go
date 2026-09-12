package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/config"
	infraRedis "github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/infra/cache/redis"
	infraHttp "github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/infra/http"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/infra/http/handler"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/infra/n8n"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/notificar"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/sse"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("[NOTIFY-SERVICE] Iniciando na porta %s...", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Cliente Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()
	log.Println("[REDIS] Conectado com sucesso.")

	// 2. Broker SSE (em tempo real)
	broker := sse.NewBroker()
	go broker.Start()
	log.Println("[SSE-BROKER] Broker iniciado em background.")

	// 3. Cliente n8n
	n8nClient := n8n.NewClient(cfg.N8NBaseURL, cfg.N8NSecret)

	// 4. Orquestrador de Notificações
	orchestrator := notificar.NewNotificationOrchestrator(broker, n8nClient)

	// 5. Inicia o Subscriber Redis nos canais de pedidos, pagamentos e carrinho abandonado
	channels := []string{"order.criado", "payment.confirmado", "cart.abandonado"}
	subscriber := infraRedis.NewRedisSubscriber(rdb, orchestrator, channels)
	go subscriber.StartListening(ctx)

	// 6. Handlers e Router
	notifyHandler := handler.NewNotifyHandler(broker)
	router := infraHttp.NewRouter(notifyHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  0, // ReadTimeout 0 é essencial para streaming SSE duradouro
		WriteTimeout: 0, // WriteTimeout 0 é essencial para conexões SSE longas
		IdleTimeout:  120 * time.Second,
	}

	// 7. Servidor HTTP com Graceful Shutdown
	go func() {
		log.Printf("[NOTIFY-SERVICE] Servidor pronto e escutando na porta :%s (SSE ativo em /events)", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no servidor HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[NOTIFY-SERVICE] Encerrando conexões com segurança...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Falha no shutdown: %v", err)
	}

	log.Println("[NOTIFY-SERVICE] Servidor finalizado.")
}
