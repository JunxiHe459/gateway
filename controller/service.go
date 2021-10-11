package controller

import (
	"errors"
	"fmt"
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"strings"
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

	tx := global.DB
	tx.Begin()
	// 查看有没有重复的 ServiceName
	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err == nil {
		tx.Rollback()
		println("Service Already Exists: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Service Already Exists"))
		return
	}

	// 查看有无重复前缀 或 域名
	httpService := &dao.HttpRule{
		RuleType: params.RuleType,
		Rule:     params.Rule,
	}
	_, err = httpService.Find(c, tx, httpService)
	if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		println("Http url or domain name Already Exists: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Http url or domain name Already Exists"))
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		tx.Rollback()
		println("IP list should have same length as weight list: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("IP list should have same length as weight list"))
		return
	}

	// 创建 service 本体到 serviceinfo 中
	serviceinfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	err = serviceinfo.Save(c, tx)
	if err != nil {
		tx.Rollback()
		println("Save service info error: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Save service info error"))
		return
	}

	// id 作为 foreign key 去关联其他表
	id := serviceinfo.ID

	// 关联 http_rule
	httpRule := &dao.HttpRule{
		ServiceID:      id,
		RuleType:       params.RuleType,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfer: params.HeaderTransfer,
	}
	err = httpRule.Save(c, tx)
	if err != nil {
		tx.Rollback()
		println("Save http rule error: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Save http rule error"))
		return
	}

	// 关联 access_control
	accessControl := &dao.AccessControl{
		ServiceID:         id,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	err = accessControl.Save(c, tx)
	if err != nil {
		tx.Rollback()
		println("Save access conrol error: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Save access conrol error"))
		return
	}

	// 关联 load_balance
	loadbalance := &dao.LoadBalance{
		ServiceID:              id,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	err = loadbalance.Save(c, tx)
	if err != nil {
		tx.Rollback()
		println("Save load balance error: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Save load balance error"))
		return
	}

	// 提交事务
	tx.Commit()
	middleware.ResponseSuccess(c, "New HTTP serviced added")
}
