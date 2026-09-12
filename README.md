# 🛒 Commerce-MS — Plataforma de Microsserviços em Go

> Plataforma completa de comércio eletrônico desenvolvida em **Go (Golang)** sob os princípios da **Clean Architecture**, orientada a eventos com **Redis Pub/Sub**, atualizações em tempo real via **Server-Sent Events (SSE)**, automações com **n8n / WhatsApp (WAHA)** e infraestrutura em nuvem como código (**Terraform**).

---

## 🚀 Tecnologias & Arquitetura

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![TailwindCSS](https://img.shields.io/badge/Tailwind-3-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Multi--Stage-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-AWS-7B42BC?style=for-the-badge&logo=terraform&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/CI%2FCD-Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white)

---

## 🏛️ Visão da Arquitetura Distribuída

```
                     ┌───────────────────────────────┐
                     │   CloudFront CDN / Nginx :80  │
                     └──────────────┬────────────────┘
                                    │
                                    ▼
                     ┌───────────────────────────────┐
                     │     API Gateway (:8080)       │
                     │  • Rate Limiting (Redis)      │
                     │  • Validação JWT RS256        │
                     └──────────────┬────────────────┘
                                    │
    ┌──────────────┬────────────────┼────────────────┬──────────────┐
    ▼              ▼                ▼                ▼              ▼
┌─────────┐  ┌───────────┐    ┌───────────┐    ┌───────────┐  ┌───────────┐
│  Auth   │  │  Catalog  │    │   Order   │    │  Payment  │  │  Notify   │
│ Service │  │  Service  │    │  Service  │    │  Service  │  │  Service  │
│  :8081  │  │   :8082   │    │   :8083   │    │   :8084   │  │   :8085   │
└────┬────┘  └─────┬─────┘    └─────┬─────┘    └─────┬─────┘  └─────┬─────┘
     │             │                │                │              │
     ├─────────────┴────────────────┴────────────────┴──────────────┤
     ▼                                                              ▼
┌───────────────────────────┐                      ┌───────────────────────────┐
│       PostgreSQL 16       │                      │          Redis 7          │
│ • auth.users              │                      │ • Sliding Window Limit    │
│ • catalog.products        │                      │ • Carrinho com TTL        │
│ • orders.orders           │                      │ • Keyspace Notifications  │
│ • payments.transactions   │                      │ • Pub/Sub de Eventos      │
└───────────────────────────┘                      └─────────────┬─────────────┘
                                                                 │
                                                                 ▼
                                                   ┌───────────────────────────┐
                                                   │      n8n + WAHA WhatsApp  │
                                                   │  (Recuperação & Alertas)  │
                                                   └───────────────────────────┘
```

---

## 📦 Mapa de Microsserviços

| Serviço | Porta | Responsabilidade & Destaques de Engenharia |
|---|:---:|---|
| **API Gateway** | `8080` | Ponto único de entrada, rate limiting via Sliding Window no Redis, autenticação JWT RS256 por chave pública e proxy reverso. |
| **Auth Service** | `8081` | Gerenciamento de credenciais com **Argon2id (RFC 9106)**, assinatura assimétrica JWT RS256 e rotação de Refresh Token. |
| **Catalog Service** | `8082` | Catálogo de produtos e categorias com slug automático e **baixa atômica de estoque** contra condições de corrida (*race conditions*). |
| **Order Service** | `8083` | Carrinho em Redis com TTL de 30 min, **checkout idempotente** via `Idempotency-Key` e ouvinte de **carrinho abandonado**. |
| **Payment Service** | `8084` | Mock de gateway (Stripe/Mercado Pago), validação de assinatura **HMAC SHA256** e métricas financeiras. |
| **Notify Service** | `8085` | **Server-Sent Events (SSE)** em tempo real para navegadores, consumidor Redis Pub/Sub e disparo de webhooks para o **n8n / WAHA (WhatsApp)**. |
| **Frontend SPA** | `3000` | Interface em React + TypeScript + Vite + Tailwind CSS com validações Zod e dashboard em tempo real via SSE. |

---

## ⚡ Como Rodar o Projeto Completo Localmente

Você não precisa instalar Go nem ferramentas de banco na sua máquina. Apenas com o **Docker**:

```bash
# 1. Clone o repositório
git clone https://github.com/Ronan-Rodrigues/commerce-ms.git
cd commerce-ms

# 2. Copie as variáveis de ambiente
cp .env.example .env

# 3. Suba todos os contêineres simultaneamente
docker compose up --build
```

### URLs de Acesso:
- **Frontend Web:** [http://localhost:3000](http://localhost:3000)
- **Nginx (Load Balancer):** [http://localhost:80](http://localhost:80)
- **API Gateway:** [http://localhost:8080](http://localhost:8080)
- **SSE Stream:** [http://localhost:8080/events](http://localhost:8080/events)
- **PostgreSQL:** `localhost:5432` (user: `commerce`, pass: `commerce123`)
- **Redis:** `localhost:6379`

---

## 🧪 Testes Automatizados

Cada serviço possui testes unitários isolados com mocks em memória, garantindo alta velocidade de execução:

```bash
# Executar testes de qualquer serviço (ex: auth-service):
cd services/auth-service
go test -v ./...
```

---

## 📚 Documentação Adicional

- [Documentação Completa de Arquitetura](docs/ARCHITECTURE.md)
- [Especificação de Endpoints da API REST](docs/API.md)
- [Infraestrutura como Código (Terraform AWS)](infrastructure/terraform/README.md)

---

## 👨‍💻 Autor

Desenvolvido por **Ronan Rodrigues**
- **GitHub:** [@Ronan-Rodrigues](https://github.com/Ronan-Rodrigues)
- **Foco:** Engenharia de Software, Microsserviços em Go, Automações com n8n e Computação em Nuvem.
