package models

type MongoDebtorModel struct {
	Code  string              `json:"code"`
	Names []LanguageNameModel `json:"names"`
}
