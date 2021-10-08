package dto

import (
	"github.com/JunxiHe459/gateway/public"
	"github.com/gin-gonic/gin"
)

type ServiceListInput struct {
	Keyword    string `json:"keyword" form:"keyword" validate:""`
	PageNumber int    `json:"page_number" form:"page_number" validate:"required"`
	PageSize   int    `json:"page_size" form:"page_size" validate:"required"`
}

type SingleService struct {
	ID             int64  `json:"id" form:"id"`                        //id
	ServiceName    string `json:"service_name" form:"service_name"`    //服务名称
	ServiceDesc    string `json:"service_desc" form:"service_desc"`    //服务描述
	LoadType       int    `json:"load_type" form:"load_type"`          //类型
	ServiceAddress string `json:"service_address" form:"service_addr"` //服务地址
	QPS            int64  `json:"qps" form:"qps"`                      //qps
	QPD            int64  `json:"qpd" form:"qpd"`                      //qpd
	TotalNode      int    `json:"total_node" form:"total_node"`        //节点数
}

type ServiceListOutput struct {
	TotalPage   int             `json:"total_page" form:"total_page" validate:"required"`
	ServiceList []SingleService `json:"single_service" form:"single_service" validate:"required"`
}

func (param *ServiceListInput) BindParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
