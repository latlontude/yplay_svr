package cache

import (
	"svr/st"
)

var (
	QUESTIONS map[int]*st.QuestionInfo //缓存所有问题列表   10万*50Bytes
	SCHOOLS   map[int]*st.SchoolInfo   //缓存所有学校信息   10万*100Bytes

	PHONE2UIN map[string]int64 //手机号码到UIN的映射 100万*20Bytes

	QICONS      map[int]*QIconInfo //iconid -> url 1万*20Bytes
	QICONSMAXTS int

	ALL_GENE_QIDS []int //通用题目
	ALL_BOY_QIDS  []int //男性题目
	ALL_GIRL_QIDS []int //女性题目
)

func Init() (err error) {

	err = CacheQuestions()
	if err != nil {
		return
	}

	err = CacheSchools()
	if err != nil {
		return
	}

	err = CachePhones()
	if err != nil {
		return
	}

	err = CacheQIcons()
	if err != nil {
		return
	}

	return
}
