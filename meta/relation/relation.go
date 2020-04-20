package relation

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gopkg.in/yaml.v2"
	"time"
)

type DbConfig struct {
	//Dialect     string        `yaml:"db_dialect"`
	Config      string        `yaml:"db_config"`
	MaxOpen     int           `yaml:"db_maxopen"`
	MaxIdle     int           `yaml:"db_maxidle"`
	MaxLifeTime time.Duration `yaml:"db_maxlifetime"`
}

func InitRelationConfig(dialect, configString string) (*gorm.DB, error) {
	dbConfig := new(DbConfig)
	err := yaml.Unmarshal([]byte(configString), dbConfig)
	if err != nil {
		return nil, err
	}
	var db *gorm.DB
	db, err = gorm.Open(dialect, dbConfig.Config)
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxOpenConns(dbConfig.MaxOpen)
	db.DB().SetMaxIdleConns(dbConfig.MaxIdle)
	db.DB().SetConnMaxLifetime(time.Minute * dbConfig.MaxLifeTime)
	return db, nil
}

