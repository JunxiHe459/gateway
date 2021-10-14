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
	"time"
)

type ServiceController struct {
}

func RegisterService(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("service_list", service.ServiceList)
	group.GET("delete", service.DeleteService)

	group.POST("add_http", service.AddHTTPService)
	group.POST("update_http", service.UpdateHTTPService)
	group.POST("add_tcp", service.ServiceAddTcp)
	group.POST("update_tcp", service.ServiceUpdateTcp)
	group.POST("add_grpc", service.ServiceAddGRPC)
	group.POST("update_grpc", service.ServiceUpdateGRPC)

	group.GET("service_details", service.ServiceDetail)
	group.GET("service_stats", service.ServiceStats)
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

	// 获取与数据库的连接 因为需要事务，所以不用 global.db
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()

	// 查看有没有重复的 ServiceName
	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != gorm.ErrRecordNotFound {
		//gorm.ErrRecordNotFound
		tx.Rollback()
		println("Service Already Exists")
		middleware.ResponseError(c, 400, errors.New("Service Already Exists"))
		return
	}

	// 查看有无重复前缀 或 域名
	httpService := &dao.HttpRule{
		RuleType: params.RuleType,
		Rule:     params.Rule,
	}
	_, err = httpService.Find(c, tx, httpService)
	if err == nil {
		tx.Rollback()
		println("Http url or domain name Already Exists: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Http url or domain name Already Exists"))
		return
	}

	println("Err is: ", err.Error())

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
		println("Save access control error: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("Save access control error"))
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

// Service godoc
// @Summary Update an existing HTTP Service
// @Description 更新一个 HTTP 服务
// @Tags 服务管理
// @ID /service/update_http
// @Accept json
// @Produce json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/add_http [POST]
func (s *ServiceController) UpdateHTTPService(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	err := params.BindParam(c)
	if err != nil {
		if err != nil {
			println("ServiceList bind params error: ", err.Error())
			middleware.ResponseError(c, 400, err)
			return
		}
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		println("IP list should have same length as weight list: ", err.Error())
		middleware.ResponseError(c, 400, errors.New("IP list should have same length as weight list"))
		return
	}

	// 获取与数据库的连接 因为需要事务，所以不用 global.db
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	tx = tx.Begin()

	// 查看有没有 ServiceName
	serviceInfo := &dao.ServiceInfo{
		ID: params.ID,
	}

	serviceDetail, err := serviceInfo.GetServiceDetail(c, tx, serviceInfo)
	if err != nil {
		//gorm.ErrRecordNotFound
		tx.Rollback()
		println("Service Not Exists")
		middleware.ResponseError(c, 400, errors.New("Service Not Exists"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	info.ServiceName = params.ServiceName
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		println("Save http rule error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.RuleType = params.RuleType
	httpRule.Rule = params.Rule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfer = params.HeaderTransfer
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		println("Save http rule error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		println("Save access control error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	loadbalance := serviceDetail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		println("Save load balance error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}

	// 提交事务
	tx.Commit()
	middleware.ResponseSuccess(c, "HTTP service updated")
}

// Service godoc
// @Summary Details of a service
// @Description 服务详情
// @Tags 服务管理
// @ID /service/service_details
// @Accept json
// @Produce json
// @Param ID query int true "ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_details [GET]
func (service *ServiceController) ServiceDetail(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 400, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//读取基本信息
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		println("Find Service Info error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}
	serviceDetail, err := serviceInfo.GetServiceDetail(c, tx, serviceInfo)
	if err != nil {
		println("Get service detail error: ", err.Error())
		middleware.ResponseError(c, 400, err)
		return
	}
	middleware.ResponseSuccess(c, serviceDetail)
}

// Service godoc
// @Summary Network Flow statistics
// @Description 流量数据
// @Tags 服务管理
// @ID /service/service_detail
// @Accept json
// @Produce json
// @Param ID query int true "ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatsOutput} "success"
// @Router /service/service_stats [GET]
func (service *ServiceController) ServiceStats(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	err := params.BindParam(c)
	if err != nil {
		middleware.ResponseError(c, 400, err)
		return
	}

	//读取基本信息
	//serviceInfo := &dao.ServiceInfo{ID: params.ID}
	//serviceInfo, err = serviceInfo.Find(c, global.DB, serviceInfo)
	//if err != nil {
	//	println("Find Service Info error: ", err.Error())
	//	middleware.ResponseError(c, 400, err)
	//	return
	//}
	//serviceDetail, err := serviceInfo.GetServiceDetail(c, global.DB, serviceInfo)
	//if err != nil {
	//	println("Get service detail error: ", err.Error())
	//	middleware.ResponseError(c, 400, err)
	//	return
	//}

	var today []int
	for i := 0; i <= time.Now().Hour(); i++ {
		today = append(today, 0)
	}
	var yesterday [24]int

	middleware.ResponseSuccess(c, &dto.ServiceStatsOutput{
		Today:     today,
		Yesterday: yesterday,
	})
}

// ServiceAddHttp godoc
// @Summary Add a new TCP service
// @Description tcp服务添加
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/add_tcp [post]
func (admin *ServiceController) ServiceAddTcp(c *gin.Context) {
	params := &dto.ServiceAddTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	httpRule := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary Update an existing TCP service
// @Description tcp服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_tcp [post]
func (admin *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.GetServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TCPRule != nil {
		tcpRule = detail.TCPRule
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceAddHttp godoc
// @Summary Add a new GRPC service
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/add_grpc [post]
func (admin *ServiceController) ServiceAddGRPC(c *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfer: params.HeaderTransfer,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary Update an existing GRPC service
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/update_grpc [post]
func (admin *ServiceController) ServiceUpdateGRPC(c *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.GetServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPCRule != nil {
		grpcRule = detail.GRPCRule
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfer = params.HeaderTransfer
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}
