package core

import (
	"strings"
)

//配置文件字符串解析

func StrToList(str string) []int {
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")
	strList := strings.Split(str, ",")
	list := make([]int, 0)
	for _, v := range strList {
		list = append(list, Atoi(v))
	}
	return list
}
func StrToInt64List(str string) []int64 {
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")
	strList := strings.Split(str, ",")
	list := make([]int64, 0)
	for _, v := range strList {
		list = append(list, int64(Atoi(v)))
	}
	return list
}

func StrToInt32List(str string) []int32 {
	str = strings.ReplaceAll(str, "[", "")
	str = strings.ReplaceAll(str, "]", "")
	strList := strings.Split(str, ",")
	list := make([]int32, 0)
	for _, v := range strList {
		list = append(list, int32(Atoi(v)))
	}
	return list
}

func StrTo2dList(str string) [][]int {
	str = strings.ReplaceAll(str, "[[", "")
	str = strings.ReplaceAll(str, "]]", "")
	strList := strings.Split(str, "],[")
	idGroups := make([][]int, 0)
	for _, v := range strList {
		list := strings.Split(v, ",")
		IdList := make([]int, 0)
		for _, strId := range list {
			IdList = append(IdList, Atoi(strId))
		}
		idGroups = append(idGroups, IdList)
	}
	return idGroups
}

func StrTo2dInt64List(str string) [][]int64 {
	str = strings.ReplaceAll(str, "[[", "")
	str = strings.ReplaceAll(str, "]]", "")
	strList := strings.Split(str, "],[")
	idGroups := make([][]int64, 0)
	for _, v := range strList {
		list := strings.Split(v, ",")
		IdList := make([]int64, 0)
		for _, strId := range list {
			IdList = append(IdList, int64(Atoi(strId)))
		}
		idGroups = append(idGroups, IdList)
	}
	return idGroups
}

func StrTo2dInt32List(str string) [][]int32 {
	str = strings.ReplaceAll(str, "[[", "")
	str = strings.ReplaceAll(str, "]]", "")
	strList := strings.Split(str, "],[")
	idGroups := make([][]int32, 0)
	for _, v := range strList {
		list := strings.Split(v, ",")
		IdList := make([]int32, 0)
		for _, strId := range list {
			IdList = append(IdList, int32(Atoi(strId)))
		}
		idGroups = append(idGroups, IdList)
	}
	return idGroups
}

//func StrToItems(str string) []*message.PropItem {
//	Groups := StrTo2dList(str)
//	items := make([]*message.PropItem, 0)
//	for _, group := range Groups {
//		items = append(items,
//			&message.PropItem{
//				Id:       int32(group[0]),
//				Count:    int64(group[1]),
//				SingleId: Sprintf("%d", group[0]),
//			})
//	}
//	return items
//}
//
//func StrToItem(effect string) *message.PropItem {
//	effect = strings.ReplaceAll(effect, "[", "")
//	effect = strings.ReplaceAll(effect, "]", "")
//	stringList := strings.Split(effect, ",")
//	return &message.PropItem{
//		SingleId: stringList[0],
//		Id:       int32(Atoi(stringList[0])),
//		Count:    int64(Atoi(stringList[1])),
//	}
//
//}

func CompareList(list1, list2 []int32) bool {
	for _, Id1 := range list1 {
		var isok bool
		for _, Id2 := range list2 {
			if Id1 == Id2 {
				isok = true
			}
		}
		if !isok {
			return false
		}
	}
	return true
}
