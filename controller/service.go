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
	}

	// 从数据库拿到 []serviceInfo
	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.GetPageList(c, global.DB, params)
	if err != nil {
		println("Get serviceList error: ", err.Error())
		middleware.ResponseError(c, 2001, err)
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
		cluster_ip := lib.GetStringConf("base.cluster.cluster_ip")
		cluster_port := lib.GetStringConf("base.cluster.cluster_port")
		cluster_SSL_port := lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP {
			// HTTP 后缀接入 需要 Prefix URL 和 HTTPS
			if serviceDetail.HTTPRule.RuleType == public.HTTPPrefixURL &&
				serviceDetail.HTTPRule.NeedHttps == 1 {
				serviceAddress = cluster_ip + cluster_SSL_port + serviceDetail.HTTPRule.Rule
			}

			// HTTP 后缀接入  需要 Prefix URL 不需要 HTTPS
			if serviceDetail.HTTPRule.RuleType == public.HTTPPrefixURL &&
				serviceDetail.HTTPRule.NeedHttps == 0 {
				serviceAddress = cluster_ip + cluster_port + serviceDetail.HTTPRule.Rule
			}

			// HTTP 域名接入
			if serviceDetail.HTTPRule.RuleType == public.HTTPDomain {
				serviceAddress = serviceDetail.HTTPRule.Rule
			}
		}

		// TCP
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddress = fmt.Sprintf("%s:%d", cluster_ip, serviceDetail.TCPRule.Port)
		}

		// gRPC
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddress = fmt.Sprintf("%s:%d", cluster_ip, serviceDetail.GRPCRule.Port)
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
		total,
		serviceList,
	}
	middleware.ResponseSuccess(c, out)
}
