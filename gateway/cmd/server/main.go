package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/gateway/internal"
	"github.com/Ronan-Rodrigues/commerce-ms/gateway/internal/middleware"
	"github.com/Ronan-Rodrigues/commerce-ms/gateway/internal/proxy"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := getEnv("GATEWAY_PORT", "8080")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	pubKeyPath := getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem")

	log.Printf("[API-GATEWAY] Iniciando na porta %s...", port)

	// 1. Conexão Redis (para Rate Limiting e Blacklist)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()
	log.Println("[REDIS] Cliente de Rate Limit conectado.")

	// 2. Chave Pública RSA
	pubKey, err := loadOrGenerateRSAPublicKey(pubKeyPath)
	if err != nil {
		log.Fatalf("[ERRO] Falha ao carregar chave pública: %v", err)
	}

	// 3. Middlewares
	authMW := middleware.NewAuthMiddleware(pubKey, rdb)
	rateLimiter := middleware.NewRateLimiter(rdb, 100) // 100 req/min por IP

	// 4. Configuração dos Proxies Reversos para os microsserviços
	authURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	catalogURL := getEnv("CATALOG_SERVICE_URL", "http://localhost:8082")
	orderURL := getEnv("ORDER_SERVICE_URL", "http://localhost:8083")
	paymentURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084")
	notifyURL := getEnv("NOTIFY_SERVICE_URL", "http://localhost:8085")

	authProxy, _ := proxy.NewReverseProxy(authURL)
	catalogProxy, _ := proxy.NewReverseProxy(catalogURL)
	orderProxy, _ := proxy.NewReverseProxy(orderURL)
	paymentProxy, _ := proxy.NewReverseProxy(paymentURL)
	notifyProxy, _ := proxy.NewReverseProxy(notifyURL)

	proxies := internal.GatewayProxies{
		Auth:    authProxy,
		Catalog: catalogProxy,
		Order:   orderProxy,
		Payment: paymentProxy,
		Notify:  notifyProxy,
	}

	// 5. Router
	router := internal.NewRouter(proxies, authMW, rateLimiter)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  0, // Permite SSE duradouro
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	// 6. Graceful Shutdown
	go func() {
		log.Printf("[API-GATEWAY] Gateway ativo e roteando na porta :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no gateway HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[API-GATEWAY] Encerrando conexões com segurança...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Falha no shutdown do gateway: %v", err)
	}

	log.Println("[API-GATEWAY] Gateway finalizado.")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func loadOrGenerateRSAPublicKey(pubPath string) (*rsa.PublicKey, error) {
	pubBytes, err := os.ReadFile(pubPath)
	if err == nil {
		block, _ := pem.Decode(pubBytes)
		if block != nil {
			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err == nil {
				return pubKey.(*rsa.PublicKey), nil
			}
		}
	}

	log.Println("[RSA] Chave pública local não encontrada. Gerando par efêmero em memória para ambiente de desenvolvimento...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &priv.PublicKey, nil
}

