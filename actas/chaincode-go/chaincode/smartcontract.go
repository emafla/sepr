package chaincode

import (
   "encoding/json"
   "fmt"
   "log"
   "time"
   "github.com/golang/protobuf/ptypes"
   "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct { // SmartContract provides functions for managing an Acta
   contractapi.Contract
}

// Acta describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
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

//  Estructura para recibir el historico de Acta
type  HistoryQueryResult struct{
   Record       *Acta     `json:"record"`
   TxId         string    `json:"txId"`
   Timestamp    time.Time `json:"timestamp"`
   IsDelete     bool      `json:"isDelete"`
}


// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) CrearActas(ctx contractapi.TransactionContextInterface, str string) error {
   byt := []byte(str)
   var actas []Acta
   err := json.Unmarshal(byt, &actas)
   if err != nil {
       return fmt.Errorf("no pudo unmarshal el string al slice de actas. %v", err)
   }
   for _, acta := range actas {
      actaJSON, err := json.Marshal(acta)
      if err != nil {
         return fmt.Errorf("no pudo marshal el acta a JSON. %v", err)
      }

      err = ctx.GetStub().PutState(acta.Id, actaJSON)
      if err != nil {
         return fmt.Errorf("no pudo grabar el acta al world state. %v", err)
      }
   }
   return nil
}


// Retorna acta con Id, si dicha acta existe
func (s *SmartContract) ConsultarActa(ctx contractapi.TransactionContextInterface, id string) (*Acta, error) {
   actaJSON, err := ctx.GetStub().GetState(id)
   if err != nil {
      return nil, fmt.Errorf("failed to read from world state: %v", err)
   }
   if actaJSON == nil {
      return nil, fmt.Errorf("El acta %s no existe", id)
   }

   var acta Acta
   err = json.Unmarshal(actaJSON, &acta)
   if err != nil {
      return nil, fmt.Errorf("no pudo unmarshal JSON a Acta. %v", err)
   }
   return &acta, nil
}

// Registra acta escaneada con CID
func  (s *SmartContract) RegistrarActa(ctx contractapi.TransactionContextInterface, id string, cid string) error {

   acta, err := s.ConsultarActa(ctx, id)
   if err != nil {
      return fmt.Errorf("error al consultar acta. %v", err)
   }

   if acta.Status == 1 || acta.Status == 2 {
      return fmt.Errorf("acta ya esta registrada")
   }

   acta.Cid = cid
   acta.Status = 1
   actaJSON, err := json.Marshal(acta)
   if err != nil {
      return fmt.Errorf("error al marshal Acta a JSON. %v", err)
   }
   err = ctx.GetStub().PutState(id, actaJSON)
   if err != nil {
      return fmt.Errorf("no pudo grabar acta al world state. %v", err)
   }
   return nil
}

// Anula acta registrada o con resultados
func  (s *SmartContract) AnularActa(ctx contractapi.TransactionContextInterface, id string) error {
   acta, err := s.ConsultarActa(ctx, id)
   if err != nil {
      return fmt.Errorf("error al consultar acta. %v", err)
   }

   if acta.Status == 0 {
      return fmt.Errorf("acta no esta registrada")
   }

   if acta.Status == 3 {
      return fmt.Errorf("acta ya esta anulada")
   }

   if acta.Status != 1 && acta.Status != 2 {
      return fmt.Errorf("acta no esta registrada o no tiene resultados")
   }

   acta.Status = 3
   acta.Cid = ""
   acta.Blancos = 0
   acta.Nulos = 0
   acta.Candidato1 = 0
   acta.Candidato2 = 0
   acta.Votos = 0

   actaJSON, err := json.Marshal(acta)
   if err != nil {
      return fmt.Errorf("error al marshal Acta a JSON. %v", err)
   }
   err = ctx.GetStub().PutState(acta.Id, actaJSON)
   if err != nil {
      return fmt.Errorf("No pudo grabar acta al world state. %v", err)
   }
   return nil
}


// Registra resultados de acta
func (s *SmartContract) RegistrarResultados(ctx contractapi.TransactionContextInterface, id string, bl, nl, c1, c2, v int) error {

   acta, err := s.ConsultarActa(ctx, id)
   if err != nil {
      return fmt.Errorf("error al consultar acta. %v", err)
   }

   if acta.Status != 1 {
      return fmt.Errorf("acta no esta registrada")
   }

   acta.Status = 2
   acta.Blancos = bl
   acta.Nulos = nl
   acta.Candidato1 = c1
   acta.Candidato2 = c2
   acta.Votos = v

   actaJSON, err := json.Marshal(acta)
   if err != nil {
      return fmt.Errorf("no pudo marsha Acta a JSON. %v", err)
      return err
   }
   err = ctx.GetStub().PutState(acta.Id, actaJSON)
   if err != nil {
      return fmt.Errorf("no pudo grabar acta al world state. %v", err)
   }
   return nil
 }


//  Historial de una acta - devuelve todos los registros de una acta, recorriendo la cadena de bloques
func  (s *SmartContract) HistorialActa(ctx contractapi.TransactionContextInterface, id string) ([]HistoryQueryResult, error){
   log.Print("GetAssetHistory: ID %v", id)

   resultsIterator, err:= ctx.GetStub().GetHistoryForKey(id)
   if err != nil{
      return nil, fmt.Errorf("no se pudo obtener historia del acta. %v", err)
   }
   defer resultsIterator.Close()

   var records []HistoryQueryResult
   for resultsIterator.HasNext(){
      response, err := resultsIterator.Next()
      if err != nil{
         return nil, fmt.Errorf("no pudo obtener siguiente version del acta. %v", err)
      }

     var acta Acta
     if len(response.Value) > 0 {
        err=json.Unmarshal(response.Value, &acta)
        if err != nil{
           return nil, fmt.Errorf("no pudo unmarshall JSON a Acta. %v", err)
        }
     } else {
        acta = Acta{
           Id: id,
        }
     }
     timestamp, err := ptypes.Timestamp(response.Timestamp)
     if err != nil{
        return nil, fmt.Errorf("no pudo obtener el timestamp del acta. %v", err)
     }

     record := HistoryQueryResult{
        TxId: response.TxId,
        Timestamp: timestamp,
        Record: &acta,
        IsDelete: response.IsDelete,
      }
      records = append(records, record)
   }
   return records, nil
}


// Obtener todas las actas en el world state 
func (s *SmartContract) ListarActas(ctx contractapi.TransactionContextInterface) ([]*Acta, error) {
   // range query with empty string for startKey and endKey does an
   // open-ended query of all assets in the chaincode namespace.
   resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
   if err != nil {
      return nil, fmt.Errorf("no pudo obtener las actas del world state. %v", err)
   }
   defer resultsIterator.Close()

   var actas []*Acta
   for resultsIterator.HasNext() {
      queryResponse, err := resultsIterator.Next()
      if err != nil {
         return nil, fmt.Errorf("error al obtener la siguiente acta. %v", err)
      }

      var acta Acta
      err = json.Unmarshal(queryResponse.Value, &acta)
      if err != nil {
         return nil, fmt.Errorf("error al unmarshal JSON a Acta. %v", err)
      }
      actas = append(actas, &acta)
   }

   return actas, nil
}

func  (s *SmartContract) CheckUUID(ctx contractapi.TransactionContextInterface, uuid string) (bool, error) {
  assetJSON, err := ctx.GetStub().GetState(uuid)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }
  return assetJSON != nil, nil
}
