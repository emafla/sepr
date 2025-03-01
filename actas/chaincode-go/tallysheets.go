/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode"
)

func main() {
	actaChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error al crear chaincode actas: %v", err)
	}

	if err := actaChaincode.Start(); err != nil {
		log.Panicf("Error al inciar chaincode actas: %v", err)
	}
}
