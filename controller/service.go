package controller

import (
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
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
	// 绑定参数
	params := &dto.ServiceListInput{}
	err := params.BindParam(c)
	if err != nil {
		println("ServiceList bind params error: ", err.Error())
		middleware.ResponseError(c, 400, err)
	}

	// 从数据库拿到 []serviceInfo
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.GetPageList(c, global.DB, params)
	if err != nil {
		println("Get serviceList error: ", err.Error())
		middleware.ResponseError(c, 2001, err)
	}

	// 将 []serviceInf 转换成 []singleService
	var serviceList []dto.SingleService
	for _, item := range list {
		singleService := dto.SingleService{
			ID:          item.ID,
			ServiceName: item.ServiceName,
			ServiceDesc: item.ServiceDesc,
		}
		serviceList = append(serviceList, singleService)
	}

	// 绑定响应
	out := &dto.ServiceListOutput{
		total,
		serviceList,
	}
	middleware.ResponseSuccess(c, out)
}
