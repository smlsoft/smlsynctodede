package models

type MongoUnitModel struct {
	Names    []LanguageNameModel `json:"names"`
	UnitCode string              `json:"unitcode"`
}
