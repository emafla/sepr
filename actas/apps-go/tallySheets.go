/*
Copyright 2021 IBM All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
   "bytes"
   "strings"
   "crypto/x509"
   "encoding/json"
   "encoding/csv"
   "encoding/base64"
   "encoding/pem"
   "fmt"
   "os"
   "io"
   "path"
   "time"
   "github.com/hyperledger/fabric-gateway/pkg/client"
   "github.com/hyperledger/fabric-gateway/pkg/identity"
   "github.com/golang/protobuf/proto"
   "github.com/hyperledger/fabric-protos-go/peer"
   "github.com/hyperledger/fabric-protos-go/common"
   "github.com/hyperledger/fabric-protos-go/ledger/rwset"
   msp "github.com/hyperledger/fabric-protos-go/msp"
   "google.golang.org/grpc"
   "google.golang.org/grpc/credentials"
   "github.com/google/uuid"
   shell "github.com/ipfs/go-ipfs-api"
)

type Acta struct {
   Id          string  `json:"id"`
   Cda         int     `json:"cda"`
   Cargo       int     `json:"cargo"`
   Provincia   int     `json:"provincia"`
   Canton      int     `json:"canton"`
   Parroquia   int     `json:"parroquia"`
   Junta       int     `json:"junta"`
   Status      int     `json:"status"`
   Cid         string  `json:"cid"`
   Electores   int     `json:"electores"`
   Blancos     int     `json:"blancos"`
   Nulos       int     `json:"nulos"`
   Candidato1  int     `json:"candidato1"`
   Candidato2  int     `json:"candidato2"`
   Votos       int     `json:"votos"`
}

type  HistoryQueryResult struct{
   Record       *Acta     `json:"record"`
   TxId         string    `json:"txId"`
   Timestamp    time.Time `json:"timestamp"`
   IsDelete     bool      `json:"isDelete"`
}

const (
   mspID        = "CNEMSP"
   cryptoPath   = "../../redSEPR/organizations/peerOrganizations/cne.example.com"
   certPath     = cryptoPath + "/users/User1@cne.example.com/msp/signcerts/cert.pem"
   keyPath      = cryptoPath + "/users/User1@cne.example.com/msp/keystore/"
   tlsCertPath  = cryptoPath + "/peers/peer0.cne.example.com/tls/ca.crt"
   peerEndpoint = "localhost:7051"
   gatewayPeer  = "peer0.cne.example.com"
)

func main() {
   // The gRPC client connection is shared by all Gateway connections to this endpoint
   clientConnection := newGrpcConnection()
   defer clientConnection.Close()

   id := newIdentity()
   sign := newSign()

   // Create a Gateway connection for a specific client identity
   gw, err := client.Connect(
      id,
      client.WithSign(sign),
      client.WithClientConnection(clientConnection),
      // Default timeouts for different gRPC calls
      client.WithEvaluateTimeout(5*time.Second),
      client.WithEndorseTimeout(15*time.Second),
      client.WithSubmitTimeout(5*time.Second),
      client.WithCommitStatusTimeout(1*time.Minute),
   )
   if err != nil {
      panic(err)
   }
   defer gw.Close()

   // Override values for chaincode and channel name; they may differ in testing contexts.
   chaincodeName := "seprcc"
   if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
      chaincodeName = ccname
   }

   channelName := "seprchannel"
   if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
      channelName = cname
   }

   network := gw.GetNetwork(channelName)
   contract := network.GetContract(chaincodeName)
   contracts := network.GetContract("qscc")

   CrearActas(contract)
   GrabarActas(contract)
   // Create tally sheets. Init them with the fixed values. Save them "tallySheets.csv" file
   fmt.Println("\n\n\n\n\n\n\n\n\n\n\n====================================================================")
   fmt.Println("Creating tally sheets with their fixes values; saving them in \"tallySheets.csv\"")
   fmt.Println("That file contains the UUID created for each tally sheet")
   fmt.Println("The UUIDs are used to ID the tally sheets in the transactions below")


   fmt.Println("------------- Start of Simulator -------------")

   for true {
      fmt.Println("\n===============================")
      fmt.Println("1.- Register a tally sheet image (stores TIFF image in IPFS)")
      fmt.Println("2.- Invalidate a tally sheet")
      fmt.Println("3.- Register results of a tally sheet")
      fmt.Println("4.- Display a tally sheet's status")
      fmt.Println("5.- Display election results")
      fmt.Println("6.- Save a tally sheet' history in \"./TXs/UUID-history.csv\")")
      fmt.Println("7.- Save all tally sheets to \"tallySheets.csv\"")
      fmt.Println("8.- Display a transaction (given its TxId)")
      fmt.Println("9.- Exit")
      fmt.Print("\nPlease enter an option: ")
      var op int
      fmt.Scanf("%d", &op)
      switch op{
         case 1:
            fmt.Println("--> Registering the CID of a tally sheet image")
            RegistrarActa(contract)

         case 2:
            fmt.Println("--> Invalidating a tally sheet")
            AnularActa(contract)

         case 3:
            fmt.Println("--> Registering results of a tally sheet")
            RegistrarResultados(contract)

         case 4:
            fmt.Println("--> Displaying a tally sheet's status")
            ConsultarActa(contract)

         case 5:
            fmt.Println("--> Computing Results")
            ConsultarResultados(contract)

         case 6:
            fmt.Println("--> Saving a tally sheet history to file \"UUID-history.csv\"")
            ConsultarHistorial(contract)

         case 7:
            fmt.Println("--> Saving tally sheets to file \"tallySheets.csv\"")
            GrabarActas(contract)

         case 8:
            fmt.Println("--> Displaying a transaction")
            ConsultarTx(contracts, network)

         case 9:
            fmt.Println("------------- End of Simulator -------------")
            os.Exit(0)

         default:
            fmt.Println("Wrong option.")
      }
   }
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
   certificate, err := loadCertificate(tlsCertPath)
   if err != nil {
      panic(err)
   }

   certPool := x509.NewCertPool()
   certPool.AddCert(certificate)
   transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

   connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
   if err != nil {
      panic(fmt.Errorf("failed to create gRPC connection: %w", err))
   }

   return connection
}

// Create a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
   certificate, err := loadCertificate(certPath)
   if err != nil {
      panic(err)
   }

   id, err := identity.NewX509Identity(mspID, certificate)
   if err != nil {
      panic(err)
   }

   return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
   certificatePEM, err := os.ReadFile(filename)
   if err != nil {
      return nil, fmt.Errorf("failed to read certificate file: %w", err)
   }
   return identity.CertificateFromPEM(certificatePEM)
}

// Create a function that generates a digital sign from a hash using a private key.
func newSign() identity.Sign {
   files, err := os.ReadDir(keyPath)
   if err != nil {
      panic(fmt.Errorf("failed to read private key directory: %w", err))
   }
   privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

   if err != nil {
      panic(fmt.Errorf("failed to read private key file: %w", err))
   }

   privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
   if err != nil {
      panic(err)
   }

   sign, err := identity.NewPrivateKeySign(privateKey)
   if err != nil {
      panic(err)
   }

   return sign
}

// Solo se ejecuta una vez para crear las actas (el jueves)
func CrearActas(contract *client.Contract) {
   var actas []Acta
   var actasJSON []byte
   byt, err := os.ReadFile("tallySheets.json")
   if err != nil {
      panic(fmt.Errorf("failed to read the tally sheets JSON file: %w", err))
   }
   err = json.Unmarshal(byt, &actas)
   if err != nil {
       panic(fmt.Errorf("failed to unmarshal the tally sheets. %w", err))
   }
   for i := range actas {
      actas[i].Id = uuid.NewString()
   }
   actasJSON, err1 := json.Marshal(actas)
   if err1 != nil {
        panic(fmt.Errorf("Error encoding JSON: %w", err1))
        return
   }

   _, err = contract.SubmitTransaction("CrearActas", string(actasJSON))
   if err != nil {
      panic(fmt.Errorf("failed to submit the transaction: %w", err))
   }

   fmt.Printf("Tally sheets have been created successfully\n")
}

func RegistrarActa(contract *client.Contract) {
   var id, cid string
   id = getUUID()
   if id == "" {
      return
   }
   cid = getCid()
   if cid == "" {
      return
   }

   _, err := contract.SubmitTransaction("RegistrarActa", id, cid)
   if err != nil {
      fmt.Printf ("failed to submit the transaction: %v", err)
      return
   }

   fmt.Println("Tally sheet image stored in IFPS and its CID has been registered")
}

func AnularActa(contract *client.Contract) {
   var id string
   id = getUUID()
   if id == "" {
      return
   }

   _, err := contract.SubmitTransaction("AnularActa", id)
   if err != nil {
      fmt.Printf("failed to submit the transaction: %v", err)
      return
   }

   fmt.Println("Tally sheet has been invalidates")
}

func RegistrarResultados(contract *client.Contract) {
   var id, bl, nl, c1, c2, v string
   id = getUUID()
   if id == "" {
      return
   }
   bl, nl, c1, c2, v = getResults()

   _, err := contract.SubmitTransaction("RegistrarResultados", id, bl, nl, c1, c2, v)
   if err != nil {
      fmt.Printf("failed to submit the transaction: %v", err)
      return
   }
   fmt.Println("Resultados have been registered")
}


// Consultar los resultados totales
func ConsultarResultados(contract *client.Contract) {
   var result []byte
   result, err := contract.EvaluateTransaction("ListarActas")
   if err != nil {
      fmt.Printf ("failed to submit the transaction: %v", err)
      return
   }
   computeResults(result)
  }

// Consultar estado de una acta, por su ID
func ConsultarActa(contract *client.Contract) {
   var id string
   id = getUUID()
   if id == "" {
      return
   }

   evaluateResult, err := contract.EvaluateTransaction("ConsultarActa", id)
   if err != nil {
      fmt.Printf("failed to submit the transaction: %v", err)
   }

   result := formatJSON(evaluateResult)
   if result == "" {
      return
   }

   fmt.Printf("Tally Sheet %s:\n%s\n", id, result)

   // CID of the acta file
   var acta Acta
   if err := json.Unmarshal([]byte(result), &acta); err != nil {
      fmt.Printf("failed to unmarshal JSON string: %v", err)
      return
    }
   cid := acta.Cid

   if cid != "" {
      // Connect to the local IPFS node
      sh := shell.NewShell("localhost:5001")

      // Create a local file to save the retrieved content
      outFile, err := os.Create("./tallySheetsCID/" + cid + ".tif")
      if err != nil {
         fmt.Printf("failed to create local IPFS file for tally sheet image: %v", err)
         return
      }
      defer outFile.Close()

      // Retrieve the file from IPFS
      reader, err := sh.Cat(cid)
      if err != nil {
         fmt.Printf("failed to retrieve tally sheet image file from IPFS: %v", err)
         return
      }
      defer reader.Close()

      // Copy the content to the local file
      _, err = io.Copy(outFile, reader)
      if err != nil {
         fmt.Printf ("failed to copy tally sheet from IPFS to local file: %v", err)
         return
      }

      fmt.Println("Tally sheet copied from IPFS to local file: %s", "tallySheetsCID/"+cid+".tif")
   }
}

func ConsultarHistorial(contract *client.Contract) {
   var records []HistoryQueryResult
   var result []byte
   var id string
   id = getUUID()
   if id == "" {
      return
   }

   result, err := contract.SubmitTransaction("HistorialActa", id)
   if err != nil {
      fmt.Printf("failed to evaluate transaction: %v", err)
      return
   }

   err = json.Unmarshal(result, &records)
   if err != nil {
      fmt.Printf("failed to unmarshal JSON to slice: %v", err)
      return
   }

   csvFile, err1 := os.Create("./TXs/" + id + "-history.csv")
   if err1 != nil {
      fmt.Printf("failed to create file: %v", err1)
       return
   }
   defer csvFile.Close()

   writer := csv.NewWriter(csvFile)
   defer writer.Flush()

   header := []string{"UUID", "CDA", "Cargo", "Provincia", "Canton", "Parroquia",
      "Junta", "Status", "Electores", "Blancos", "Nulos", "Candidato1",
      "Candidato2", "Votos", "CID", "TxId", "Time", "Eliminado?"}
   if err = writer.Write(header); err != nil {
      return
   }

   for _, r := range records {
      var csvRow []string
      csvRow = append(csvRow, r.Record.Id, fmt.Sprint(r.Record.Cda),
         fmt.Sprint(r.Record.Cargo), fmt.Sprint(r.Record.Provincia),
         fmt.Sprint(r.Record.Canton), fmt.Sprint(r.Record.Parroquia),
         fmt.Sprint(r.Record.Junta), fmt.Sprint(r.Record.Status),
         fmt.Sprint(r.Record.Electores), fmt.Sprint(r.Record.Blancos),
         fmt.Sprint(r.Record.Nulos), fmt.Sprint(r.Record.Candidato1),
         fmt.Sprint(r.Record.Candidato2), fmt.Sprint(r.Record.Votos), r.Record.Cid,
         r.TxId, fmt.Sprint(r.Timestamp), fmt.Sprint(r.IsDelete))
      if err = writer.Write(csvRow); err != nil {
         return
      }
   }
   showResults(string(result))
}

// Listar todas las actas
func GrabarActas(contract *client.Contract) {
   var actas []Acta
   evaluateResult, err := contract.EvaluateTransaction("ListarActas")
   if err != nil {
      fmt.Printf("failed to submit the transaction: %v", err)
      return
   }

   err = json.Unmarshal(evaluateResult, &actas)
   if err != nil {
      fmt.Printf ("failed to unmarshal JSON to TallySheet slice: %v", err)
      return
   }

   csvFile, err1 := os.Create("tallySheets.csv")
   if err1 != nil {
       return
   }
   defer csvFile.Close()

   writer := csv.NewWriter(csvFile)
   defer writer.Flush()

   header := []string{"UUID", "CDA", "Cargo", "Provincia", "Canton", "Parroquia",
      "Junta", "Status", "Electores", "Blancos", "Nulos", "Candidato1",
      "Candidato2", "Votos", "CID"}
   if err = writer.Write(header); err != nil {
      return
   }

   for _, r := range actas {
      var csvRow []string
      csvRow = append(csvRow,r.Id, fmt.Sprint(r.Cda), fmt.Sprint(r.Cargo),
         fmt.Sprint(r.Provincia), fmt.Sprint(r.Canton), fmt.Sprint(r.Parroquia),
         fmt.Sprint(r.Junta), fmt.Sprint(r.Status), fmt.Sprint(r.Electores),
         fmt.Sprint(r.Blancos), fmt.Sprint(r.Nulos), fmt.Sprint(r.Candidato1),
         fmt.Sprint(r.Candidato2), fmt.Sprint(r.Votos), r.Cid)
      if err = writer.Write(csvRow); err != nil {
         return
      }
   }
}


// Consultar estado de una acta, por su ID
func ConsultarTx(contracts *client.Contract, network *client.Network) {
   var txId string
   txId = getTxId()
   if txId == "" {
      return
   }

   txBytes, err := contracts.EvaluateTransaction("GetTransactionByID", network.Name(), txId);
   if err != nil {
      fmt.Printf ("failed to get transaction: %v", err)
      return
   }

   var processedTx peer.ProcessedTransaction
   err = proto.Unmarshal(txBytes, &processedTx)
   if err != nil {
      fmt.Printf ("failed to unmarshal ProcessedTransaction: %v", err)
      return
   }
   fmt.Printf ("valCode: %s\n", peer.TxValidationCode_name[processedTx.ValidationCode])

   envelope := processedTx.TransactionEnvelope
   var payload common.Payload
   err = proto.Unmarshal(envelope.Payload, &payload)
   if err != nil {
      fmt.Printf("failed to unmarshal Payload: %v", err)
      return
   }
   var ccHeader common.ChannelHeader
   err = proto.Unmarshal(payload.Header.ChannelHeader, &ccHeader)
   if err != nil {
      fmt.Printf("failed to unmarshal ChannelHeader: %v", err)
      return
   }
   fmt.Printf("chHdrType: %s\nVersion: %d\nTS: %s\nccID: %s\nTxId: %s\nEp: %d\n",
      common.HeaderType_name[ccHeader.Type], ccHeader.Version, ccHeader.Timestamp.AsTime().Format(time.RFC1123Z), ccHeader.ChannelId, ccHeader.TxId, ccHeader.Epoch)
   if len(ccHeader.TlsCertHash) > 0 {
      fmt.Printf("TlsCertHash: %s\n",
      base64.StdEncoding.EncodeToString(ccHeader.TlsCertHash))
   }

   var signHeader common.SignatureHeader
   err = proto.Unmarshal(payload.Header.SignatureHeader, &signHeader)
   if err != nil {
      fmt.Printf("failed to unmarshal SignatureHeader: %v", err)
      return
   }

   id, err := deserializeIdentity(signHeader.Creator)
   if err != nil {
      fmt.Println("failed to deserialize identity:", err)
      return
   }
   fmt.Println("Creator\nMSP ID:", id.Mspid)
   cert, err := parseCertificate(id)
   if err != nil {
      fmt.Println("Error parsing certificate:", err)
      return
   }
   fmt.Println("Certificate Subject:", cert.Subject)
   fmt.Println("Certificate Issuer:", cert.Issuer)


   fmt.Printf("Nonce: %s\n", base64.StdEncoding.EncodeToString(signHeader.Nonce))

   fmt.Printf("==========\n")
   err = decodeTransactionPayload(envelope.Payload)
   if err != nil {
      fmt.Println("Error decoding transaction payload:", err)
      return
   }

   return
}


// Format JSON data
func formatJSON(data []byte) string {
   var prettyJSON bytes.Buffer
   if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
      fmt.Printf ("\n*** failed to parse JSON: %v", err)
      return ""
   }
   return prettyJSON.String()
}

// Obtener ID de acta
func getUUID() (string){
   var uuidS string
   fmt.Print("Enter tally sheet UUID (get from \"tallySheets.csv\" file): ")
   fmt.Scanf("%s", &uuidS)
   if err := uuid.Validate(uuidS); err != nil {
      fmt.Printf ("\n*** Please, enter a valid UUID")
      return ""
   }
   return uuidS
}

// Obtener TXID de transaccion
func getTxId() (string){
   var txid string
   fmt.Print("Enter transaction ID (get from \"UUID-history.csv\" file): ")
   fmt.Scanf("%s", &txid)
   if len(txid) != 64 {
      fmt.Printf("\n*** Please, enter a valid TxID")
      return ""
   }
   return txid
}

func getCid() (string){
   var fname string
   fmt.Print("Enter tally sheet image file name (e.g. tallySheets/A...tif): ")
   fmt.Scanf("%s", &fname)

   sh := shell.NewShell("localhost:5001")

   // Read the file you want to add
   file, err := os.Open(fname)
   if err != nil {
      fmt.Printf ("\n*** failed to open tally sheet TIF file: %v", err)
      return ""
   }
   defer file.Close()

    // Add the file to IPFS
   cid, err := sh.Add(file)
   if err != nil {
      fmt.Printf("\n*** failed to add file to IPFS: %v", err)
      return ""
   }

   return cid
}

func getResults() (string, string, string, string, string) {
   var bl, nl, c1, c2, v string
   fmt.Print("Enter number of blank votes: ")
   fmt.Scanf("%s", &bl)
   fmt.Print("Enter number of spoiled votes: ")
   fmt.Scanf("%s", &nl)
   fmt.Print(":Enter number of candidate 1 votes ")
   fmt.Scanf("%s", &c1)
   fmt.Print(":Enter number of candidate 2 votes ")
   fmt.Scanf("%s", &c2)
   fmt.Print("Enter number of voters: ")
   fmt.Scanf("%s", &v)

   return bl, nl, c1, c2, v
}

func computeResults(result []byte){
   var actas []Acta
   err := json.Unmarshal(result, &actas)
   if err != nil {
      fmt.Printf("\n*** failed to unmarshal JSON to TallySheet slice: %v", err)
      return
   }

   tElect := 0
   tNulos := 0
   tBlanc := 0
   tCand1 := 0
   tCand2 := 0
   tVotos := 0

   // Iterate over the vector to calculate the totals
   for _, item := range actas {
      tCand1 += item.Candidato1
      tCand2 += item.Candidato2
      tElect += item.Electores
      tNulos += item.Nulos
      tBlanc += item.Blancos
      tVotos += item.Votos
   }

   // Print the totals
   fmt.Printf("Total registered voters: %d\n", tElect)
   fmt.Printf("Blank votes            : %d\n", tBlanc)
   fmt.Printf("Spoiled Votes          : %d\n", tNulos)
   fmt.Printf("Candidato 1 votes      : %d\n", tCand1)
   fmt.Printf("Candidato 2 votes      : %d\n", tCand2)
   fmt.Printf("Total voters           : %d\n", tVotos)
}

func showResults(result string){
   result = strings.Replace( result , ",\"" , ",   \"" , -1)
   result = strings.Replace( result , "," , ",\n" , -1)
   result = strings.Replace( result , "{" , "{\n   " , -1)
   result = strings.Replace( result , "}" , "\n}" , -1)

   fmt.Println(result)
}

func deserializeIdentity(serialized []byte) (*msp.SerializedIdentity, error) {
   identity := &msp.SerializedIdentity{}
   err := proto.Unmarshal(serialized, identity)
   if err != nil {
      return nil, fmt.Errorf("\n*** failed to unmarshal SerializedIdentity: %w", err)
   }
   return identity, nil
}

func parseCertificate(identity *msp.SerializedIdentity) (*x509.Certificate, error) {
   block, _ := pem.Decode(identity.IdBytes)
   if block == nil {
      return nil, fmt.Errorf("\n*** failed to decode PEM block from identity bytes")
   }

   cert, err := x509.ParseCertificate(block.Bytes)
   if err != nil {
      return nil, fmt.Errorf("\n*** failed to parse x509 certificate: %w", err)
   }
   return cert, nil
}


func decodeTransactionPayload(payloadBytes []byte) error {
   // Unmarshal the Payload
   payload := &common.Payload{}
   err := proto.Unmarshal(payloadBytes, payload)
   if err != nil {
      return fmt.Errorf("\n*** failed to unmarshal payload: %w", err)
   }

   // Extract Channel Header
   header := &common.ChannelHeader{}
   err = proto.Unmarshal(payload.Header.ChannelHeader, header)
   if err != nil {
      return fmt.Errorf("\n*** failed to unmarshal channel header: %w", err)
   }

   fmt.Println("Channel ID:", header.ChannelId)
   fmt.Println("Transaction ID:", header.TxId)

   // Extract the Transaction
   transaction := &peer.Transaction{}
   err = proto.Unmarshal(payload.Data, transaction)
   if err != nil {
      return fmt.Errorf("\n*** failed to unmarshal transaction: %w", err)
   }

   // Process each Action in the transaction
   for _, action := range transaction.Actions {
      // Extract ChaincodeActionPayload
      chaincodeActionPayload := &peer.ChaincodeActionPayload{}
      err = proto.Unmarshal(action.Payload, chaincodeActionPayload)
      if err != nil {
         return fmt.Errorf("\n*** failed to unmarshal chaincode action payload: %w", err)
      }

      // Extract ProposalResponsePayload
      proposalResponsePayload := &peer.ProposalResponsePayload{}
      err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload)
      if err != nil {
         return fmt.Errorf("\n*** failed to unmarshal proposal response payload: %w", err)
      }

      // Extract ChaincodeAction
      chaincodeAction := &peer.ChaincodeAction{}
      err = proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction)
      if err != nil {
         return fmt.Errorf("\n*** failed to unmarshal chaincode action: %w", err)
      }

      // Print Chaincode Response
      fmt.Println("Chaincode Response Status:", chaincodeAction.Response.Status)
      fmt.Println("Chaincode Response Message:", chaincodeAction.Response.Message)
      fmt.Println("Chaincode Response Payload:", string(chaincodeAction.Response.Payload))

      // Print ReadWriteSet (KVs modified)
      readWriteSet := &rwset.TxReadWriteSet{}
      err = proto.Unmarshal(chaincodeAction.Results, readWriteSet)
      if err != nil {
         return fmt.Errorf("\n*** failed to unmarshal read-write set: %w", err)
      }

      rwSetJSON, _ := json.MarshalIndent(readWriteSet, "", "  ")

      fmt.Println("ReadWriteSet:", string(rwSetJSON))
   }

   return nil
}
