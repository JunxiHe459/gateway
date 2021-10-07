package controller

import (
	"encoding/json"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminInfoController struct {
}

func RegiterAdminInfo(group *gin.RouterGroup) {
	admin := &AdminInfoController{}
	group.GET("/", admin.AdminInfo)
}

// AdminLogin godoc
// @Summary Admin Login
// @Description Admin Login 接口
// @Tags 管理员接口
// @ID /admin/info
// @Accept json
// @Produce json
// @Sucess 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
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
