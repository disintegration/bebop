package api

import (
	"net/http"
	"strconv"

	"github.com/disintegration/bebop/store"
)

func (h *Handler) handleGetComments(w http.ResponseWriter, r *http.Request) {
	topic, err := strconv.ParseInt(r.URL.Query().Get("topic"), 10, 64)
	if err != nil || topic < 1 {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid topic ID")
		return
	}

	_, err = h.Store.Topics().Get(topic)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Topic not found")
			return
		}
		h.logError("get topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
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

	comments, count, err := h.Store.Comments().GetByTopic(topic, offset, limit)
	if err != nil {
		h.logError("get comments by topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Comments []*store.Comment `json:"comments"`
		Count    int              `json:"count"`
	}{
		Comments: comments,
		Count:    count,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleNewComment(w http.ResponseWriter, r *http.Request) {
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
		Topic   *int64  `json:"topic"`
		Content *string `json:"content"`
	}{}

	err := h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Topic == nil || *req.Topic < 1 {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid topic ID")
		return
	}

	if req.Content == nil || !store.ValidCommentContent(*req.Content) {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid comment content")
		return
	}

	_, err = h.Store.Topics().Get(*req.Topic)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Topic not found")
			return
		}
		h.logError("get topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	id, err := h.Store.Comments().New(*req.Topic, currentUser.ID, *req.Content)
	if err != nil {
		h.logError("create comment: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	_, count, err := h.Store.Comments().GetByTopic(*req.Topic, 0, 0)
	if err != nil {
		h.logError("get comments by topic: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		ID    int64 `json:"id"`
		Count int   `json:"count"`
	}{
		ID:    id,
		Count: count,
	}

	h.render(w, http.StatusCreated, response)
}

func (h *Handler) handleGetComment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid comment ID")
		return
	}

	comment, err := h.Store.Comments().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Comment not found")
			return
		}
		h.logError("get comment: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	response := struct {
		Comment *store.Comment `json:"comment"`
	}{
		Comment: comment,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
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
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid comment ID")
		return
	}

	_, err = h.Store.Comments().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "Comment not found")
			return
		}
		h.logError("get comment: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	err = h.Store.Comments().Delete(id)
	if err != nil {
		h.logError("delete comment: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	h.render(w, http.StatusOK, struct{}{})
}
