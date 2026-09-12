package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/config"
	infraRedis "github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/cache/redis"
	infraPostgres "github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/database/postgres"
	infraHttp "github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/http"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/http/handler"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/security"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/login"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/refresh"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/usecase/register"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("[AUTH-SERVICE] Iniciando na porta %s...", cfg.Port)

	// 1. Contexto raiz
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Conexão com PostgreSQL
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[WARN] Não foi possível conectar ao Postgres agora: %v (suba o docker-compose)", err)
	} else {
		defer dbPool.Close()
		log.Println("[POSTGRES] Conectado com sucesso.")
	}

	// 3. Conexão com Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	defer rdb.Close()
	log.Println("[REDIS] Cliente configurado.")

	// 4. Par de Chaves RSA (Gera em memória para dev se os arquivos pem não existirem)
	privKey, pubKey, err := loadOrGenerateRSAKeys(cfg.JWTKeyPath, cfg.JWTPublicPath)
	if err != nil {
		log.Fatalf("[ERRO] Falha ao carregar/gerar chaves RSA: %v", err)
	}

	// 5. Infraestrutura & Segurança
	userRepo := infraPostgres.NewPostgresUserRepository(dbPool)
	tokenStore := infraRedis.NewRedisTokenStore(rdb)

	hasher := security.NewArgon2Hasher(security.Argon2Params{
		Memory:      cfg.Argon2Memory,
		Iterations:  cfg.Argon2Iter,
		Parallelism: cfg.Argon2Parallel,
		SaltLength:  16,
		KeyLength:   32,
	})

	jwtService := security.NewJWTService(privKey, pubKey, "commerce-ms")

	// 6. Casos de Uso
	registerUC := register.NewUseCase(userRepo, hasher)
	loginUC := login.NewUseCase(userRepo, hasher, jwtService, login.TokenConfig{
		AccessTTL: cfg.JWTAccessTTL,
	})
	refreshUC := refresh.NewUseCase(userRepo, tokenStore, jwtService, refresh.TokenConfig{
		AccessTTL:  cfg.JWTAccessTTL,
		RefreshTTL: cfg.JWTRefreshTTL,
	})

	// 7. Handlers HTTP e Router
	authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshUC)
	router := infraHttp.NewRouter(authHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful Shutdown
	go func() {
		log.Printf("[AUTH-SERVICE] Servidor pronto e escutando na porta :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERRO] Falha no servidor HTTP: %v", err)
		}
	}()

	// Aguarda sinal de encerramento do SO
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[AUTH-SERVICE] Sinal de desligamento recebido. Encerrando conexões com segurança...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERRO] Erro no encerramento forçado do servidor: %v", err)
	}

	log.Println("[AUTH-SERVICE] Servidor finalizado.")
}

// loadOrGenerateRSAKeys carrega do disco ou gera um par novo caso não exista (ideal para dev)
func loadOrGenerateRSAKeys(privPath, pubPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Tenta ler do disco
	privBytes, err := os.ReadFile(privPath)
	if err == nil {
		block, _ := pem.Decode(privBytes)
		if block != nil {
			privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err == nil {
				return privKey, &privKey.PublicKey, nil
			}
		}
	}

	// Caso não encontre arquivos locais de chave, gera par efêmero seguro de 2048 bits
	log.Println("[RSA] Chaves locais não encontradas. Gerando par de chaves RSA de 2048 bits em memória para ambiente de desenvolvimento...")
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao gerar chave RSA: %w", err)
	}

	return privKey, &privKey.PublicKey, nil
}
