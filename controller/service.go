package controller

import (
	"fmt"
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
)

type ServiceController struct {
}

func RegiterService(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("delete", service.DeleteService)
	group.POST("add_http", service.AddHTTPService)
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
	// 拿到参数 第几页 每几个 关键词
	params := &dto.ServiceListInput{}
	err := params.BindParam(c)
	if err != nil {
		println("ServiceList bind params error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	// 从数据库拿到 []serviceInfo
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.GetPageList(c, global.DB, params)
	if err != nil {
		println("Get serviceList error: ", err.Error())
		middleware.ResponseError(c, 2001, err)
		return
	}

	// 将 []serviceInfo 转换成 []singleService
	// 根据 serviceInfo 的 id，查 singleService
	var serviceList []dto.SingleService
	for _, item := range list {
		serviceDetail, err := item.GetServiceDetail(c, global.DB, &item)
		if err != nil {
			println("Get service detail error: ", err.Error())
			middleware.ResponseError(c, 2002, err)
			return
		}

		// 1. Http 后缀接入   clusterIP + clusterPort + path
		// 2. Http 域名接入 	 domain
		// 3. Tcp / Grpc 接入 clusterIP + servicePort
		serviceAddress := "unknown"
		clusterIp := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		SSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP {
			// HTTP 后缀接入 需要 Prefix URL 和 HTTPS
			if serviceDetail.HTTPRule.RuleType == public.HTTPPrefixURL &&
				serviceDetail.HTTPRule.NeedHttps == 1 {
				serviceAddress = fmt.Sprintf("%s:%s%s", clusterIp, SSLPort, serviceDetail.HTTPRule.Rule)
			}

			// HTTP 后缀接入  需要 Prefix URL 不需要 HTTPS
			if serviceDetail.HTTPRule.RuleType == public.HTTPPrefixURL &&
				serviceDetail.HTTPRule.NeedHttps == 0 {
				serviceAddress = fmt.Sprintf("%s:%s%s", clusterIp, clusterPort, serviceDetail.HTTPRule.Rule)
			}

			// HTTP 域名接入
			if serviceDetail.HTTPRule.RuleType == public.HTTPDomain {
				serviceAddress = serviceDetail.HTTPRule.Rule
			}
		}

		// TCP
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddress = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.TCPRule.Port)
		}

		// gRPC
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddress = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.GRPCRule.Port)
		}

		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		singleService := dto.SingleService{
			ID:             item.ID,
			ServiceName:    item.ServiceName,
			ServiceDesc:    item.ServiceDesc,
			ServiceAddress: serviceAddress,
			QPS:            0,
			QPD:            0,
			TotalNode:      len(ipList),
		}
		serviceList = append(serviceList, singleService)
	}

	// 绑定响应
	out := &dto.ServiceListOutput{
		TotalServices: total,
		ServiceList:   serviceList,
	}
	middleware.ResponseSuccess(c, out)
}

// Service godoc
// @Summary Delete service
// @Description 删除一个服务
// @Tags 服务管理
// @ID /service/delete
// @Accept json
// @Produce json
// @Param id query int true "服务 ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/delete [GET]
func (s *ServiceController) DeleteService(c *gin.Context) {
	params := dto.ServiceDeleteInput{}
	err := params.BindParam(c)
	if err != nil {
		println("Bind Params Error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	// Find Service
	service, err = service.Find(c, global.DB, service)
	if err != nil {
		println("Find Service Error: ", err.Error())
		middleware.ResponseError(c, 2000, err)
		return
	}

	// 硬删除
	//err = service.Delete(c, global.DB)
	//if err != nil{
	//	println("Delete Service Error: ", err.Error())
	//	middleware.ResponseError(c, 2001, err)
	//}

	// 软删除
	service.IsDelete = 1
	err = service.Save(c, global.DB)
	if err != nil {
		println("Save Service Error: ", err.Error())
		middleware.ResponseError(c, 2001, err)
		return
	}

	middleware.ResponseSuccess(c, "Deleted")
}

// Service godoc
// @Summary Add a new HTTP Service
// @Description 添加一个新的 HTTP 服务
// @Tags 服务管理
// @ID /service/add_http
// @Accept json
// @Produce json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/add_http [POST]
func (s *ServiceController) AddHTTPService(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	err := params.BindParam(c)
	if err != nil {
		if err != nil {
			println("ServiceList bind params error: ", err.Error())
			middleware.ResponseError(c, 400, err)
			return
		}
	}

}
