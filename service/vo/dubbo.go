package vo

import (
	"dubbo-gateway/service/entry"
)

type Registry struct {
	entry.Registry
	References   []entry.Reference `json:"references"`
	NoReferences []string          `json:"noReferences"`
}

type Reference struct {
	entry.Reference
	Methods []*Method `json:"methods"`
}

type ReferenceMethodInfo struct {
	Reference
	NoMethods []string `json:"noMethods"`
}

type Method struct {
	entry.Method
	Result *entry.MethodParamStructure   `json:"result"`
	Params []*entry.MethodParamStructure `json:"params"`
}

type MethodDeclaration struct {
	entry.Method
	Params []entry.Entry `json:"params"`
}

type ParamMethodInfo struct {
	MethodId    int64
	MethodName  string
	Seq         int
	EntryId     int64
	EntryTypeId int
	ParamClass  string
	ParamTypeId int
}

//type Entry struct {
//	entry.Entry
//	entry.Structure
//}
//
//func (e *Entry) InitStructure(entryMap map[int64]entry.Entry) error {
//	str := e.Entry.Structure
//	if strings.TrimSpace(str)  == "" {
//		return nil
//	}
//	structure := make(entry.Structure, 0)
//	err := json.Unmarshal([]byte(str), &structure)
//	if err != nil {
//		return err
//	}
//	count := map[int64]int{e.ID: 1}
//	err = structure.Convert(count, entryMap)
//	if err != nil {
//		return err
//	}
//	e.Structure = structure
//	return nil
//}
//
//func (e *Entry) GetStructureInfo() (string, error) {
//	if e.Structure == nil {
//		return "", nil
//	}
//	bs, err := json.Marshal(e.Structure)
//	if err != nil {
//		return "", err
//	}
//	return string(bs), nil
//}
