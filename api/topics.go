package api

import (
	"net/http"
	"strconv"

	"github.com/disintegration/bebop/store"
)

func (h *Handler) handleGetTopics(w http.ResponseWriter, r *http.Request) {
	var err error

	var category int64
	if v := r.URL.Query().Get("category"); v != "" {
		category, err = strconv.ParseInt(v, 10, 64)
		if err != nil || category < 1 {
			h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid category ID")
			return
		}
		_, err = h.Store.Categories().Get(category)
		if err != nil {
			if err == store.ErrNotFound {
				h.renderError(w, http.StatusNotFound, "NotFound", "Category not found")
				return
			}
			h.logError("get category: %s", err)
			h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
			return
		}
	}

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

	var topics []*store.Topic
	var count int

	topics, count, err = h.Store.Topics().GetByCategory(category, offset, limit)
	if err != nil {
		h.logError("get all topics: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Topics []*store.Topic `json:"topics"`
		Count  int            `json:"count"`
	}{
		Topics: topics,
		Count:  count,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleNewTopic(w http.ResponseWriter, r *http.Request) {
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

	req := struct {
		Category *int64  `json:"category"`
		Title    *string `json:"title"`
		Content  *string `json:"content"`
	}{}

	err := h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Category == nil || *req.Category <= 0 {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid category")
		return
	}

	if req.Title == nil || !store.ValidTopicTitle(*req.Title) {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid topic title")
		return
	}

	if req.Content == nil || !store.ValidCommentContent(*req.Content) {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid comment content")
		return
	}

	id, err := h.Store.Topics().New(*req.Category, currentUser.ID, *req.Title)
	if err != nil {
		h.logError("create topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	commentID, err := h.Store.Comments().New(id, currentUser.ID, *req.Content)
	if err != nil {
		h.Store.Topics().Delete(id)
		h.logError("create comment: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		ID        int64 `json:"id"`
		CommentID int64 `json:"commentId"`
	}{
		ID:        id,
		CommentID: commentID,
	}

	h.render(w, http.StatusCreated, response)
}

func (h *Handler) handleGetTopic(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid topic ID")
		return
	}

	topic, err := h.Store.Topics().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Topic not found")
			return
		}
		h.logError("get topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Topic *store.Topic `json:"topic"`
	}{
		Topic: topic,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleDeleteTopic(w http.ResponseWriter, r *http.Request) {
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
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid topic ID")
		return
	}

	_, err = h.Store.Topics().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Topic not found")
			return
		}
		h.logError("get topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	err = h.Store.Topics().Delete(id)
	if err != nil {
		h.logError("delete topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	h.render(w, http.StatusOK, struct{}{})
}
