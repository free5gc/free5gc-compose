package main

type AkmaKeyInfo struct {
	SuppFeat string `json:"suppFeat,omitempty"`
	Supi     string `json:"supi,omitempty"`
	Gpsi     string `json:"gpsi,omitempty"`
	AKId     string `json:"aKId"`
	KAkma    string `json:"kAkma"`
}

type CtxRemove struct {
	Supi string `json:"supi"`
}

type AkmaAfKeyRequest struct {
	// Add fields according to TS29522_AKMA.yaml#/components/schemas/AkmaAfKeyRequest
}

type AkmaAfKeyData struct {
	// Add fields according to TS29522_AKMA.yaml#/components/schemas/AkmaAfKeyData
}
