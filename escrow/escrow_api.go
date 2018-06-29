package escrow

import (
	"fmt"
	"github.com/themis-network/go-themis/rpc"
	"log"
	"encoding/json"
	"math/big"
)

var(
	modules = []string{"escrow"}
	cors = []string{"*"}
	vhosts = []string{"*"}
)

// PublicWeb3API offers helper utils
type EscrowAPI struct {
	escrowNode *EscrowNode
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (t *EscrowNode) startApiServer(){

	log.Println("Start Escrow api service...")

	apis := []rpc.API{
		{
			Namespace: "escrow",
			Version:   "1.0",
			Service:   NewEscrowAPI(t),
			Public:    true,
		},
	}

	rpc.StartHTTPEndpoint(t.config.Endpoint, apis, modules, cors, vhosts)
}

// NewPublicWeb3API creates a new Web3Service instance
func NewEscrowAPI(t *EscrowNode) *EscrowAPI {
	return &EscrowAPI{
		escrowNode: t,
	}
}

/**
 GetDecryptSecret API, RPC "method":"escrow_getDecryptSecret"
1. verify the order's arbitrate result
2. try get decrypt fragment from map. if fail, get the fragment from contract, then decrypt it
 */
func (t *EscrowAPI) GetDecryptSecret(orderId int64) (string, error){
	log.Println("Request escrow_getDecryptSecret, orderId:", orderId)

	if v, ok := t.escrowNode.secrets[orderId]; ok {
		return v, nil
	}

	var winner *big.Int
	if v, ok := t.escrowNode.orderWinner[orderId]; ok {
		winner = v
	}else {
		w, err := t.escrowNode.getWinner(orderId)
		if err != nil {
			winner = nil
			log.Println("get winner error")
		}else {
			winner = w
		}
	}

	if winner == nil || winner.Int64() == 0{
		//no winner error
		log.Println("no winner error, orderid: ", orderId)
		return "", &noWinnerError{"no winner error"}
	}

	log.Println("winner is: ", winner.Int64())

	sectet, err := t.escrowNode.getFragment(orderId, winner)
	log.Println("secret from contract: ", sectet)
	decSectet, err:= t.escrowNode.decrypt(sectet)

	if err != nil {
		log.Println("decrypt secret error, ", err)
		return "", &decryptError{fmt.Sprintf("decrypt secret error, %v", err)}
	}
	return decSectet, nil
}


func errorJson(code int, message string) string{
	errorJson := &jsonError{Code: code, Message: message}
	jsons, errs := json.Marshal(errorJson)
	if errs != nil {
		fmt.Println(errs.Error())
		return ""
	}
	return string(jsons)
}