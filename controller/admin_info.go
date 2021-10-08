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
)

type AdminInfoController struct {
}

func RegiterAdminInfo(group *gin.RouterGroup) {
	admin := &AdminInfoController{}
	group.GET("", admin.AdminInfo)
	group.POST("/change_password", admin.ChangePassword)
}

// AdminInfo godoc
// @Summary Admin Info
// @Description Admin Info 接口
// @Tags 管理员接口
// @ID /admin/info
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin/info [GET]
func (adminInfo *AdminInfoController) AdminInfo(c *gin.Context) {
	s := sessions.Default(c)
	sessionInfo := s.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	err := json.Unmarshal([]byte(sessionInfo.(string)), adminSessionInfo)
	if err != nil {
		print("Json Unmarshall Err: ", err.Error())
		middleware.ResponseError(c, 2000, err)
		return
	}

	out := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		UsernName:    adminSessionInfo.UsernName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Introduction: "A super administrator",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c, out)
}

// AdminInfo godoc
// @Summary Admin Change Password
// @Description Admin 修改密码接口
// @Tags 管理员接口
// @ID /admin/info/change_password
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin/info [POST]
func (adminInfo *AdminInfoController) ChangePassword(c *gin.Context) {
	// 读取 POST 传过来的信息
	params := &dto.ChangePasswordInput{}
	err := params.BindParam(c)
	if err != nil {
		// 如果出现错误，就返回 400 状态码
		print("Line 20 of admin_log_in: ", err.Error())
		middleware.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	// 从 Session 中得到用户名
	s := sessions.Default(c)
	sessionInfo := s.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	err = json.Unmarshal([]byte(sessionInfo.(string)), adminSessionInfo)
	if err != nil {
		print("Json Unmarshall Err: ", err.Error())
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 根据用户名找到这个用户
	admin := &dao.Admin{}
	admin, err = admin.Find(c, global.DB, &dao.Admin{
		UserName: adminSessionInfo.UsernName,
	})
	if err != nil {
		print("Admin.Find Error: ", err.Error())
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 得到加盐后的密码
	pwd := public.SaltPassword(admin.Salt, params.Password)
	admin.Password = pwd
	err = admin.Save(c, global.DB)
	if err != nil {
		print("Admin.Save Error: ", err.Error())
		middleware.ResponseError(c, 2002, err)
		return
	}

	middleware.ResponseSuccess(c, "Changed password successfully")
}
