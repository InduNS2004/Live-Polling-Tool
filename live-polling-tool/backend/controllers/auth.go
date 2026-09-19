package controllers

import (
	"github.com/gin-gonic/gin"
	"live-polling-tool/backend/services"
	"live-polling-tool/backend/utils"
	"net/http"
	"strings"
)

type AuthController struct {
	S      *services.AuthService
	Secure bool
}
type authReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cn *AuthController) Register(c *gin.Context) {
	var r authReq
	if c.ShouldBindJSON(&r) != nil {
		utils.Error(c, 400, "INVALID_REQUEST", "Valid JSON is required")
		return
	}
	u, e := cn.S.Register(c.Request.Context(), r.Email, r.Password)
	if e != nil {
		utils.Error(c, 400, "REGISTRATION_FAILED", "Use a valid email and password (minimum 8 characters)")
		return
	}
	utils.OK(c, gin.H{"id": u.ID, "email": u.Email})
}
func (cn *AuthController) Login(c *gin.Context) {
	var r authReq
	if c.ShouldBindJSON(&r) != nil {
		utils.Error(c, 400, "INVALID_REQUEST", "Valid JSON is required")
		return
	}
	u, t, e := cn.S.Login(c.Request.Context(), r.Email, r.Password)
	if e != nil {
		utils.Error(c, 401, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}
	if cn.Secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("auth_token", t, 86400, "/", "", cn.Secure, true)
	utils.OK(c, gin.H{"user": gin.H{"id": u.ID, "email": u.Email}})
}
func (cn *AuthController) Me(c *gin.Context) { utils.OK(c, gin.H{"id": c.GetString("userID")}) }
func NormalizeEmail(s string) string         { return strings.ToLower(strings.TrimSpace(s)) }

var _ = http.StatusOK
