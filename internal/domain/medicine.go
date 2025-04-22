package domain

type Medicine struct {
	ID               int    `json:"id"`
	NDC              string `json:"ndc"`
	Name             string `json:"name"`
	Dosage           string `json:"dosage"`
	Form             string `json:"form"`
	ActiveIngredient string `json:"active_ingredient"`
	PharmaCompany    string `json:"pharma_company"`
}

type UpdateMedicine struct {
	NDC              *string `json:"ndc"`
	Name             *string `json:"name"`
	Dosage           *string `json:"dosage"`
	Form             *string `json:"form"`
	ActiveIngredient *string `json:"active_ingredient"`
	PharmaCompany    *string `json:"pharma_company"`
}
