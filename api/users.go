package api

import (
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/bebop/avatar"
	"github.com/disintegration/bebop/store"
)

const avatarUploadMaxBytes = 5 * 1024 * 1024

// extUser is a copy of store.User with more fields marshalled to JSON.
type extUser struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	AuthService string    `json:"authService"`
	AuthID      string    `json:"-"`
	Blocked     bool      `json:"blocked"`
	Admin       bool      `json:"admin"`
	Avatar      string    `json:"avatar"`
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)

	if currentUser != nil && currentUser.Avatar != "" {
		currentUser.Avatar = h.AvatarService.URL(currentUser)
	}

	response := struct {
		Authenticated bool     `json:"authenticated"`
		User          *extUser `json:"user,omitempty"`
	}{
		Authenticated: currentUser != nil,
		User:          (*extUser)(currentUser),
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)

	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Parameter 'ids' required")
		return
	}

	parts := strings.Split(idsParam, ",")
	if len(parts) > 1000 {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Too many ids")
		return
	}

	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			h.renderError(w, http.StatusBadRequest, "BadRequest", "Parameter 'ids' contains invalid id")
			return
		}
		ids = append(ids, id)
	}

	usermap, err := h.Store.Users().GetMany(ids)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "User(s) not found")
			return
		}
		h.logError("get many users: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	for _, user := range usermap {
		if user.Avatar != "" {
			user.Avatar = h.AvatarService.URL(user)
		}
	}

	if currentUser != nil && currentUser.Admin {
		users := make([]*extUser, 0, len(usermap))
		for _, user := range usermap {
			users = append(users, (*extUser)(user))
		}
		sort.Slice(users, func(i, j int) bool {
			return users[i].ID < users[j].ID
		})

		response := struct {
			Users []*extUser `json:"users"`
		}{
			Users: users,
		}

		h.render(w, http.StatusOK, response)
		return
	}

	users := make([]*store.User, 0, len(usermap))
	for _, user := range usermap {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	response := struct {
		Users []*store.User `json:"users"`
	}{
		Users: users,
	}

	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)

	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid user ID")
		return
	}

	user, err := h.Store.Users().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "User not found")
			return
		}
		h.logError("get user by id: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	if user.Avatar != "" {
		user.Avatar = h.AvatarService.URL(user)
	}

	if currentUser != nil && currentUser.Admin {
		response := struct {
			User *extUser `json:"user"`
		}{
			User: (*extUser)(user),
		}
		h.render(w, http.StatusOK, response)
		return
	}

	response := struct {
		User *store.User `json:"user"`
	}{
		User: user,
	}
	h.render(w, http.StatusOK, response)
}

func (h *Handler) handleSetUserName(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)
	if currentUser == nil {
		w.Header().Set("WWW-Authenticate", "Bearer")
		h.renderError(w, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid user ID")
		return
	}

	if !currentUser.Admin && currentUser.ID != id {
		h.renderError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	req := struct {
		Name *string `json:"name"`
	}{}

	err = h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Name == nil || !store.ValidUserName(*req.Name) {
		h.renderError(w, http.StatusBadRequest, "InvalidUserName", "Invalid user name")
		return
	}

	user, err := h.Store.Users().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "User not found")
			return
		}
		h.logError("get user by id: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	if user.Name == *req.Name {
		h.render(w, http.StatusOK, struct{}{})
		return
	}

	err = h.Store.Users().SetName(id, *req.Name)
	if err != nil {
		if err == store.ErrConflict {
			h.renderError(w, http.StatusConflict, "UnavailableUserName", "Username is already taken")
			return
		}
		h.logError("set user name: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	user.Name = *req.Name

	if user.Avatar == "" {
		err := h.AvatarService.Generate(user)
		if err != nil {
			h.logError("gen user avatar: %s", err)
		}
	}

	h.render(w, http.StatusOK, struct{}{})
}

func (h *Handler) handleSetUserAvatar(w http.ResponseWriter, r *http.Request) {
	currentUser := h.currentUser(r)
	if currentUser == nil {
		w.Header().Set("WWW-Authenticate", "Bearer")
		h.renderError(w, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	id, err := strconv.ParseInt(h.urlParam(r, "id"), 10, 64)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid user ID")
		return
	}

	if !currentUser.Admin && currentUser.ID != id {
		h.renderError(w, http.StatusForbidden, "Forbidden", "Access denied")
		return
	}

	user, err := h.Store.Users().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "User not found")
			return
		}
		h.logError("get user by id: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	req := struct {
		Avatar *string `json:"avatar"`
	}{}

	err = h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Avatar == nil || *req.Avatar == "" {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid avatar")
		return
	}

	avatarDataLen := base64.StdEncoding.DecodedLen(len(*req.Avatar))
	if avatarDataLen > avatarUploadMaxBytes {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Avatar data too large")
		return
	}

	avatarData, err := base64.StdEncoding.DecodeString(*req.Avatar)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid avatar data")
		return
	}

	err = h.AvatarService.Save(user, avatarData)
	switch {
	case err == avatar.ErrImageDecode:
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Avatar image decode failed")
		return

	case err == avatar.ErrImageTooLarge:
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Avatar image too large")
		return

	case err == avatar.ErrImageTooSmall:
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Avatar image too small")
		return

	case err != nil:
		h.logError("failed to save avatar: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	h.render(w, http.StatusOK, struct{}{})
}

func (h *Handler) handleSetUserBlocked(w http.ResponseWriter, r *http.Request) {
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
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid user ID")
		return
	}

	req := struct {
		Blocked *bool `json:"blocked"`
	}{}

	err = h.parseRequest(r, &req)
	if err != nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid request body")
		return
	}

	if req.Blocked == nil {
		h.renderError(w, http.StatusBadRequest, "BadRequest", "Invalid blocked")
		return
	}

	user, err := h.Store.Users().Get(id)
	if err != nil {
		if err == store.ErrNotFound {
			h.renderError(w, http.StatusNotFound, "NotFound", "User not found")
			return
		}
		h.logError("get user by id: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	if user.Blocked == *req.Blocked {
		h.render(w, http.StatusOK, struct{}{})
		return
	}

	err = h.Store.Users().SetBlocked(id, *req.Blocked)
	if err != nil {
		h.logError("set user blocked: %s", err)
		h.renderError(w, http.StatusInternalServerError, "ServerError", "Server error")
		return
	}

	h.render(w, http.StatusOK, struct{}{})
}
