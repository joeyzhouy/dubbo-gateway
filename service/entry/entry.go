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

type Entry struct {
	Base
	Name      string `gorm:"column:name" json:"name,omitempty"`
	Key       string `gorm:"column:key" json:"key,omitempty"`
	TypeId    int    `gorm:"column:type_id" json:"typeId,omitempty"`
	ReferIds  string `gorm:"column:refer_ids" json:"referIds,omitempty"`
	Generics  string `gorm:"column:generics" json:"generics,omitempty"`
	Structure string `gorm:"column:structure" json:"structure,omitempty"`
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
	Label          string            `json:"label,omitempty"`
	FieldName      string            `json:"fieldName,omitempty"`
	GenericsValues map[string]string `json:"genericsValues,omitempty"`
	GenericsKey    string            `json:"genericsKey,omitempty"`
	GenericsEntry  map[string]int64  `json:"genericsEntry,omitempty"`
	Entry
}

type Structure []*Field

func (s *Structure) GetTopRefIds() []int64 {
	ids := make([]int64, 0)
	if len(*s) == 0 {
		return ids
	}
	for _, field := range *s {
		if field.Entry.ID != 0 {
			ids = append(ids, field.Entry.ID)
		}
	}
	return ids
}

type EntryStructure struct {
	Entry
	Structure `json:"structure,omitempty"`
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

type GenericsConfig map[string]*Generics

type Generics struct {
	ID int64                `json:"id,omitempty"`
	GS map[string]*Generics `json:"genericsValues,omitempty"`
}

type MethodParamStructure struct {
	MethodParam
	EntryStructure
	GenericsConfig
}

func (m *MethodParamStructure) InitStructure() error {
	str := strings.TrimSpace(m.MethodParam.GenericsValues)
	if str == "" {
		return nil
	}
	gconfig := new(GenericsConfig)
	if err := json.Unmarshal([]byte(str), gconfig); err != nil {
		return err
	}
	m.GenericsConfig = *gconfig
	return nil
}
