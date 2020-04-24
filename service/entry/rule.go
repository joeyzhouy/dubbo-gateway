package entry

import (
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
)

const (
	FieldTypeSingle   = 1
	FieldTypeMultiple = 10
)

type Structure []*Field

func (s *Structure) GetTopRefIds() []int64 {
	ids := make([]int64, 0)
	if len(*s) == 0 {
		return ids
	}
	for _, field := range *s {
		//TODO default rule
		if field.RefId >= 100 {
			ids = append(ids, field.RefId)
		}
	}
	return ids
}

func (s *Structure) Convert(count map[int64]int, entryMap map[int64]Entry) error {
	for _, field := range *s {
		if field.RefId == 0 {
			continue
		}
		entry, ok := entryMap[field.RefId]
		if !ok {
			return errors.New(fmt.Sprintf("not mapping entry with entryId: %d", field.RefId))
		}
		field.Identity = entry.Key
		if entry.TypeId == BaseType {
			continue
		}
		num, ok := count[field.RefId]
		if !ok {
			num = 0
		}
		// max depth is 2
		if num > 1 {
			continue
		}
		num++
		count[field.RefId] = num
		fmt.Printf("entryId: %d, num: %d", field.RefId, num)
		subStructure := make(Structure, 0)
		fmt.Println(entry.Structure)
		err := json.Unmarshal([]byte(entry.Structure), &subStructure)
		if err != nil {
			return err
		}
		if err = subStructure.Convert(count, entryMap); err != nil {
			return err
		}
		field.Structure = &subStructure
	}
	return nil
}

type Field struct {
	Name      string     `json:"name"`
	Label     string     `json:"label"`
	RefId     int64      `json:"refId"`
	FieldType int        `json:"fieldType"`
	Identity  string     `json:"identity"`
	Structure *Structure `json:"structure"`
}
