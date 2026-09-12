variable "aws_region" {
  description = "Região AWS de provisionamento"
  type        = string
  default     = "sa-east-1" # São Paulo
}

variable "environment" {
  description = "Ambiente de deploy (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "vpc_cidr" {
  description = "Bloco CIDR da VPC principal"
  type        = string
  default     = "10.0.0.0/16"
}

variable "db_password" {
  description = "Senha mestra do RDS PostgreSQL"
  type        = string
  sensitive   = true
  default     = "CommerceRDS2026Secure!"
}

