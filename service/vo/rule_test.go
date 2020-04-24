package vo

import (
	"dubbo-gateway/service/entry"
	"fmt"
	"testing"
)

func getTestEntryMap() map[int64]entry.Entry {
	result := make(map[int64]entry.Entry)
	memberEntry := entry.Entry{
		Key: "cn.test.Member",
		Base: entry.Base{
			ID: 101,
		},
		TypeId: entry.ComplexType,
		ReferIds:  "100",
		Name:      "会员结构体",
		Structure: `[{"name":"id","label":"会员ID","fieldType":1,"refId":3},{"name":"memberName","label":"会员名","fieldType":1,"refId":1},{"name":"mobile","label":"手机号","fieldType":1,"refId":1},{"name":"age","label":"年龄","fieldType":1,"refId":2},{"name":"account","label":"会员账号","fieldType":10,"refId":100}]`,
	}
	result[memberEntry.ID] = memberEntry
	accountEntry := entry.Entry{
		Key: "cn.test.MemberAccount",
		TypeId: entry.ComplexType,
		Base: entry.Base{
			ID: 100,
		},
		ReferIds:  "",
		Name:      "会员账号结构体",
		Structure: `[{"name":"memberId","label":"会员ID","fieldType":1,"refId":3},{"name":"points","label":"积分数","fieldType":1,"refId":2},{"name":"member","label":"会员","fieldType":1,"refId":101}]`,
	}
	result[accountEntry.ID] = accountEntry
	result[1] = entry.Entry{
		Base:entry.Base{
			ID: 1,
		},
		TypeId: entry.BaseType,
		Name: "string",
		Key:"java.lang.String",
	}
	result[2] = entry.Entry{
		Base:entry.Base{
			ID: 2,
		},
		TypeId: entry.BaseType,
		Name: "integer",
		Key:"java.lang.Integer",
	}
	result[3] = entry.Entry{
		Base:entry.Base{
			ID: 3,
		},
		TypeId: entry.BaseType,
		Name: "long",
		Key:"java.lang.Long",
	}
	return result
}

func TestRule(t *testing.T) {
	entryMap := getTestEntryMap()
	memberEntry, _ := entryMap[101]
	en := Entry{Entry: memberEntry}
	err := en.InitStructure(entryMap)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = en.Structure.Convert(map[int64]int{}, entryMap)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(en)
}
