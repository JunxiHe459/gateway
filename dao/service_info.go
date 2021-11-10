package dao

import (
	"fmt"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/global"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type ServiceInfo struct {
	ID          int64 `json:"id" gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LoadType    int    `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	IsDelete    int8   `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (service *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (service *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	fmt.Printf("%+v\n", search)
	err := db.SetCtx(public.GetGinTraceContext(c)).Where("is_delete = ? AND service_name = ?", 0, search.ServiceName).First(out).Error
	fmt.Printf("%+v\n", out)
	if err != nil {
		println("hahha")
		print(err.Error())
		return nil, err
	}
	return out, nil
}

func (service *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	err := db.SetCtx(public.GetGinTraceContext(c)).Save(service).Error
	if err != nil {
		print(err.Error())
		return err
	}
	return nil
}

// 硬删除，最好别用
func (service *ServiceInfo) Delete(c *gin.Context, db *gorm.DB) error {
	err := db.SetCtx(public.GetGinTraceContext(c)).Delete(service).Error
	if err != nil {
		print(err.Error())
		return err
	}
	return nil
}

func (service *ServiceInfo) GetPageList(c *gin.Context, db *gorm.DB, param *dto.ServiceListInput) (list []ServiceInfo, total int64, err error) {
	offset := (param.PageNumber - 1) * param.PageNumber
	query := db.SetCtx(public.GetGinTraceContext(c)).Table(service.TableName()).Where("is_delete=0")

	if param.Keyword != "" {
		query = query.Where("service_name like ? or service_desc like ?", "%"+param.Keyword+"%", "%"+param.Keyword+"%")
	}

	// 如果数据库为空，则 err 会是 ErrRecordNotFound，这是没问题的
	// 如果 err 不是 ErrRecordNotFound 才证明是真的有问题，所以下面进行一个判断
	err = query.Offset(offset).Limit(param.PageSize).Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	print("Length of list", len(list))
	query.Offset(offset).Limit(param.PageSize).Count(&total)

	return
}

func (service *ServiceInfo) GetServiceDetail(c *gin.Context, db *gorm.DB, info *ServiceInfo) (detail *ServiceDetail, err error) {
	if info.ServiceName == "" {
		search, err := service.Find(c, global.DB, info)
		if err != nil {
			return nil, err
		}
		info = search
	}

	http := &HttpRule{ServiceID: info.ID}
	http, err = http.Find(c, db, http)
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	tcp := &TcpRule{ServiceID: info.ID}
	tcp, err = tcp.Find(c, db, tcp)
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	grpc := &GrpcRule{ServiceID: info.ID}
	grpc, err = grpc.Find(c, db, grpc)
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	access := &AccessControl{ServiceID: info.ID}
	access, err = access.Find(c, db, access)
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	loadbalance := &LoadBalance{ServiceID: info.ID}
	loadbalance, err = loadbalance.Find(c, db, loadbalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return
	}

	detail = &ServiceDetail{
		Info:          info,
		HTTPRule:      http,
		TCPRule:       tcp,
		GRPCRule:      grpc,
		AccessControl: access,
		LoadBalance:   loadbalance,
	}

	return
}

func (service *ServiceInfo) GroupByLoadType(c *gin.Context, db *gorm.DB) (list []dto.DashServiceStatItemOutput, err error) {
	err = db.SetCtx(public.GetGinTraceContext(c)).Table(service.TableName()).Where(
		"is_delete=0").Select("load_type, count(*) as value").Group(
		"load_type").Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, err
}
