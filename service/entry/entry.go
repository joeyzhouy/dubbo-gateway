package entry

import (
	"encoding/json"
	"strconv"
	"strings"
)

const (
	FieldTypeSingle   = 1
	FieldTypeMultiple = 10
)

//
//func (s *Structure) Convert(count map[int64]int, entryMap map[int64]Entry) error {
//	for _, field := range *s {
//		if field.RefId == 0 {
//			continue
//		}
//		entry, ok := entryMap[field.RefId]
//		if !ok {
//			return errors.New(fmt.Sprintf("not mapping entry with entryId: %d", field.RefId))
//		}
//		field.Identity = entry.Key
//		if entry.TypeId == BaseType {
//			continue
//		}
//		num, ok := count[field.RefId]
//		if !ok {
//			num = 0
//		}
//		// max depth is 2
//		if num > 1 {
//			continue
//		}
//		num++
//		count[field.RefId] = num
//		fmt.Printf("entryId: %d, num: %d", field.RefId, num)
//		subStructure := make(Structure, 0)
//		fmt.Println(entry.Structure)
//		err := json.Unmarshal([]byte(entry.Structure), &subStructure)
//		if err != nil {
//			return err
//		}
//		if err = subStructure.Convert(count, entryMap); err != nil {
//			return err
//		}
//		field.Structure = &subStructure
//	}
//	return nil
//}

type Entry struct {
	Base
	Name      string `gorm:"column:name" json:"name"`
	Key       string `gorm:"column:key" json:"key"`
	TypeId    int    `gorm:"column:type_id" json:"typeId"`
	ReferIds  string `gorm:"column:refer_ids" json:"refer_ids"`
	Generics  string `gorm:"column:generics" json:"generics"`
	Structure string `gorm:"column:structure" json:"structure"`
}

func (Entry) TableName() string {
	return "d_entry"
}

type EntryRelation struct {
	Base
	EntryId int64 `gorm:"column:entry_id" json:"entryId"`
	ReferId int64 `gorm:"column:refer_id" json:"referId"`
}

func (EntryRelation) TableName() string {
	return "d_entry_relation"
}

type Field struct {
	Label          string `json:"label"`
	GenericsValues map[string]string
	GenericsKey    string `json:"genericsKey"`
	GenericsEntry  map[string]int64
	Entry
}

type Structure []*Field

func (s *Structure) GetTopRefIds() []int64 {
	ids := make([]int64, 0)
	if len(*s) == 0 {
		return ids
	}
	for _, field := range *s {
		ids = append(ids, field.Entry.ID)
	}
	return ids
}

type EntryStructure struct {
	Entry
	Structure `json:"structure"`
}

func (e *EntryStructure) GetStructureInfo() (string, error) {
	if e.Structure == nil {
		return "", nil
	}
	bs, err := json.Marshal(e.Structure)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (e *EntryStructure) InitStructure() error {
	if e.TypeId != ComplexType {
		return nil
	}
	subStructure := new(Structure)
	err := json.Unmarshal([]byte(e.Entry.Structure), subStructure)
	if err != nil {
		return err
	}
	for _, field := range *subStructure {
		if len(field.GenericsValues) == 0 {
			continue
		}
		ges := make(map[string]int64)
		for key, value := range field.GenericsValues {
			id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err == nil {
				ges[key] = id
			}
		}
		if len(ges) > 0 {
			for key, _ := range ges {
				delete(field.GenericsValues, key)
			}
		}
		field.GenericsEntry = ges
	}
	e.Structure = *subStructure
	return nil
}

type MethodParam struct {
	Base
	TypeId         int    `gorm:"column:type_id" json:"typeId"`
	GenericsValues string `gorm:"column:generics_values" json:"genericsValues"`
	MethodId       int64  `gorm:"column:method_id" json:"methodId"`
	EntryId        int64  `gorm:"column:entry_id" json:"entryId"`
}

type GenericsConfig map[string]*Generics

type Generics struct {
	ID int64
	GS map[string]*Generics `json:"g"`
}

type MethodParamStructure struct {
	MethodParam
	EntryStructure
	GenericsConfig
}

func NewMethodParamStructure(param MethodParam, entryMap map[int64]Entry) (*MethodParamStructure, error) {
	result := new(MethodParamStructure)
	//result.MethodParam = param
	//entry := entryMap[param.EntryId]
	//s, err := NewEntryStructure(entry)
	//if err != nil {
	//	return nil, err
	//}
	//result.EntryStructure = *s
	//str := strings.TrimSpace(param.GenericsValues)
	//if str == "" {
	//	return result, nil
	//}
	//gconfig := new(GenericsConfig)
	//err = json.Unmarshal([]byte(str), gconfig)
	//if err != nil {
	//	return nil, err
	//}
	//result.GenericsConfig = *gconfig
	return result, err
}

//func NewEntryStructure(entry Entry) (*EntryStructure, error) {
//	es := new(EntryStructure)
//	es.Entry = entry
//	// BaseType
//	if es.TypeId != 10 {
//		return es, nil
//	}
//	subStructure := new(Structure)
//	err := json.Unmarshal([]byte(entry.Structure), subStructure)
//	if err != nil {
//		return nil, err
//	}
//	for _, field := range *subStructure {
//		if len(field.GenericsValues) == 0 {
//			continue
//		}
//		ges := make(map[string]int64)
//		for key, value := range field.GenericsValues {
//			id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
//			if err == nil {
//				ges[key] = id
//			}
//		}
//		if len(ges) > 0 {
//			for key, _ := range ges {
//				delete(field.GenericsValues, key)
//			}
//		}
//		field.GenericsEntry = ges
//	}
//	es.Structure = *subStructure
//	return es, nil
//}
