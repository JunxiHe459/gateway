package controller

import (
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/gin-gonic/gin"
)

type ServiceController struct {
}

func RegiterService(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
}

// Service godoc
// @Summary Service List
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept json
// @Produce json
// @Param keyword query string false "关键词"
// @Param page_size query int true "每页个数"
// @param page_number query int true "总页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [GET]
func (s *ServiceController) ServiceList(c *gin.Context) {
	params := &dto.ServiceListInput{}
	err := params.BindParam(c)
	if err != nil {
		println("ServiceList bind params error: ", err.Error())
		middleware.ResponseError(c, 400, err)
	}

	out := &dto.ServiceListOutput{}
	middleware.ResponseSuccess(c, out)
}
