package handler

import (
	"net/http"

	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type CategoryHandler struct {
	catRepo *repository.CategoryRepository
}

func NewCategoryHandler(catRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{catRepo: catRepo}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}
