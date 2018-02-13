package api

import (
	"net/http"
	"strconv"

	"github.com/disintegration/bebop/store"
)

func (h *Handler) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	var err error

	offset := 0
	offsetParam := r.URL.Query().Get("offset")
	if offsetParam != "" {
		offset, err = strconv.Atoi(offsetParam)
		if err != nil || offset < 0 {
			h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid offset")
			return
		}
	}

	limit := 10
	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil || limit < 1 || limit > 1000 {
			h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid limit")
			return
		}
	}

	var cats []*store.Category
	var count int

	cats, count, err = h.Store.Categories().GetLatest(offset, limit)
	if err != nil {
		h.logError("get all categories: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Categories []*store.Category `json:"categories"`
		Count      int               `json:"count"`
	}{
		Categories: cats,
		Count:      count,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleNewCategory(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)
	if currentUser == nil {
		w.Header().Set("WWW-Authenticate", "Bearer")
		h.renderError(w, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	if currentUser.Name == "" {
		h.renderError(w, http.StatusForbidden, "Forbidden", "User name is empty")
		return
	}

	if !currentUser.Admin {
		h.renderError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	req := struct {
		Title *string `json:"title"`
	}{}

	err := h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Title == nil || !store.ValidCategoryTitle(*req.Title) {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid category title")
		return
	}

	id, err := h.Store.Categories().New(currentUser.ID, *req.Title)
	if err != nil {
		h.logError("create category: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		ID int64 `json:"id"`
	}{
		ID: id,
	}

	h.render(w, http.StatusCreated, response)
}

func (h *Handler) handleGetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid category ID")
		return
	}

	cat, err := h.Store.Categories().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Category not found")
			return
		}
		h.logError("get category: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Category *store.Category `json:"category"`
	}{
		Category: cat,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)
	if currentUser == nil {
		w.Header().Set("WWW-Authenticate", "Bearer")
		h.renderError(w, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	if !currentUser.Admin {
		h.renderError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid category ID")
		return
	}

	_, err = h.Store.Categories().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Category not found")
			return
		}
		h.logError("get category: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	err = h.Store.Categories().Delete(id)
	if err != nil {
		h.logError("delete category: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	h.render(w, http.StatusOK, struct{}{})
}
