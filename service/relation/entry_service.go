package relation

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"errors"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
)

func NewEntryService(db *gorm.DB) service.EntryService {
	return &entryService{db}
}

type entryService struct {
	*gorm.DB
}

func (e *entryService) SearchEntries(name string, pageSize int) ([]*vo.Entry, error) {
	var ids []int64
	db := e.Table("d_entry")
	if name != "" {
		db = db.Where("`key` LIKE ?", "%"+name+"%")
	}
	err := db.Order("id").Limit(pageSize).Pluck("id", &ids).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	if len(ids) == 0 {
		return make([]*vo.Entry, 0), nil
	}
	return e.GetEntries(ids)
}

func (e *entryService) DeleteEntriesByIdsIgnoreError(ids []int64) {
	if len(ids) == 0 {
		return
	}
	for _, entryId := range ids {
		if err := e.DeleteEntry(entryId); err != nil {
			logger.Errorf("delete entry[%d] error: %v", entryId, err)
		}
	}
}

func (e *entryService) GetByType(typeId int) ([]entry.Entry, error) {
	result := make([]entry.Entry, 0)
	err := e.Where("type_id = ?", typeId).Find(&result).Error
	return result, err
}

func (e *entryService) SaveEntry(es *entry.EntryStructure) error {
	var err error
	en := es.Entry
	en.TypeId = entry.ComplexType
	en.Structure, err = es.GetStructureInfo()
	if err != nil {
		return err
	}
	referIds := es.GetTopRefIds()
	length := len(referIds)
	if length > 0 {
		for index, id := range referIds {
			en.ReferIds += strconv.FormatInt(id, 10)
			if index != length-1 {
				en.ReferIds = ","
			}
		}
	}
	tx := e.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	err = tx.Save(es).Error
	if err != nil {
		return err
	}
	if length > 0 {
		if err = e.batchRelation(en.ID, referIds, tx) err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}


func (e *entryService) UpdateEntry(enVo *vo.Entry) error {
	var err error
	en := enVo.Entry
	en.TypeId = entry.ComplexType
	en.Structure, err = enVo.GetStructureInfo()
	if err != nil {
		return err
	}
	referIds := enVo.GetTopRefIds()
	referIdsLen := len(referIds)
	if referIdsLen > 0 {
		for index, id := range referIds {
			en.ReferIds = strconv.FormatInt(id, 10)
			if index != referIdsLen-1 {
				en.ReferIds = ","
			}
		}
	}
	tx := e.Begin()
	err = tx.Model(&entry.Entry{}).Where("id = ?", en.ID).Updates(map[string]interface{}{
		"name":      en.Name,
		"key":       en.Key,
		"refer_ids": en.ReferIds,
		"structure": en.Structure,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Where("entry_id = ?", en.ID).Delete(entry.EntryRelation{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	if referIdsLen > 0 {
		if err = e.batchRelation(en.ID, referIds, tx); err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (e *entryService) batchRelation(entryId int64, ids []int64, tx *gorm.DB) error {
	length := len(ids)
	valueStrings := make([]string, 0, length)
	valueArgs := make([]interface{}, 0, length)
	for _, id := range ids {
		valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d)", entryId, id))
	}
	smt := "INSERT IGNORE INTO d_entry_relation(entry_id, refer_id) VALUES " + strings.Join(valueStrings, ",")
	if err := tx.Exec(smt, valueArgs...).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

var EntryAlreadyUsedError = errors.New("entry already used")

func (e *entryService) DeleteEntry(id int64) error {
	relation := new(entry.EntryRelation)
	err := e.Where("refer_id = ?", id).Find(relation).Error
	if err == nil && relation.ID > 0 {
		return EntryAlreadyUsedError
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	tx := e.Begin()
	err = tx.Delete(entry.EntryRelation{}, "entry_id = ?", id).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Delete(entry.Entry{}, "id = ?", id).Error
	if err != nil {
		tx.Rollback()
	}
	return err
}

func (e *entryService) GetEntry(id int64) (*vo.Entry, error) {
	entries, err := e.GetEntries([]int64{id})
	if err != nil {
		return nil, err
	}
	return entries[0], nil
}

func (e *entryService) GetEntries(ids []int64) ([]*vo.Entry, error) {
	result := make([]*vo.Entry, 0)
	entryMap := make(map[int64]entry.Entry)
	if err := e.GetAllReferEntryMap(ids, entryMap); err != nil {
		return nil, err
	}
	if bases, err := e.GetByType(entry.BaseType); err != nil {
		return nil, err
	} else {
		for _, base := range bases {
			entryMap[base.ID] = base
		}
	}
	for _, id := range ids {
		if en, ok := entryMap[id]; !ok {
			return nil, errors.New(fmt.Sprintf("miss entry with id: %d", id))
		} else {
			voEn := &vo.Entry{Entry: en}
			if err := voEn.InitStructure(entryMap); err != nil {
				return nil, err
			}
			result = append(result, voEn)
		}
	}
	return result, nil
}

func (e *entryService) GetAllReferEntryMap(ids []int64, result map[int64]entry.Entry) error {
	temp := make([]entry.Entry, 0)
	err := e.Where("id IN (?)", ids).Find(&temp).Error
	if err != nil {
		return err
	}
	ids = make([]int64, 0)
	for _, item := range temp {
		result[item.ID] = item
		if item.ReferIds == "" {
			continue
		}
		for _, str := range strings.Split(item.ReferIds, ",") {
			if id, err := strconv.ParseInt(str, 10, 64); err != nil {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return e.GetAllReferEntryMap(ids, result)
}
