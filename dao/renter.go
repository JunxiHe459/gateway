package dao

import (
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"sync"
	"time"
)

type Renter struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	RenterID  string    `json:"renter_id" gorm:"column:renter_id" description:"租户id"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (t *Renter) TableName() string {
	return "gateway_renter"
}

func (t *Renter) Find(c *gin.Context, tx *gorm.DB, search *Renter) (*Renter, error) {
	model := &Renter{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *Renter) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}

func (t *Renter) GetRenterList(c *gin.Context, tx *gorm.DB, params *dto.RenterListInput) ([]Renter, int64, error) {
	var list []Renter
	var count int64
	pageNo := params.PageNumber
	pageSize := params.PageSize

	//limit offset,pagesize
	offset := (pageNo - 1) * pageSize
	query := tx.SetCtx(public.GetGinTraceContext(c))
	query = query.Table(t.TableName()).Select("*")
	query = query.Where("is_delete=?", 0)
	if params.Info != "" {
		query = query.Where(" (name like ? or renter_id like ?)", "%"+params.Info+"%", "%"+params.Info+"%")
	}
	err := query.Limit(pageSize).Offset(offset).Order("id desc").Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	errCount := query.Count(&count).Error
	if errCount != nil {
		return nil, 0, err
	}
	return list, count, nil
}

var RenterManagerHandler *RenterManager

func init() {
	RenterManagerHandler = NewRenterManager()
}

type RenterManager struct {
	RenterMap  map[string]*Renter
	RenterList []*Renter
	Locker     sync.RWMutex
	init       sync.Once
	err        error
}

func NewRenterManager() *RenterManager {
	return &RenterManager{
		RenterMap:  map[string]*Renter{},
		RenterList: []*Renter{},
		Locker:     sync.RWMutex{},
		init:       sync.Once{},
	}
}

func (s *RenterManager) GetRenterList() []*Renter {
	return s.RenterList
}

func (s *RenterManager) LoadOnce() error {
	s.init.Do(func() {
		renterInfo := &Renter{}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			s.err = err
			return
		}
		params := &dto.RenterListInput{PageNumber: 1, PageSize: 99999}
		list, _, err := renterInfo.GetRenterList(c, tx, params)
		if err != nil {
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _, listItem := range list {
			tmpItem := listItem
			s.RenterMap[listItem.RenterID] = &tmpItem
			s.RenterList = append(s.RenterList, &tmpItem)
		}
	})
	return s.err
}
