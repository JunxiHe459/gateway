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
)

type DashboardController struct{}

func DashboardRegister(group *gin.RouterGroup) {
	service := &DashboardController{}
	group.GET("/panel_group_data", service.PanelGroupData)
	group.GET("/flow_stat", service.FlowStat)
	group.GET("/service_stat", service.ServiceStat)
}

// PanelGroupData godoc
// @Summary General Stats
// @Description 指标统计
// @Tags Dashboard
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (service *DashboardController) PanelGroupData(c *gin.Context) {
	serviceInfo := dao.ServiceInfo{}
	_, total_service, err := serviceInfo.GetPageList(c, global.DB,
		&dto.ServiceListInput{
			PageNumber: 1,
			PageSize:   1})
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	renterInfo := dao.Renter{}
	_, total_renter, err := renterInfo.GetRenterList(c, global.DB, &dto.RenterListInput{PageNumber: 1, PageSize: 1})

	out := &dto.PanelGroupDataOutput{
		ServiceNum:      total_service,
		RenterNum:       total_renter,
		CurrentQPS:      0, // TODO: 之后添加
		TodayRequestNum: 0, // TODO： 之后添加
	}
	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary Flow Stats
// @Description 流量统计
// @Tags Dashboard
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatsOutput} "success"
// @Router /dashboard/flow_stat [get]
func (service *DashboardController) FlowStat(c *gin.Context) {
	//counter, err := public.FlowCounterHandler.GetCounter(public.FlowTotal)
	//if err != nil {
	//	middleware.ResponseError(c, 2001, err)
	//	return
	//}
	//todayList := []int{}
	//currentTime := time.Now()
	//for i := 0; i <= currentTime.Hour(); i++ {
	//	dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
	//	hourData, _ := counter.GetHourData(dateTime)
	//	todayList = append(todayList, int(hourData))
	//}
	//
	//yesterdayList := []int{}
	//yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	//for i := 0; i <= 23; i++ {
	//	dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, lib.TimeLocation)
	//	hourData, _ := counter.GetHourData(dateTime)
	//	yesterdayList = append(yesterdayList, int(hourData))
	//}
	//middleware.ResponseSuccess(c, &dto.ServiceStatsOutput{
	//	Today:     todayList,
	//	Yesterday: yesterdayList,
	//})
	middleware.ResponseSuccess(c, &dto.ServiceStatsOutput{})
}

//ServiceStat godoc
//@Summary Service Stats
//@Description 服务统计
//@Tags Dashboard
//@ID /dashboard/service_stat
//@Accept  json
//@Produce  json
//@Success 200 {object} middleware.Response{data=dto.DashServiceStatOutput} "success"
//@Router /dashboard/service_stat [get]
func (service *DashboardController) ServiceStat(c *gin.Context) {
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	list, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	legend := []string{}
	for index, item := range list {
		name, ok := public.LoadTypeMap[item.LoadType]
		if !ok {
			middleware.ResponseError(c, 2003, errors.New("load_type not found"))
			return
		}
		list[index].Name = name
		legend = append(legend, name)
	}

	out := &dto.DashServiceStatOutput{
		Legend: legend,
		Data:   list,
	}
	middleware.ResponseSuccess(c, out)
}
