package models

type MongoDebtorModel struct {
	Code              string               `json:"code"`              // รหัสลูกหนี้
	Names             []LanguageNameModel  `json:"names"`             // ชื่อลูกหนี้
	PersonalType      int                  `json:"personaltype"`      // 1 = บุคคลธรรมดา, 2 = นิติบุคคล
	TaxId             string               `json:"taxid"`             // เลขประจำตัวผู้เสียภาษี
	CustomerType      int                  `json:"customertype"`      // 1 = สำนักงานใหญ่ 2 = สาขา
	BranchNumber      string               `json:"branchnumber"`      // รหัสสาขา
	Email             string               `json:"email"`             // อีเมลล์
	CreditDay         int                  `json:"creditday"`         // จำนวนวันเครดิต
	IsMember          bool                 `json:"ismember"`          // false = ไม่ใช่สมาชิก, true = เป็นสมาชิก
	AddressForBilling CustomerAddressModel `json:"addressforbilling"` // ที่อยู่ในการออกใบกำกับ
	Groups            []string             `json:"groups"`            // กลุ่มลูกหนี้
	Images            []string             `json:"images"`            // รูปภาพ
}
