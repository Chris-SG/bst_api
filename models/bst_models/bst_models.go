package bst_models

type BstProfile struct {
	User string `json:"user" gorm:"column:user;primary_key"'`
	UserId int `json:"userid" gorm:"column:user_id; type:serial"`
	Nickname string `json:"nickname" gorm:"column:nickname"`
	Public bool `json:"public" gorm:"column:public"`
}

func (BstProfile) TableName() string {
	return "bstProfile"
}