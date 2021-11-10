package controller

import (
	"github.com/JunxiHe459/gateway/dao"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/middleware"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

//RenterControllerRegister admin路由注册
func RenterRegister(router *gin.RouterGroup) {
	admin := APPController{}
	router.GET("/renter_list", admin.RenterList)
	router.GET("/renter_detail", admin.RenterDetail)
	router.GET("/renter_stats", admin.RenterStats)
	router.GET("/delete_renter", admin.DeleteRenter)
	router.POST("/add_renter", admin.AddRenter)
	router.POST("/update_renter", admin.UpdateRenter)
}

type APPController struct {
}

// RenterList godoc
// @Summary Renter list
// @Description 租户列表
// @Tags Renter Management
// @ID /renter/renter_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query string true "每页多少条"
// @Param page_no query string true "页码"
// @Success 200 {object} middleware.Response{data=dto.RenterListOutput} "success"
// @Router /renter/renter_list [get]
func (admin *APPController) RenterList(c *gin.Context) {
	params := &dto.RenterListInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	info := &dao.Renter{}
	list, total, err := info.GetRenterList(c, global.DB, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	outputList := []dto.RenterListItemOutput{}
	for _, item := range list {
		appCounter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + item.RenterID)
		if err != nil {
			middleware.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		outputList = append(outputList, dto.RenterListItemOutput{
			ID:       item.ID,
			AppID:    item.RenterID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.Qpd,
			Qps:      item.Qps,
			RealQpd:  appCounter.TotalCount,
			RealQps:  appCounter.QPS,
		})
	}
	output := dto.RenterListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(c, output)
	return
}

// RenterDetail godoc
// @Summary Renter Info
// @Description 租户详情
// @Tags Renter Management
// @ID /renter/renter_detail
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.Renter} "success"
// @Router /renter/renter_detail [get]
func (admin *APPController) RenterDetail(c *gin.Context) {
	params := &dto.RenterDetailInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.Renter{
		ID: params.ID,
	}
	detail, err := search.Find(c, global.DB, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	middleware.ResponseSuccess(c, detail)
	return
}

// DeleteRenter godoc
// @Summary Delete Renter
// @Description 租户删除
// @Tags Renter Management
// @ID /renter/delete_renter
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /renter/delete_renter [get]
func (admin *APPController) DeleteRenter(c *gin.Context) {
	params := &dto.RenterDetailInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.Renter{
		ID: params.ID,
	}
	info, err := search.Find(c, global.DB, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	info.IsDelete = 1
	if err := info.Save(c, global.DB); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// AddRenter godoc
// @Summary Add Renter
// @Description 租户添加
// @Tags Renter Management
// @ID /renter/add_renter
// @Accept  json
// @Produce  json
// @Param body body dto.AddRenterHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /renter/add_renter [post]
func (admin *APPController) AddRenter(c *gin.Context) {
	params := &dto.AddRenterHttpInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证app_id是否被占用
	search := &dao.Renter{
		RenterID: params.RenterID,
	}
	if _, err := search.Find(c, global.DB, search); err == nil {
		middleware.ResponseError(c, 2002, errors.New("租户ID被占用，请重新输入"))
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.RenterID)
	}
	tx := global.DB
	info := &dao.Renter{
		RenterID: params.RenterID,
		Name:     params.Name,
		Secret:   params.Secret,
		WhiteIPS: params.WhiteIPS,
		Qps:      params.Qps,
		Qpd:      params.Qpd,
	}
	if err := info.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// UpdateRenter godoc
// @Summary Update Renter
// @Description 租户更新
// @Tags Renter Management
// @ID /renter/update_renter
// @Accept  json
// @Produce  json
// @Param body body dto.UpdateRenterHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /renter/update_renter [post]
func (admin *APPController) UpdateRenter(c *gin.Context) {
	params := &dto.UpdateRenterHttpInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	search := &dao.Renter{
		ID: params.ID,
	}
	info, err := search.Find(c, global.DB, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	if params.Secret == "" {
		params.Secret = public.MD5(params.RenterID)
	}
	info.Name = params.Name
	info.Secret = params.Secret
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(c, global.DB); err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}
	middleware.ResponseSuccess(c, "")
	return
}

// RenterStats godoc
// @Summary Renter Stats
// @Description 租户统计
// @Tags Renter Management
// @ID /renter/renter_stats
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /renter/renter_stats [get]
func (admin *APPController) RenterStats(c *gin.Context) {
	params := &dto.RenterDetailInput{}
	if err := params.BindParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	search := &dao.Renter{
		ID: params.ID,
	}
	detail, err := search.Find(c, global.DB, search)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//今日流量全天小时级访问统计
	todayStat := []int64{}
	counter, err := public.FlowCounterHandler.GetCounter(public.FlowAppPrefix + detail.RenterID)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		c.Abort()
		return
	}
	currentTime := time.Now()
	for i := 0; i <= time.Now().In(lib.TimeLocation).Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		todayStat = append(todayStat, hourData)
	}

	//昨日流量全天小时级访问统计
	yesterdayStat := []int64{}
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := counter.GetHourData(dateTime)
		yesterdayStat = append(yesterdayStat, hourData)
	}
	stat := dto.StatisticsOutput{
		Today:     todayStat,
		Yesterday: yesterdayStat,
	}
	middleware.ResponseSuccess(c, stat)
	return
}
