package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/GoPolymarket/polymarket-go-sdk/v2/pkg/ctf"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	rpcURL := strings.TrimSpace(os.Getenv("CTF_RPC_URL"))
	privKey := strings.TrimSpace(os.Getenv("CTF_PRIVATE_KEY"))
	if rpcURL == "" || privKey == "" {
		fmt.Println("Set CTF_RPC_URL and CTF_PRIVATE_KEY to use CTF transactions.")
		return
	}

	chainID := int64(ctf.PolygonChainID)
	if raw := strings.TrimSpace(os.Getenv("CTF_CHAIN_ID")); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			chainID = parsed
		}
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("RPC dial error: %v", err)
	}

	key, err := crypto.HexToECDSA(strings.TrimPrefix(privKey, "0x"))
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(chainID))
	if err != nil {
		log.Fatalf("Transactor error: %v", err)
	}

	useNegRisk := strings.TrimSpace(os.Getenv("CTF_NEG_RISK")) == "1"
	var ctfClient ctf.Client
	if useNegRisk {
		ctfClient, err = ctf.NewClientWithNegRisk(client, txOpts, chainID)
	} else {
		ctfClient, err = ctf.NewClientWithBackend(client, txOpts, chainID)
	}
	if err != nil {
		log.Fatalf("CTF client init error: %v", err)
	}
	fmt.Println("CTF client ready.")

	if strings.TrimSpace(os.Getenv("CTF_DO_TX")) != "1" {
		fmt.Println("Set CTF_DO_TX=1 to send a transaction.")
		return
	}

	action := strings.ToLower(strings.TrimSpace(os.Getenv("CTF_ACTION")))
	if action == "" {
		fmt.Println("Set CTF_ACTION=split|merge|redeem|redeem_neg_risk")
		return
	}

	switch action {
	case "split":
		req, err := buildSplitRequest()
		if err != nil {
			log.Fatalf("split request error: %v", err)
		}
		resp, err := ctfClient.SplitPosition(context.Background(), req)
		if err != nil {
			log.Fatalf("split error: %v", err)
		}
		fmt.Printf("Split tx: %s (block %d)\n", resp.TransactionHash.Hex(), resp.BlockNumber)
	case "merge":
		req, err := buildMergeRequest()
		if err != nil {
			log.Fatalf("merge request error: %v", err)
		}
		resp, err := ctfClient.MergePositions(context.Background(), req)
		if err != nil {
			log.Fatalf("merge error: %v", err)
		}
		fmt.Printf("Merge tx: %s (block %d)\n", resp.TransactionHash.Hex(), resp.BlockNumber)
	case "redeem":
		req, err := buildRedeemRequest()
		if err != nil {
			log.Fatalf("redeem request error: %v", err)
		}
		resp, err := ctfClient.RedeemPositions(context.Background(), req)
		if err != nil {
			log.Fatalf("redeem error: %v", err)
		}
		fmt.Printf("Redeem tx: %s (block %d)\n", resp.TransactionHash.Hex(), resp.BlockNumber)
	case "redeem_neg_risk":
		req, err := buildRedeemNegRiskRequest()
		if err != nil {
			log.Fatalf("redeem neg risk request error: %v", err)
		}
		resp, err := ctfClient.RedeemNegRisk(context.Background(), req)
		if err != nil {
			log.Fatalf("redeem neg risk error: %v", err)
		}
		fmt.Printf("Redeem neg risk tx: %s (block %d)\n", resp.TransactionHash.Hex(), resp.BlockNumber)
	default:
		log.Fatalf("unknown CTF_ACTION: %s", action)
	}
}

func buildSplitRequest() (*ctf.SplitPositionRequest, error) {
	collateral, err := parseAddressEnv("CTF_COLLATERAL")
	if err != nil {
		return nil, err
	}
	conditionID, err := parseHashEnv("CTF_CONDITION_ID")
	if err != nil {
		return nil, err
	}
	partition, err := parseBigIntListEnv("CTF_PARTITION")
	if err != nil {
		return nil, err
	}
	amount, err := parseBigIntEnv("CTF_AMOUNT")
	if err != nil {
		return nil, err
	}
	parent := parseHashEnvOptional("CTF_PARENT_COLLECTION_ID")

	return &ctf.SplitPositionRequest{
		CollateralToken:    collateral,
		ParentCollectionID: parent,
		ConditionID:        conditionID,
		Partition:          partition,
		Amount:             amount,
	}, nil
}

func buildMergeRequest() (*ctf.MergePositionsRequest, error) {
	collateral, err := parseAddressEnv("CTF_COLLATERAL")
	if err != nil {
		return nil, err
	}
	conditionID, err := parseHashEnv("CTF_CONDITION_ID")
	if err != nil {
		return nil, err
	}
	partition, err := parseBigIntListEnv("CTF_PARTITION")
	if err != nil {
		return nil, err
	}
	amount, err := parseBigIntEnv("CTF_AMOUNT")
	if err != nil {
		return nil, err
	}
	parent := parseHashEnvOptional("CTF_PARENT_COLLECTION_ID")

	return &ctf.MergePositionsRequest{
		CollateralToken:    collateral,
		ParentCollectionID: parent,
		ConditionID:        conditionID,
		Partition:          partition,
		Amount:             amount,
	}, nil
}

func buildRedeemRequest() (*ctf.RedeemPositionsRequest, error) {
	collateral, err := parseAddressEnv("CTF_COLLATERAL")
	if err != nil {
		return nil, err
	}
	conditionID, err := parseHashEnv("CTF_CONDITION_ID")
	if err != nil {
		return nil, err
	}
	indexSets, err := parseBigIntListEnv("CTF_INDEX_SETS")
	if err != nil {
		return nil, err
	}
	parent := parseHashEnvOptional("CTF_PARENT_COLLECTION_ID")

	return &ctf.RedeemPositionsRequest{
		CollateralToken:    collateral,
		ParentCollectionID: parent,
		ConditionID:        conditionID,
		IndexSets:          indexSets,
	}, nil
}

func buildRedeemNegRiskRequest() (*ctf.RedeemNegRiskRequest, error) {
	conditionID, err := parseHashEnv("CTF_CONDITION_ID")
	if err != nil {
		return nil, err
	}
	amounts, err := parseBigIntListEnv("CTF_AMOUNTS")
	if err != nil {
		return nil, err
	}
	return &ctf.RedeemNegRiskRequest{
		ConditionID: conditionID,
		Amounts:     amounts,
	}, nil
}

func parseAddressEnv(key string) (common.Address, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return common.Address{}, fmt.Errorf("%s is required", key)
	}
	if !common.IsHexAddress(raw) {
		return common.Address{}, fmt.Errorf("invalid address for %s", key)
	}
	return common.HexToAddress(raw), nil
}

func parseHashEnv(key string) (common.Hash, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return common.Hash{}, fmt.Errorf("%s is required", key)
	}
	return common.HexToHash(raw), nil
}

func parseHashEnvOptional(key string) common.Hash {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return common.Hash{}
	}
	return common.HexToHash(raw)
}

func parseBigIntEnv(key string) (*big.Int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil, fmt.Errorf("%s is required", key)
	}
	value, ok := new(big.Int).SetString(raw, 0)
	if !ok {
		return nil, fmt.Errorf("invalid %s: %s", key, raw)
	}
	return value, nil
}

func parseBigIntListEnv(key string) ([]*big.Int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil, fmt.Errorf("%s is required", key)
	}
	parts := strings.Split(raw, ",")
	list := make([]*big.Int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, ok := new(big.Int).SetString(part, 0)
		if !ok {
			return nil, fmt.Errorf("invalid %s value: %s", key, part)
		}
		list = append(list, val)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("%s is empty", key)
	}
	return list, nil
}
