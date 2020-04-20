package relation

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"errors"
	"github.com/jinzhu/gorm"
)

var NoRight = errors.New("no right")

func NewRegistryService(db *gorm.DB) service.RegisterService {
	return &registryService{db}
}

type registryService struct {
	*gorm.DB
}

func (d *registryService) ListAll() ([]entry.Registry, error) {
	result := make([]entry.Registry, 0)
	err := d.Where("is_delete = 0").Find(&result).Error
	return result, err
}

func (d *registryService) RegisterDetail(userId, registerId int64) (*entry.Registry, error) {
	reg := new(entry.Registry)
	err := d.Where("id = ?", registerId).Find(&reg).Error
	if err != nil {
		return nil, err
	}
	if userId != reg.UserId {
		return nil, NoRight
	}
	return reg, nil
}

func (d *registryService) AddRegistryConfig(config entry.Registry) error {
	return d.Save(config).Error
}

func (d *registryService) DeleteRegistryConfig(registryId, userId int64) error {
	dbRegistry := new(entry.Registry)
	err := d.Where("user_id = ? and id = ?", userId, registryId).
		Find(dbRegistry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return NoRight
		}
		return err
	}
	return d.Where("id = ?", registryId).UpdateColumn("is_delete", 1).Error
}

func (d *registryService) ListRegistryByUser(userId int64) ([]entry.Registry, error) {
	result := make([]entry.Registry, 0)
	err := d.Where("user_id = ?", userId).Find(&result).Error
	return result, err
}

type referenceService struct {
	*gorm.DB
}

func NewReferenceService(db *gorm.DB) service.ReferenceService {
	return &referenceService{db}
}

func (r *referenceService) GetByIds(ids []int64) ([]entry.Reference, error) {
	result := make([]entry.Reference, 0)
	if len(ids) == 0 {
		return result, nil
	}
	err := r.Where("id IN (?)", &ids).Find(&result).Error
	return result, err
}

func (r *referenceService) ListByUser(userId int64) ([]entry.Reference, error) {
	result := make([]entry.Reference, 0)
	err := r.Where("user_id = ? and is_delete = 0", userId).Find(&result).Error
	return result, err
}

func (r *referenceService) AddReference(reference entry.Reference) error {
	return r.Save(&reference).Error
}

func (r *referenceService) DeleteReference(id int64) error {
	var count int
	if err := r.Model(&entry.Method{}).Where("reference_id = ? and is_delete = 0", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("must delete method first")
	}
	return r.Model(&entry.Reference{}).Where("reference_id = ?", id).UpdateColumn("is_delete", 1).Error
}

func (r *referenceService) ListAll() ([]entry.Reference, error) {
	result := make([]entry.Reference, 0)
	err := r.Where("is_delete = 0").Find(&result).Error
	return result, err
}

type methodService struct {
	*gorm.DB
}

func (m *methodService) ListByUserIdAndMethodName(userId int64, methodName string) ([]vo.MethodDesc, error) {
	result := make([]vo.MethodDesc, 0)
	if err := m.Table("d_method").Select("d_method.method_name as method_name, d_method.id as id, d_reference.interface_name as interface_name").
		Joins("JOIN d_reference on d_reference.id = d_method.reference_id").
		Where("d_method.user_id = ? and d_method.method_name LIKE ?", userId, methodName+"%").Scan(&result).Error;
		err != nil {
		return result, err
	}
	return result, nil
}

func (m *methodService) AddMethod(method *vo.Method) error {
	tx := m.Begin()
	if err := tx.Save(&method.Method).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, mm := range method.Params {
		mm.MethodId = method.Method.ID
		if err := tx.Save(&mm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (m *methodService) GetMethodDetail(methodId int64) (*vo.Method, error) {
	method := new(vo.Method)
	if err := m.Where("id = ?", methodId).Find(&method.Method).Error; err != nil {
		return nil, err
	}
	if err := m.Where("method_id = ?", methodId).Find(&method.Params).Error; err != nil {
		return nil, err
	}
	return method, nil
}

func (m *methodService) DeleteMethod(methodId int64) error {
	tx := m.Begin()
	if err := tx.Model(&entry.Method{}).Where("id = ?", methodId).UpdateColumn("is_delete", 1).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&entry.MethodParam{}).Where("method_id = ?", methodId).UpdateColumn("is_delete", 1).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (m *methodService) GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error) {
	result := make([]entry.Method, 0)
	if err := m.Where("registry_id = ?", referenceId).Find(&result).Error; err != nil {
		return result, err
	}
	return result, nil
}

func NewMethodService(db *gorm.DB) service.MethodService {
	return &methodService{db}
}
