package vo

import (
	"dubbo-gateway/service/entry"
	"encoding/json"
)

type Reference struct {
	entry.Reference
	Methods []Method
}

type Method struct {
	entry.Method
	Params []entry.MethodParam `json:"params"`
}

type MethodDesc struct {
	ID            int64  `json:"id"`
	MethodName    string `json:"methodName"`
	InterfaceName string `json:"interfaceName"`
}

type Entry struct {
	entry.Entry
	entry.Structure
}

func (e *Entry) InitStructure(entryMap map[int64]entry.Entry) error {
	str := e.Entry.Structure
	structure := make(entry.Structure, 0)
	err := json.Unmarshal([]byte(str), &structure)
	if err != nil {
		return err
	}
	count := map[int64]int{e.ID: 1}
	err = structure.Convert(count, entryMap)
	if err != nil {
		return err
	}
	e.Structure = structure
	return nil
}

func (e *Entry) GetStructureInfo() (string, error) {
	if e.Structure == nil {
		return "", nil
	}
	bs, err := json.Marshal(e.Structure)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
