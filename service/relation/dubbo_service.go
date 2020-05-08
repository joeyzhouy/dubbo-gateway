package relation

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/url"
	"strings"
)

var NoRight = errors.New("no right")

func NewRegistryService(db *gorm.DB) service.RegisterService {
	return &registryService{db}
}

type registryService struct {
	*gorm.DB
}

func (d *registryService) GetByRegistryId(registryId int64) (*vo.Registry, error) {
	registry := new(vo.Registry)
	err := d.Where("id = ?", registryId).Find(registry).Error
	if err != nil {
		return nil, err
	}
	references := make([]entry.Reference, 0)
	err = d.Where("registry_id = ?", registryId).Find(&references).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	dbReferenceMap := make(map[string]int)
	for _, reference := range references {
		dbReferenceMap[reference.InterfaceName] = 1
	}
	dis, err := extension.GetDiscover(config.DiscoverConfig{
		Password: registry.Password,
		UserName: registry.UserName,
		Protocol: registry.Protocol,
		Address:  registry.Address,
		Timeout:  registry.Timeout,
	})
	if err != nil {
		return nil, err
	}
	nodes, err := dis.GetChildrenInterface()
	if err != nil {
		return nil, err
	}
	noInterfaces := make([]string, 0)
	for _, node := range nodes {
		if _, ok := dbReferenceMap[node.SubPath]; !ok {
			noInterfaces = append(noInterfaces, node.SubPath)
		}
	}
	registry.NoReferences = noInterfaces
	registry.References = references
	return registry, nil
}

func (d *registryService) GetRegistryByName(name string) ([]entry.Registry, error) {
	result := make([]entry.Registry, 0)
	var err error
	if name == "" {
		err = d.Find(&result).Error
	} else {
		err = d.Where("name LIKE ?", "%"+name+"%").Find(&result).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return result, nil
		} else {
			return nil, err
		}
	}
	return result, nil
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

func (d *registryService) AddRegistryConfig(config entry.Registry) (err error) {
	defer func() {
		if err == nil {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Registry,
				Type:   extension.Add,
				Key:    config.ID,
			})
		}
	}()
	return d.Save(&config).Error
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
	return d.Where("id = ?", registryId).Delete(entry.Registry{}).Error
}

func (d *registryService) ListRegistryByUser(userId int64) ([]entry.Registry, error) {
	result := make([]entry.Registry, 0)
	err := d.Where("user_id = ?", userId).Find(&result).Error
	return result, err
}

type referenceService struct {
	*gorm.DB
	service.MethodService
}

func (r *referenceService) GetReferenceByApiId(apiId int64) ([]entry.Reference, error) {
	var result []entry.Reference
	err := r.Table("d_reference").
		Select("d_reference.id, d_reference.registry_id, d_reference.protocol, d_reference.interface_name, d_reference.cluster").
		Joins("JOIN d_api_chain on d_api_chain.reference_id = d_reference.id and d_api_chain.is_delete = 0").
		Where("d_api_chain.api_id = ?", apiId).Find(&result).Error
	return result, err
}

func (r *referenceService) GetReferenceEntryById(id int64) (*entry.Reference, error) {
	result := new(entry.Reference)
	err := r.Where("id = ?", id).Find(result).Error
	return result, err
}

func (r *referenceService) GetByRegistryIdAndName(registryId int64, name string) ([]entry.Reference, error) {
	result := make([]entry.Reference, 0)
	var err error
	var db *gorm.DB
	if registryId != 0 {
		db = r.Where("registry_id = ?", registryId)
	}
	if name != "" {
		if db == nil {
			db = r.Where("interface_name LIKE ?", "%"+name+"%")
		} else {
			db = db.Where("interface_name LIKE ?", "%"+name+"%")
		}
	}
	if db == nil {
		err = r.Find(&result).Error
	} else {
		err = db.Find(&result).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return result, nil
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (r *referenceService) GetReferenceById(id int64) (*vo.Reference, error) {
	result := new(vo.Reference)
	err := r.Where("id = ?", id).Find(result).Error
	if err != nil {
		return nil, err
	}
	methods := make([]entry.Method, 0)
	if err := r.Where("reference_id = ?", id).Find(&methods).Error; err != nil {
		return nil, err
	}
	if len(methods) > 0 {
		ms := make([]*vo.Method, 0, len(methods))
		for _, method := range methods {
			me, err := r.MethodService.GetMethodDetailByMethod(method)
			if err != nil {
				return nil, err
			}
			ms = append(ms, me)
		}
		result.Methods = ms
	}
	return result, nil
}

func NewReferenceService(db *gorm.DB) service.ReferenceService {
	return &referenceService{db, NewMethodService(db)}
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

func (r *referenceService) AddReference(reference entry.Reference) (err error) {
	defer func() {
		if err == nil {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Reference,
				Type:   extension.Add,
				Key:    reference.ID,
			})
		}
	}()
	return r.Save(&reference).Error
}

func (r *referenceService) DeleteReference(id int64) (err error) {
	var count int
	if err := r.Model(&entry.Method{}).Where("reference_id = ? and is_delete = 0", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("must delete method first")
	}
	defer func() {
		if err == nil {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Reference,
				Type:   extension.Delete,
				Key:    id,
			})
		}
	}()
	return r.Where("id = ?", id).Delete(entry.Reference{}).Error
}

func (r *referenceService) ListAll() ([]entry.Reference, error) {
	result := make([]entry.Reference, 0)
	err := r.Where("is_delete = 0").Find(&result).Error
	return result, err
}

type methodService struct {
	*gorm.DB
	service.EntryService
}

func (m *methodService) SearchMethods(registryId, referenceId int64, methodName string) ([]*vo.Method, error) {
	db := m.Table("d_method").Select("d_method.id, d_method.reference_id, d_method.method_name, d_method.create_time, d_method.modify_time, d_method.is_delete")
	if registryId != 0 {
		db = db.Joins("JOIN d_reference ON d_reference.id = d_method.reference_id").Where("d_reference.registry_id = ?", registryId)
	}
	if referenceId != 0 {
		db = db.Where("d_method.reference_id = ?", referenceId)
	}
	if methodName != "" {
		db = db.Where("d_method.method_name LIKE ?", "%"+methodName+"%")
	}
	var methods []entry.Method
	err := db.Find(&methods).Error
	if err != nil {
		return nil, err
	}
	return m.GetMethodDetailByMethods(methods)
}

func (m *methodService) GetMethodInfoByReferenceId(referenceId int64) (*vo.ReferenceMethodInfo, error) {
	result := new(vo.ReferenceMethodInfo)
	err := m.Where("id = ?", referenceId).Find(&result).Error
	if err != nil {
		return nil, err
	}
	registryConfig := new(entry.Registry)
	err = m.Where("id = ?", result.RegistryId).Find(&registryConfig).Error
	if err != nil {
		return nil, err
	}
	dis, err := extension.GetDiscover(config.DiscoverConfig{
		Timeout:  registryConfig.Timeout,
		Address:  registryConfig.Address,
		Protocol: registryConfig.Protocol,
		Password: registryConfig.Password,
		UserName: registryConfig.UserName,
	})
	if err != nil {
		return nil, err
	}
	nodes, err := dis.GetChildrenMethod(result.InterfaceName)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("no provider interfaceName: " + result.InterfaceName)
	}
	node := nodes[0]
	str, err := url.QueryUnescape(node.SubPath)
	if err != nil {
		return nil, err
	}
	temp, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	interfaceMethods := strings.Split(temp.Query()["methods"][0], ",")
	methods := make([]entry.Method, 0)
	err = m.Where("reference_id = ?", referenceId).Find(&methods).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	if len(methods) > 0 {
		if voMethods, err := m.GetMethodDetailByMethods(methods); err != nil {
			return nil, err
		} else {
			result.Methods = voMethods
		}
	} else {
		result.Methods = make([]*vo.Method, 0)
	}
	noMethods := make([]string, 0)
	methodMap := make(map[string]int)
	for _, method := range methods {
		methodMap[method.MethodName] = 1
	}
	for _, methodName := range interfaceMethods {
		if _, ok := methodMap[methodName]; !ok {
			noMethods = append(noMethods, methodName)
		}
	}
	result.NoMethods = noMethods
	return result, nil
}

func (m *methodService) GetMethodDetailByIds(methodIds []int64) ([]*vo.Method, error) {
	methods := make([]entry.Method, 0)
	err := m.Where("id IN (?)", &methodIds).Find(&methodIds).Error
	if err != nil {
		return nil, err
	}
	return m.GetMethodDetailByMethods(methods)
}

func (m *methodService) GetMethodDetailByMethods(methods []entry.Method) ([]*vo.Method, error) {
	var (
		entryIds  []int64
		err       error
		methodIds []int64
	)
	result := make([]*vo.Method, 0, len(methods))
	for _, method := range methods {
		methodIds = append(methodIds, method.ID)
	}
	var params []entry.MethodParam
	err = m.Where("method_id IN (?)", methodIds).Find(&params).Error
	if err != nil {
		return nil, err
	}
	paramMap := make(map[int64]map[int64]entry.MethodParam)
	entryIds = make([]int64, 0)
	for _, p := range params {
		temp, ok := paramMap[p.MethodId]
		if !ok {
			temp = make(map[int64]entry.MethodParam)
		}
		temp[p.ID] = p
		paramMap[p.MethodId] = temp
		entryIds = append(entryIds, p.EntryId)
	}
	entryMap := make(map[int64]*entry.EntryStructure)
	if len(entryIds) > 0 {
		if es, err := m.EntryService.GetEntries(entryIds); err != nil {
			return nil, err
		} else {
			for _, e := range es {
				entryMap[e.Entry.ID] = e
			}
		}
	}
	for _, method := range methods {
		ms := &vo.Method{Method: method}
		if params, ok := paramMap[method.ID]; ok {
			for _, value := range params {
				if en, ok := entryMap[value.EntryId]; ok {
					mp := &entry.MethodParamStructure{
						MethodParam:    value,
						EntryStructure: *en,
					}
					if err := mp.InitStructure(); err != nil {
						return nil, err
					}
					if value.TypeId == entry.MethodEntryResult {
						ms.Result = mp
					} else if value.TypeId == entry.MethodEntryParam {
						if ms.Params == nil {
							ms.Params = make([]*entry.MethodParamStructure, 0)
						}
						ms.Params = append(ms.Params, mp)
					}
				}
			}
		}
		result = append(result, ms)
	}
	return result, nil
}

func (m *methodService) GetMethodDetailByMethod(me entry.Method) (*vo.Method, error) {
	temp, err := m.GetMethodDetailByMethods([]entry.Method{me})
	if err != nil {
		return nil, err
	}
	return temp[0], nil
}

func (m *methodService) AddMethod(method *vo.Method) (err error) {
	relations := make([]entry.MethodParam, 0)
	if method.Result != nil && method.Result.EntryId != 0 {
		relations = append(relations, entry.MethodParam{
			TypeId:  entry.MethodEntryResult,
			EntryId: method.Result.EntryId,
			Seq:     1,
		})
	}
	if method.Params != nil && len(method.Params) > 0 {
		for index, param := range method.Params {
			relations = append(relations, entry.MethodParam{
				TypeId:  entry.MethodEntryParam,
				EntryId: param.EntryId,
				Seq:     index + 1,
			})
		}
	}
	tx := m.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Method,
				Type:   extension.Add,
				Key:    method.Method.ID,
			})
		}
	}()
	if err = tx.Save(&method.Method).Error; err != nil {
		return
	}
	length := len(relations)
	if length > 0 {
		valueStrings := make([]string, 0, length)
		valueArgs := make([]interface{}, 0, length)
		for _, relation := range relations {
			valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d, %d)", relation.TypeId, method.Method.ID, relation.EntryId, relation.Seq))
		}
		smt := "INSERT IGNORE INTO d_method_param(type_id, method_id, entry_id, seq) VALUES " + strings.Join(valueStrings, ",")
		if err := tx.Exec(smt, valueArgs...).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (m *methodService) GetMethodDetail(methodId int64) (*vo.Method, error) {
	method := new(entry.Method)
	if err := m.Where("id = ?", methodId).Find(method).Error; err != nil {
		return nil, err
	}
	return m.GetMethodDetailByMethod(*method)
}

func (m *methodService) DeleteMethod(methodId int64) (err error) {
	relations := make([]entry.MethodParam, 0)
	err = m.Where("method_id = ?", methodId).Find(&relations).Error
	if err != nil {
		return
	}
	tx := m.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Method,
				Type:   extension.Delete,
				Key:    methodId,
			})
		}
	}()
	if err = tx.Delete(entry.Method{}, "id = ?", methodId).Error; err != nil {
		return
	}
	if err = tx.Delete(entry.MethodParam{}, "method_id = ?", methodId).Error; err != nil {
		return
	}
	tx.Commit()
	if len(relations) > 0 {
		entries := make([]int64, 0)
		for _, relation := range relations {
			entries = append(entries, relation.EntryId)
		}
		go m.DeleteEntriesByIdsIgnoreError(entries)
	}
	return nil
}

func (m *methodService) GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error) {
	result := make([]entry.Method, 0)
	if err := m.Where("reference_id = ?", referenceId).Find(&result).Error; err != nil {
		return result, err
	}
	return result, nil
}

func NewMethodService(db *gorm.DB) service.MethodService {
	return &methodService{db, NewEntryService(db)}
}
