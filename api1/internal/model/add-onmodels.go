package model


type Country struct{
	CountryID string `json:"country_id"`
	Probability float64 `json:"probability"`
}
type CountryList struct{
	Countries []Country `json:"country"`
}

type Age struct{
	Age int `json:"age"`
}

type Gender struct{
	Gender string `json:"gender"`
}