package domain

type TranslationResponse struct {
	LexicalItem string `json:"lexicalItem"`
	Meaning     string `json:"meaning"`
	Example     string `json:"example"`
}
