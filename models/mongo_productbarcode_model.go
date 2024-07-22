package models

type MongoProductBarcodeModel struct {
	Barcode       string              `json:"barcode"`
	ItemUnitCode  string              `json:"itemunitcode"`
	ItemUnitNames []LanguageNameModel `json:"itemunitnames"`
	ItemType      int                 `json:"itemtype"`
	FoodType      int                 `json:"foodtype"`
	TaxType       int                 `json:"taxtype"`
	IsSumPoint    bool                `json:"issumpoint"`
	ItemCode      string              `json:"itemcode"`
	Names         []LanguageNameModel `json:"names"`
	GroupCode     string              `json:"groupcode"`
	GroupNames    []LanguageNameModel `json:"groupnames"`
	Prices        []PriceModel        `json:"prices"`
}

type LanguageNameModel struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Isauto   bool   `json:"isauto"`
	Isdelete bool   `json:"isdelete"`
}

type PriceModel struct {
	KeyNumber int     `json:"keynumber"`
	Price     float64 `json:"price"`
}
