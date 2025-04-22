package domain

type Medicine struct {
	ID               int    `json:"id"`
	GlobalID         string `json:"global_id"`
	Name             string `json:"name"`
	Dosage           string `json:"dosage"`
	Form             string `json:"form"`
	ActiveIngredient string `json:"active_ingredient"`
	PharmaCompany    string `json:"pharma_company"`
}

type UpdateMedicine struct {
	GlobalID         *string `json:"global_id"`
	Name             *string `json:"name"`
	Dosage           *string `json:"dosage"`
	Form             *string `json:"form"`
	ActiveIngredient *string `json:"active_ingredient"`
	PharmaCompany    *string `json:"pharma_company"`
}
