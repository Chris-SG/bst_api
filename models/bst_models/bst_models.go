package bst_models

type BstProfile struct {
	UserId int `json:"userid" gorm:"column:user_id;type:serial;primary_key"`
	User string `json:"user" gorm:"column:user_sub;unique;not_null"'`
	Nickname string `json:"nickname" gorm:"column:nickname"`
	Public bool `json:"public" gorm:"column:public"`
	DdrAutoUpdate bool `json:"ddrautoupdate" gorm:"column:ddr_auto_update"`
	DrsAutoUpdate bool `json:"drsautoupdate" gorm:"column:drs_auto_update"`
}

func (BstProfile) TableName() string {
	return "bstProfile"
}