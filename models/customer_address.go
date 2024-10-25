package models

type CustomerAddressModel struct {
	Contactnames []LanguageNameModel `json:"contactnames"` // ชื่อลูกหนี้
	Address      []string            `json:"address"`      // ที่อยู่
	Phoneprimary string              `json:"phoneprimary"` // เบอร์โทรศัพท์หลัก

}
