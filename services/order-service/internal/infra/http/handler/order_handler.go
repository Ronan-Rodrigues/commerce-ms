package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/usecase/cart"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/usecase/order"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	cartUC     *cart.CartUseCase
	checkoutUC *order.CheckoutUseCase
	orderRepo  repository.OrderRepository
}

func NewOrderHandler(
	cartUC *cart.CartUseCase,
	checkoutUC *order.CheckoutUseCase,
	orderRepo repository.OrderRepository,
) *OrderHandler {
	return &OrderHandler{
		cartUC:     cartUC,
		checkoutUC: checkoutUC,
		orderRepo:  orderRepo,
	}
}

type addItemRequest struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Quantity  int    `json:"quantity"`
}

func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func renderError(w http.ResponseWriter, status int, message string) {
	renderJSON(w, status, map[string]string{"error": message})
}

func getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "demo-user-123" // fallback amigável em testes locais
	}
	return userID
}

// GetCart GET /cart
func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	c, err := h.cartUC.GetCart(r.Context(), userID)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "erro ao buscar carrinho")
		return
	}
	renderJSON(w, http.StatusOK, c)
}

// AddCartItem POST /cart/items
func (h *OrderHandler) AddCartItem(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	c, err := h.cartUC.AddItem(r.Context(), cart.AddItemInput{
		UserID:    userID,
		ProductID: req.ProductID,
		Name:      req.Name,
		Price:     req.Price,
		Quantity:  req.Quantity,
	})
	if err != nil {
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	renderJSON(w, http.StatusOK, c)
}

// RemoveCartItem DELETE /cart/items/{productId}
func (h *OrderHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	productID := chi.URLParam(r, "productId")

	c, err := h.cartUC.RemoveItem(r.Context(), userID, productID)
	if err != nil {
		if errors.Is(err, entity.ErrCartItemNotFound) {
			renderError(w, http.StatusNotFound, "item não encontrado no carrinho")
			return
		}
		renderError(w, http.StatusBadRequest, err.Error())
		return
	}

	renderJSON(w, http.StatusOK, c)
}

// Checkout POST /orders/checkout
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	idempotencyKey := r.Header.Get("Idempotency-Key")

	out, err := h.checkoutUC.Execute(r.Context(), order.CheckoutInput{
		OrderID:        uuid.NewString(),
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil {
		if errors.Is(err, order.ErrCartIsEmpty) {
			renderError(w, http.StatusBadRequest, "carrinho está vazio")
			return
		}
		renderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	renderJSON(w, http.StatusCreated, out)
}

// ListOrders GET /orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	orders, err := h.orderRepo.ListByUserID(r.Context(), userID)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "erro ao listar pedidos")
		return
	}
	renderJSON(w, http.StatusOK, orders)
}

// GetOrderByID GET /orders/{id}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.orderRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			renderError(w, http.StatusNotFound, "pedido não encontrado")
			return
		}
		renderError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		return
	}
	renderJSON(w, http.StatusOK, o)
}

func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "order-service",
	})
}
