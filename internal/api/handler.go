package api

import (
	"net/http"
	"strconv"

	"encoding/json"
	"github.com/go-chi/chi/v5"
	"vectraDB/internal/models"
	"vectraDB/internal/store"
	"vectraDB/internal/utils"
	"vectraDB/pkg/errors"
	"vectraDB/pkg/response"
	"github.com/sirupsen/logrus"
	"vectraDB/internal/logger"
)

type Handler struct {
	store store.Store
}

func NewHandler(store store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Routes() *chi.Mux {
	r := chi.NewRouter()

	// Vector routes
	r.Route("/vectors", func(r chi.Router) {
		r.Post("/", h.CreateVector)
		r.Get("/{id}", h.GetVector)
		r.Put("/{id}", h.UpdateVector)
		r.Delete("/{id}", h.DeleteVector)
		r.Get("/", h.ListVectors)
	})

	// Search routes
	r.Route("/search", func(r chi.Router) {
		r.Post("/", h.SearchVectors)
		r.Post("/hybrid", h.HybridSearch)
	})

	// Document routes
	r.Route("/documents", func(r chi.Router) {
		r.Post("/", h.CreateDocument)
		r.Get("/{id}", h.GetDocument)
		r.Put("/{id}", h.UpdateDocument)
		r.Delete("/{id}", h.DeleteDocument)
		r.Get("/", h.ListDocuments)
		r.Get("/tags/{tag}", h.ListDocumentsByTag)
	})

	// Health check
	r.Get("/health", h.Health)

	return r
}

func (h *Handler) CreateVector(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVectorRequest
	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	vector := &models.Vector{
		ID:       req.ID,
		Vector:   req.Vector,
		Text:     req.Text,
		Metadata: req.Metadata,
	}

	if err := h.store.InsertVector(r.Context(), vector); err != nil {
		response.Error(w, err)
		return
	}

	response.Created(w, vector)
}

func (h *Handler) GetVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("vector ID is required"))
		return
	}

	vector, err := h.store.GetVector(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, vector)
}

func (h *Handler) UpdateVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("vector ID is required"))
		return
	}

	var req models.UpdateVectorRequest
	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	vector := &models.Vector{
		ID:       id,
		Vector:   req.Vector,
		Text:     req.Text,
		Metadata: req.Metadata,
	}

	if err := h.store.UpdateVector(r.Context(), id, vector); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, vector)
}

func (h *Handler) DeleteVector(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("vector ID is required"))
		return
	}

	if err := h.store.DeleteVector(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

func (h *Handler) ListVectors(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	vectors, err := h.store.ListVectors(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMeta(w, vectors, &response.Meta{
		Limit: limit,
		Page:  (offset/limit) + 1,
	})
}

func (h *Handler) SearchVectors(w http.ResponseWriter, r *http.Request) {
	var req models.SearchRequest
	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	result, err := h.store.SearchVectors(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMeta(w, result.Results, &response.Meta{
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

func (h *Handler) HybridSearch(w http.ResponseWriter, r *http.Request) {
	var req models.HybridSearchRequest
	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	result, err := h.store.HybridSearch(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMeta(w, result.Results, &response.Meta{
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDocumentRequest

	logger.Info("CreateDocument: received request")

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"endpoint": "/create-document",
			"action":   "decode request",
		}).Error("Failed to decode request body")
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "invalid JSON"))
		return
	}

	logger.WithFields(logrus.Fields{
		"endpoint": "/create-document",
		"action":   "decode request",
	}).Info("Request body decoded")

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"endpoint": "/create-document",
			"action":   "validate request",
		}).Error("Validation failed")
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	logger.WithFields(logrus.Fields{
		"endpoint": "/create-document",
		"action":   "validate request",
	}).Info("Request validation passed")

	document := &models.Document{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
		Tags:    req.Tags,
	}

	logger.WithFields(logrus.Fields{
		"document_id": document.ID,
		"title":       document.Title,
		"tags":        document.Tags,
	}).Debug("Constructed document struct")

	if err := h.store.InsertDocument(r.Context(), document); err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"document_id": document.ID,
			"action":      "insert document",
		}).Error("Failed to insert document")
		response.Error(w, err)
		return
	}

	logger.WithFields(logrus.Fields{
		"document_id": document.ID,
		"action":      "insert document",
	}).Info("Successfully created document")

	response.Created(w, document)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("document ID is required"))
		return
	}

	document, err := h.store.GetDocument(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, document)
}

func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("document ID is required"))
		return
	}

	var req models.UpdateDocumentRequest
	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, errors.Wrap(err, http.StatusBadRequest, "validation failed"))
		return
	}

	document := &models.Document{
		ID:      id,
		Title:   req.Title,
		Content: req.Content,
		Tags:    req.Tags,
	}

	if err := h.store.UpdateDocument(r.Context(), id, document); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, document)
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("document ID is required"))
		return
	}

	if err := h.store.DeleteDocument(r.Context(), id); err != nil {
		response.Error(w, err)
		return
	}

	response.NoContent(w)
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	documents, err := h.store.ListDocuments(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMeta(w, documents, &response.Meta{
		Limit: limit,
		Page:  (offset/limit) + 1,
	})
}

func (h *Handler) ListDocumentsByTag(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")
	if tag == "" {
		response.Error(w, errors.ErrInvalidInput.WithDetails("tag is required"))
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	documents, err := h.store.ListDocumentsByTag(r.Context(), tag, limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.SuccessWithMeta(w, documents, &response.Meta{
		Limit: limit,
		Page:  (offset/limit) + 1,
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Health(r.Context()); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status": "healthy",
	})
}
