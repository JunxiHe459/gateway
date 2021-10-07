package dto

import (
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginInput struct {
	// 如果 validate 设置为 required，则会用 gin 提供的 validator 去进行校验
	// 可以在 validate 后面添加自定义验证器, 参见 middleware.validator.go
	Username string `json:"username" form:"username" validate:"required,is_valid_username" example:"admin"`
	Password string `json:"password" form:"password" validate:"required" example:"password"`
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token"`
}

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UsernName string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}

//
func (param *AdminLoginInput) BindParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
