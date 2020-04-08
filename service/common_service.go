package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"dubbo-gateway/meta"
	"dubbo-gateway/service/entry"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/jinzhu/gorm"
)

var UserOrPasswordError = errors.New("username or password error")

type CommonService interface {
	GetUser(userName, password string) (*entry.User, error)
	CreateUser(user *entry.User) error
	UpdatePassword(user *entry.User, oldPassword string) error
}

type commonService struct {
	*gorm.DB
}

func NewCommonService() CommonService {
	return &commonService{meta.GetDB()}
}

func (c *commonService) CreateUser(user *entry.User) error {
	user.Password = Sha256(user.Password, user.Name)
	return c.Save(user).Error
}

func (c *commonService) UpdatePassword(user *entry.User, oldPassword string) error {
	dbUser, err := c.GetUser(user.Name, oldPassword)
	if err != nil {
		return err
	}
	return c.Model(&entry.User{}).Where("name = ?", dbUser.Name).
		UpdateColumn("password", Sha256(user.Password, dbUser.Name)).Error
}

func (c *commonService) GetUser(userName, password string) (*entry.User, error) {
	user := new(entry.User)
	err := c.Where("name = ? and password = ?", userName, password).Find(user).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, UserOrPasswordError
	}
	return user, err
}

func Sha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sha := hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(sha))
}
