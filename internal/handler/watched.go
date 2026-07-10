package handler

import (
	"net/http"

	"github.com/florentsorel/watchue/internal/catalog"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/labstack/echo/v5"
)

type watchedResourceResponse struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Notify bool   `json:"notify"`
}

// GetWatched lists every resource the user has chosen to watch.
func (h *Handler) GetWatched(c *echo.Context) error {
	rows, err := h.db.ListWatchedResources(c.Request().Context())
	if err != nil {
		return jsonInternalError(c, err)
	}

	resp := make([]watchedResourceResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, watchedResourceResponse{ID: r.ResourceID, Type: r.ResourceType, Name: r.Name, Notify: r.Notify != 0})
	}
	return c.JSON(http.StatusOK, resp)
}

// PutWatched marks a resource as watched. type/name are resolved from the
// bridge, not trusted from the client, so an unknown id gets a 422.
func (h *Handler) PutWatched(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return jsonError(c, http.StatusBadRequest, "resource id is required")
	}

	ctx := c.Request().Context()
	cat, err := catalog.Build(ctx, h.hue)
	if err != nil {
		return jsonBridgeError(c, err)
	}
	resourceType, name, ok := cat.Resolve(id)
	if !ok {
		return jsonError(c, http.StatusUnprocessableEntity, "no such resource on the bridge")
	}

	err = h.db.UpsertWatchedResource(ctx, db.UpsertWatchedResourceParams{
		ResourceID:   id,
		ResourceType: resourceType,
		Name:         name,
	})
	if err != nil {
		return jsonInternalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

type patchWatchedRequest struct {
	Notify *bool `json:"notify"`
}

// PatchWatched toggles notify on an already-watched resource. 404 if unwatched.
func (h *Handler) PatchWatched(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return jsonError(c, http.StatusBadRequest, "resource id is required")
	}

	var req patchWatchedRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request body")
	}
	if req.Notify == nil {
		return jsonError(c, http.StatusBadRequest, "notify is required")
	}

	notify := int64(0)
	if *req.Notify {
		notify = 1
	}
	rows, err := h.db.SetWatchedResourceNotify(c.Request().Context(), db.SetWatchedResourceNotifyParams{
		ResourceID: id,
		Notify:     notify,
	})
	if err != nil {
		return jsonInternalError(c, err)
	}
	if rows == 0 {
		return jsonError(c, http.StatusNotFound, "resource is not currently watched")
	}
	return c.NoContent(http.StatusNoContent)
}

// DeleteWatched stops watching a resource.
func (h *Handler) DeleteWatched(c *echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return jsonError(c, http.StatusBadRequest, "resource id is required")
	}
	if err := h.db.DeleteWatchedResource(c.Request().Context(), id); err != nil {
		return jsonInternalError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
