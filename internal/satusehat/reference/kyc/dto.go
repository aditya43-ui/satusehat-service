package kyc

// GenerateURLRequest merepresentasikan payload untuk membuat link validasi KYC
type GenerateURLRequest struct {
	AgentName string `json:"agent_name" binding:"required"`
	AgentNIK  string `json:"agent_nik" binding:"required,len=16"`
	PublicKey string `json:"public_key"` // Diisi otomatis oleh backend
}

// CallbackRequest merepresentasikan payload webhook/callback dari Satu Sehat setelah KYC selesai.
type CallbackRequest struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}
