package models

type MongoCreditorModel struct {
	Code  string              `json:"code"`
	Names []LanguageNameModel `json:"names"`
}
