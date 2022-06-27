package model

type SettingItem struct {
	Key    string `json:"key" gorm:"primaryKey" binding:"required"` // unique key
	Value  string `json:"value"`                                    // value
	Help   string `json:"help"`                                     // help message
	Type   string `json:"type"`                                     // string, number, bool, select
	Values string `json:"values"`                                   // values for select
	Group  int    `json:"group"`                                    // use to group setting in frontend
	Access int    `json:"access"`                                   // admin/guest/general
}
