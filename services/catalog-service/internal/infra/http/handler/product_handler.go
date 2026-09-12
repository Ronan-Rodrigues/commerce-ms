package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/usecase/product"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	createUC    *product.CreateProductUseCase
	listUC      *product.ListProductsUseCase
	deductUC    *product.DeductStockUseCase
	productRepo repository.ProductRepository
}

func NewProductHandler(
	createUC *product.CreateProductUseCase,
	listUC *product.ListProductsUseCase,
	deductUC *product.DeductStockUseCase,
	productRepo repository.ProductRepository,
) *ProductHandler {
	return &ProductHandler{
		createUC:    createUC,
		listUC:      listUC,
		deductUC:    deductUC,
		productRepo: productRepo,
	}
}

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"` // em centavos
	Stock       int    `json:"stock"`
	SKU         string `json:"sku"`
	CategoryID  string `json:"category_id"`
	ImageURL    string `json:"image_url"`
}

type deductStockRequest struct {
	Quantity int `json:"quantity"`
}

func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func renderError(w http.ResponseWriter, status int, message string) {
	renderJSON(w, status, map[string]string{"error": message})
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	input := product.CreateProductInput{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
	}

	out, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProductAlreadyExists):
			renderError(w, http.StatusConflict, err.Error())
		case errors.Is(err, repository.ErrCategoryNotFound):
			renderError(w, http.StatusNotFound, "categoria informada não existe")
		case errors.Is(err, entity.ErrInvalidPrice),
			errors.Is(err, entity.ErrNegativeStock),
			errors.Is(err, entity.ErrEmptyProductName):
			renderError(w, http.StatusBadRequest, err.Error())
		default:
			renderError(w, http.StatusInternalServerError, "erro ao criar produto")
		}
		return
	}

	renderJSON(w, http.StatusCreated, out)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	minPrice, _ := strconv.ParseInt(q.Get("min_price"), 10, 64)
	maxPrice, _ := strconv.ParseInt(q.Get("max_price"), 10, 64)

	input := product.ListProductsInput{
		CategoryID: q.Get("category_id"),
		Search:     q.Get("search"),
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		ActiveOnly: q.Get("active") != "false",
		Page:       page,
		Limit:      limit,
	}

	out, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "erro ao listar produtos")
		return
	}

	renderJSON(w, http.StatusOK, out)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.productRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			renderError(w, http.StatusNotFound, "produto não encontrado")
			return
		}
		renderError(w, http.StatusInternalServerError, "erro ao buscar produto")
		return
	}

	renderJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) DeductStock(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req deductStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	err := h.deductUC.Execute(r.Context(), product.DeductStockInput{
		ProductID: id,
		Quantity:  req.Quantity,
	})

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProductNotFound):
			renderError(w, http.StatusNotFound, "produto não encontrado")
		case errors.Is(err, product.ErrProductOutOfStock):
			renderError(w, http.StatusConflict, "estoque insuficiente")
		default:
			renderError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	renderJSON(w, http.StatusOK, map[string]string{"message": "estoque deduzido com sucesso"})
}

func (h *ProductHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "catalog-service",
	})
}
