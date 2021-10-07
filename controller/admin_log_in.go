package controller

import (
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminLoginController struct {
}

func RegiterAdminLogin(group *gin.RouterGroup) {
	admin := &AdminLoginController{}
	group.POST("/login", admin.AdminLogin)
}

// AdminLogin godoc
// @Summary Admin Login
// @Description Admin Login 接口
// @Tags 管理员登陆接口
// @ID /admin/login
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Sucess 200 {object{ middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin/login [post]
func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
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

	middleware.ResponseSuccess(c, &dto.AdminLoginOutput{Token: params.Username})
	return
}
