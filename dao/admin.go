package dao

import (
	"errors"
	"github.com/JunxiHe459/gateway/dto"
	"github.com/JunxiHe459/gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
)

type Admin struct {
	Id        int    `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string `json:"user_name" gorm:"column:user_name" description:"管理员用户名"`
	Salt      string `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt string `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt string `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int    `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (admin *Admin) TableName() string {
	return "gateway_admin"
}

func (admin *Admin) Find(c *gin.Context, db *gorm.DB, search *Admin) (*Admin, error) {
	out := &Admin{}
	err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(out).Error
	if err != nil {
		print(err.Error())
		return nil, err
	}
	return out, nil
}

func (admin *Admin) Save(c *gin.Context, db *gorm.DB) error {
	err := db.SetCtx(public.GetGinTraceContext(c)).Save(admin).Error
	if err != nil {
		print(err.Error())
		return err
	}
	return nil
}

func (admin *Admin) LoginAndCheck(c *gin.Context, db *gorm.DB,
	param *dto.AdminLoginInput) (*Admin, error) {

	// 得到数据库中，username 对应的 password
	adminInfo, err := admin.Find(c, db, &Admin{
		UserName: param.Username,
		IsDelete: 0,
	})
	if err != nil {
		return nil, errors.New("Unable to find user")
	}

	// 校验 password
	salted_password := public.SaltPassword(adminInfo.Salt, param.Password)
	if salted_password != adminInfo.Password {
		//println("Admin Password:", adminInfo.Password)
		//println("Salted_Password:", salted_password)
		return nil, errors.New("Password does not match. Please enter your password again.")
	}

	return adminInfo, nil
}
