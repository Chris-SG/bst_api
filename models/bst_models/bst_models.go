package bst_models

type BstProfile struct {
	UserId int `json:"userid" gorm:"column:user_id;type:serial;primary_key"`
	User string `json:"user" gorm:"column:user;unique;not_null"'`
	Nickname string `json:"nickname" gorm:"column:nickname"`
	Public bool `json:"public" gorm:"column:public"`
}

func (BstProfile) TableName() string {
	return "bstProfile"
}