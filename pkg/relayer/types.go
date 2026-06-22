package relayer

// Request types.

// SubmitRequest represents the request payload for submitting a gasless transaction.
// The Transaction field is a generic map because its structure varies depending
// on the type of transaction being submitted.
type SubmitRequest struct {
	Transaction map[string]interface{} `json:"transaction"`
	Signature   string                 `json:"signature"`
}

// GetTransactionsRequest holds pagination parameters for listing transactions.
type GetTransactionsRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// Response types.

// SubmitResponse is returned after successfully submitting a gasless transaction.
type SubmitResponse struct {
	TransactionID string `json:"transactionID"`
	State         string `json:"state"`
}

// Transaction represents a relayer transaction with its current state and metadata.
type Transaction struct {
	ID            string `json:"id,omitempty"`
	TransactionID string `json:"transactionID,omitempty"`
	State         string `json:"state,omitempty"`
	From          string `json:"from,omitempty"`
	To            string `json:"to,omitempty"`
	Data          string `json:"data,omitempty"`
	Value         string `json:"value,omitempty"`
	Hash          string `json:"hash,omitempty"`
	BlockNumber   int    `json:"blockNumber,omitempty"`
	GasLimit      int    `json:"gasLimit,omitempty"`
	GasPrice      string `json:"gasPrice,omitempty"`
	Nonce         int    `json:"nonce,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// NonceResponse contains the current nonce for a signer.
type NonceResponse struct {
	Nonce string `json:"nonce"`
}

// RelayPayloadResponse contains the relayer address and nonce for a signer.
type RelayPayloadResponse struct {
	RelayerAddress string `json:"relayerAddress"`
	Nonce          string `json:"nonce"`
}

// DeployedResponse indicates whether a wallet is deployed at a given address.
type DeployedResponse struct {
	Deployed bool `json:"deployed"`
}

// APIKeyResponse represents a relayer API key associated with the authenticated user.
type APIKeyResponse struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}
