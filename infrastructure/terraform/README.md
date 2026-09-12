# ☁️ Infraestrutura como Código (IaC) — AWS & Terraform

Este diretório contém o blueprint de Infraestrutura como Código (IaC) do **Commerce-MS** utilizando **Terraform**, projetado para rodar na nuvem da **Amazon Web Services (AWS)** seguindo o framework *AWS Well-Architected*.

---

## 🏗️ Recursos Provisionados

| Componente | Recurso AWS | Propósito na Arquitetura |
|---|---|---|
| **Rede** | `aws_vpc` + Subnets Multi-AZ | Isolamento de tráfego (subnets públicas para o Load Balancer e privadas para bancos de dados) |
| **Contêineres** | `aws_ecs_cluster` (Fargate) | Execução serverless dos microsserviços em Go sem necessidade de gerenciar instâncias EC2 |
| **Banco Relacional** | `aws_db_instance` (RDS PostgreSQL 16) | Persistência com schemas isolados (`auth`, `catalog`, `orders`, `payments`) |
| **Cache & Eventos** | `aws_elasticache_cluster` (Redis 7) | Rate Limiting (Sliding Window), Pub/Sub e Carrinho com TTL |
| **Load Balancer** | `aws_lb` (Application Load Balancer) | Ponto de entrada com health checks e roteamento |
| **CDN de Borda** | `aws_cloudfront_distribution` | Cache global de alta velocidade na borda, aliviando o tráfego do backend e mitigando ataques DDoS |

---

## 🚀 Como Executar Localmente (Validação)

Caso deseje inspecionar ou validar os arquivos do Terraform:

```bash
# 1. Inicializar providers
terraform init

# 2. Validar sintaxe dos arquivos
terraform validate

# 3. Planejar o provisionamento (dry-run sem custos)
terraform plan
```

> **Nota:** Para o ambiente de demonstração e portfólio gratuito, o projeto utiliza contêineres Docker locais via `docker-compose.yml` ou deploy gratuito no **Railway**. Este módulo de Terraform demonstra a capacidade de desenhar e provisionar arquiteturas de nuvem de grande porte.

