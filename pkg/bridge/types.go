package bridge

// Request types.
type (
	DepositRequest struct {
		Address string `json:"address"`
	}
	StatusRequest struct {
		Address string `json:"address"`
	}
	BridgeQuoteRequest struct {
		FromChainID   string `json:"fromChainId"`
		ToChainID     string `json:"toChainId"`
		FromTokenAddr string `json:"fromTokenAddress"`
		ToTokenAddr   string `json:"toTokenAddress"`
		FromAmountRaw string `json:"fromAmountRaw"`
		FromAddress   string `json:"fromAddress"`
	}
	WithdrawAddressRequest struct {
		FromChainID   string `json:"fromChainId"`
		ToChainID     string `json:"toChainId"`
		FromTokenAddr string `json:"fromTokenAddress"`
		ToTokenAddr   string `json:"toTokenAddress"`
		AmountRaw     string `json:"amountRaw"`
		Recipient     string `json:"recipient"`
	}
)

// Response types.
type (
	DepositResponse struct {
		Address DepositAddresses `json:"address"`
		Note    string           `json:"note,omitempty"`
	}
	DepositAddresses struct {
		EVM string `json:"evm"`
		SVM string `json:"svm"`
		BTC string `json:"btc"`
	}
	SupportedAssetsResponse struct {
		SupportedAssets []SupportedAsset `json:"supportedAssets"`
		Note            string           `json:"note,omitempty"`
	}
	SupportedAsset struct {
		ChainID        string `json:"chainId"`
		ChainName      string `json:"chainName"`
		Token          Token  `json:"token"`
		MinCheckoutUSD string `json:"minCheckoutUsd"`
	}
	Token struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
	}
	StatusResponse struct {
		Transactions []DepositTransaction `json:"transactions"`
	}
	BridgeQuoteResponse struct {
		EstimatedFees struct {
			FromAmount string `json:"fromAmount"`
			ToAmount   string `json:"toAmount"`
		} `json:"estimatedFees"`
		QuoteID string `json:"quoteId"`
	}
	WithdrawAddressResponse struct {
		Address DepositAddresses `json:"address"`
		Note    string           `json:"note,omitempty"`
	}
	DepositTransaction struct {
		FromChainID        string `json:"fromChainId"`
		FromTokenAddress   string `json:"fromTokenAddress"`
		FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
		ToChainID          string `json:"toChainId"`
		ToTokenAddress     string `json:"toTokenAddress"`
		Status             string `json:"status"`
		TxHash             string `json:"txHash,omitempty"`
		CreatedTimeMS      *int64 `json:"createdTimeMs,omitempty"`
	}
)
