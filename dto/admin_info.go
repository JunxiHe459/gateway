package dto

import (
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	ID           int       `json:"id"`
	UsernName    string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type ChangePasswordInput struct {
	Password string `json:"password" form:"password" validate:"required" example:"password"`
}

func (param *ChangePasswordInput) BindParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
