package controller

import (
	"encoding/json"
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type AdminLoginController struct {
}

func RegiterAdmin(group *gin.RouterGroup) {
	admin := &AdminLoginController{}
	group.POST("/login", admin.AdminLogin)
	group.GET("/logout", admin.AdminLogout)
}

// AdminLogin godoc
// @Summary Admin Login
// @Description Admin Login 接口
// @Tags Admin
// @ID /admin/login
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin/login [POST]
func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
	println("Admin login requested")
	params := &dto.AdminLoginInput{}
	err := params.BindParam(c)
	if err != nil {
		// 如果出现错误，就返回 400 状态码
		print("Line 20 of admin_log_in: ", err.Error())
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// 正常情况下
	// 验证密码是否正确
	admin := &dao.Admin{}
	admin, err = admin.LoginAndCheck(c, global.DB, params)
	if err != nil {
		print("LoginAndCheck failed: ", err.Error())
		// 2000 是自定义错误码
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 设置 session
	adminSession := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UsernName: params.Username,
		LoginTime: time.Now(),
	}
	adminSessionBinary, err := json.Marshal(adminSession)
	if err != nil {
		print("LoginAndCheck json marshal failed: ", err.Error())
		middleware.ResponseError(c, 2001, err)
		return
	}
	s := sessions.Default(c)
	s.Set(public.AdminSessionInfoKey, string(adminSessionBinary))
	// 存到 redis 里面
	_ = s.Save()

	middleware.ResponseSuccess(c, &dto.AdminLoginOutput{Token: params.Username})
	return
}

// AdminLogin godoc
// @Summary Admin Log out
// @Description Admin Log out 接口
// @Tags Admin
// @ID /admin/logout
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/logout [GET]
func (admin *AdminLoginController) AdminLogout(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete(public.AdminSessionInfoKey)
	_ = s.Save()

	middleware.ResponseSuccess(c, "Log out!")
}
