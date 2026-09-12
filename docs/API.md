# 📖 Especificação da API REST — Commerce-MS

Todas as requisições externas devem ser direcionadas ao **API Gateway (porta 8080 ou porta 80 do Nginx)**.

---

## 🔐 1. Autenticação (`/auth`)

### `POST /auth/register`
Cadastra um novo usuário cliente.
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ronan Rodrigues",
    "email": "ronan@exemplo.com",
    "password": "senhaSegura123"
  }'
```
**Resposta (`201 Created`):**
```json
{
  "id": "c71a39f6-6c9f-4318-97f3-e5d89f816431",
  "name": "Ronan Rodrigues",
  "email": "ronan@exemplo.com",
  "role": "customer"
}
```

### `POST /auth/login`
Autentica e retorna tokens de acesso (RS256) e refresh token.
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "ronan@exemplo.com",
    "password": "senhaSegura123"
  }'
```
**Resposta (`200 OK`):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "8f3b2a1c...",
  "expires_in": 900,
  "user": {
    "id": "c71a39f6-6c9f-4318-97f3-e5d89f816431",
    "name": "Ronan Rodrigues",
    "email": "ronan@exemplo.com",
    "role": "customer"
  }
}
```

---

## 📦 2. Catálogo de Produtos (`/products`)

### `GET /products`
Lista produtos com paginação e filtros opcionais (`category_id`, `search`, `min_price`, `max_price`, `page`, `limit`).
```bash
curl -X GET "http://localhost:8080/products?page=1&limit=10&search=teclado"
```

### `POST /products` *(Requer Bearer Token)*
Criação de produto (área de administração).
```bash
curl -X POST http://localhost:8080/products \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Teclado Mecânico RGB",
    "description": "Switch Blue de alta performance",
    "price": 29990,
    "stock": 15,
    "sku": "TEC-MEC-01"
  }'
```

---

## 🛒 3. Carrinho de Compras (`/cart`) *(Requer Bearer Token)*

### `GET /cart`
Retorna os itens atuais do carrinho armazenados no Redis (TTL de 30 min).
```bash
curl -X GET http://localhost:8080/cart \
  -H "Authorization: Bearer <TOKEN>"
```

### `POST /cart/items`
Adiciona ou incrementa um item no carrinho.
```bash
curl -X POST http://localhost:8080/cart/items \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "prod-1",
    "name": "Teclado Mecânico RGB",
    "price": 29990,
    "quantity": 1
  }'
```

---

## 💳 4. Checkout e Pedidos (`/orders`) *(Requer Bearer Token)*

### `POST /orders/checkout`
Finaliza o carrinho ativo com suporte a `Idempotency-Key` (evita compras duplicadas em cliques repetidos).
```bash
curl -X POST http://localhost:8080/orders/checkout \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Idempotency-Key: uuid-v4-unico-por-clique"
```

---

## 📡 5. Eventos em Tempo Real (`/events`)

### `GET /events`
Stream contínuo de eventos via Server-Sent Events (SSE).
```bash
curl -N -H "Accept: text/event-stream" http://localhost:8080/events
```
**Stream recebido:**
```text
event: order.criado
data: {"order_id":"ord-001","status":"pendente","total_amount":29990}

event: payment.confirmado
data: {"order_id":"ord-001","status":"pago"}
```

