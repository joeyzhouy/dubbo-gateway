package meta

import (
	"dubbo-gateway/common/constant"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const DefaultPath = "meta/gateway.yml"

var db *gorm.DB

type DbConfig struct {
	Config struct {
		Dialect     string        `yaml:"db_dialect"`
		Config      string        `yaml:"db_config"`
		MaxOpen     int           `yaml:"db_maxopen"`
		MaxIdle     int           `yaml:"db_maxidle"`
		MaxLifeTime time.Duration `yaml:"db_maxlifetime"`
	} `yaml:"dbconfig"`
}

func init() {
	gateWayConfigPath := os.Getenv(constant.CONF_GATEWAY_FILE_PATH)
	if gateWayConfigPath == "" {
		gateWayConfigPath = DefaultPath
	}
	if path.Ext(gateWayConfigPath) != ".yml" {
		logger.Errorf("application configure file name{%v} suffix must be .yml", gateWayConfigPath)
		return
	}
	confFileStream, err := ioutil.ReadFile(gateWayConfigPath)
	if err != nil {
		logger.Errorf("ioUtil.ReadFile(file:%s) = error:%v", gateWayConfigPath, perrors.WithStack(err))
		return
	}
	dbConfig := new(DbConfig)
	err = yaml.Unmarshal(confFileStream, dbConfig)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
		return
	}
	conf := dbConfig.Config
	db, err = gorm.Open(conf.Dialect, conf.Config)
	if err != nil {
		logger.Errorf("db config error: %v", err)
		return
	}
	db.DB().SetMaxOpenConns(conf.MaxOpen)
	db.DB().SetMaxIdleConns(conf.MaxIdle)
	db.DB().SetConnMaxLifetime(time.Minute * conf.MaxLifeTime)
	logger.Infof("db[%s] init success", conf.Dialect)
}

func GetDB() *gorm.DB {
	return db
}
