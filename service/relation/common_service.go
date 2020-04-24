package relation

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"github.com/jinzhu/gorm"
)

type commonService struct {
	*gorm.DB
}

func NewCommonService(db *gorm.DB) service.CommonService {
	return &commonService{db}
}

func (c *commonService) CreateUser(user *entry.User) error {
	user.Password = utils.Sha256(user.Password, user.Name)
	return c.Save(user).Error
}

func (c *commonService) UpdatePassword(user *entry.User, oldPassword string) error {
	dbUser, err := c.GetUser(user.Name, oldPassword)
	if err != nil {
		return err
	}
	return c.Model(&entry.User{}).Where("name = ?", dbUser.Name).
		UpdateColumn("password", utils.Sha256(user.Password, dbUser.Name)).Error
}

func (c *commonService) GetUser(userName, password string) (*entry.User, error) {
	user := new(entry.User)
	err := c.Where("name = ? and password = ?", userName, utils.Sha256(password, userName)).Find(user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, constant.UserOrPasswordError
	}
	return user, err
}
