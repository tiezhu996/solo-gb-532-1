package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sonar-survey-coverage-planner/backend/internal/dto"
	"sonar-survey-coverage-planner/backend/internal/service"
	"sonar-survey-coverage-planner/backend/pkg/api"
)

type AuthHandler struct {
	auth  *service.AuthService
	audit *service.AuditService
}

func NewAuthHandler(auth *service.AuthService, audit *service.AuditService) *AuthHandler {
	return &AuthHandler{auth: auth, audit: audit}
}

func (h *AuthHandler) Health(c *gin.Context) {
	api.Success(c, http.StatusOK, gin.H{"status": "ok", "service": "sonar-survey-coverage-planner"})
}

func (h *AuthHandler) Ready(c *gin.Context) {
	if err := h.audit.Ready(); err != nil {
		api.WriteError(c, api.Internal(err))
		return
	}
	api.Success(c, http.StatusOK, gin.H{"status": "ready", "database": "reachable"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		api.BindError(c, err)
		return
	}
	response, err := h.auth.Login(request)
	if err != nil {
		api.WriteError(c, err)
		return
	}
	api.Success(c, http.StatusOK, response)
}

func (h *AuthHandler) Me(c *gin.Context) {
	api.Success(c, http.StatusOK, dto.UserIdentity{ID: actorFrom(c).UserID, Username: c.GetString("username"), DisplayName: c.GetString("display_name"), Role: c.GetString("role")})
}

func actorFrom(c *gin.Context) service.Actor {
	id, _ := c.Get("user_id")
	userID, _ := id.(uint)
	return service.Actor{RequestID: c.GetString("request_id"), UserID: userID, Username: c.GetString("username"), Role: c.GetString("role")}
}

func idParam(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		api.WriteError(c, api.BadRequest("ID_INVALID", "资源 ID 无效", err))
		return 0, false
	}
	return uint(value), true
}

func pageQuery(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func uintQuery(c *gin.Context, name string) uint {
	value, _ := strconv.ParseUint(c.Query(name), 10, 64)
	return uint(value)
}

func writeServiceError(c *gin.Context, err error, entity string) {
	if err == nil {
		return
	}
	var appErr *api.AppError
	if errors.As(err, &appErr) {
		api.WriteError(c, appErr)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		api.WriteError(c, api.NotFound(entity, err))
		return
	}
	api.WriteError(c, api.Internal(err))
}

func cleanQuery(c *gin.Context, key string) string { return strings.TrimSpace(c.Query(key)) }
