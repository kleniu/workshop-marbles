/*
 * The test chaincode
 */

package main

import (
	"bytes"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"strconv"
	"crypto/x509"
	"encoding/pem"
	"time"
	"math/big"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	mp "github.com/hyperledger/fabric/protos/msp"
	"github.com/golang/protobuf/proto"
)

// Define the Smart Contract structure
type SmartContract struct {
}

func formatSerial(serial *big.Int) string {
	b := serial.Bytes()
	buf := make([]byte, 0, 3*len(b))
	x := buf[1*len(b) : 3*len(b)]
	hex.Encode(x, b)
	for i := 0; i < len(x); i += 2 {
		buf = append(buf, x[i], x[i+1], ':')
	}
	return string(buf[:len(buf)-1])
}


func parseCert(serializedID []byte) {
	sId := &mp.SerializedIdentity{}
	err := proto.Unmarshal(serializedID, sId)
	if err != nil {
		fmt.Println("Could not deserialize a SerializedIdentity, err %s", err)
	}
	// get the MSP name
	fmt.Println("MSP: " + sId.Mspid)

	bl, _ := pem.Decode(sId.IdBytes)
	if bl == nil {
		fmt.Println("Failed to decode PEM structure")
	}

	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		fmt.Println("Unable to parse certificate %s", err)
	}
	
	// print the cert
        fmt.Println("Cert SerialNumber: " + formatSerial(cert.SerialNumber) ) 
        fmt.Println("Cert Subject CommonName :" + cert.Subject.CommonName)
}


/*
 * The Init method is called when the chaincode is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("Calling instantiate method.")

	// method to just get the identity of the submitter of the transaction 

	serializedID, _ := APIstub.GetCreator() 
	parseCert(serializedID)

	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("Calling Invoke method.") 

	serializedID, _ := APIstub.GetCreator() 
	parseCert(serializedID)

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	fmt.Println("Function name: " + function)

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryAsset" {
		return s.queryAsset(APIstub, args)
	} else if function == "makeAsset" {
		return s.makeAsset(APIstub, args)
	} else if function == "changeAsset" {
		return s.changeAsset(APIstub, args)
	} else if function == "deleteAsset" {
		return s.deleteAsset(APIstub, args) 
	} else if function == "listHistory" {
		return s.listHistory(APIstub, args) 
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1. Please provide KEY of the asset.")
	}

	valAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		fmt.Println("ERR in GetState. err.Error()=" + err.Error())
		return shim.Error("ERR in GetState. err.Error()=" + err.Error())
	} else if len(valAsBytes) == 0 {
		fmt.Println("ERR the VAL is empty for KEY=" + args[0])
		return shim.Error("ERR the VAL is empty for KEY=" + args[0])
	}

	fmt.Println("OK Retrived KEY: >" + args[0] + "< VAL: >" + string(valAsBytes[:]) + "<")
	
	return shim.Success(valAsBytes)
}

func (s *SmartContract) makeAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. Please provide KEY and VAL of the asset.")	
	}

	aKeyAsStr        := args[0]
	aValAsBytes, err := json.Marshal(args[1])

	if err != nil {
		fmt.Println("ERR in json.Marshal(" + args[1] + "). err.Error()=" + err.Error())
		return shim.Error("ERR in json.Marshal(" + args[1] + "). err.Error()=" + err.Error())
	}

	APIstub.PutState(aKeyAsStr, aValAsBytes)
	fmt.Println("Added KEY=" + aKeyAsStr + " VAL=" + string(aValAsBytes[:]))

	return shim.Success(nil)
}


func (s *SmartContract) listHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	assetName := args[0]

	fmt.Printf("- start listHistory: %s\n", assetName)

	resultsIterator, err := APIstub.GetHistoryForKey(assetName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- listHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	aKeyAsStr        := args[0]
	aValAsBytes, err := json.Marshal(args[1])

	if err != nil {
		fmt.Println("ERR in json.Marshal(" + args[1] + "). err.Error()=" + err.Error())
		return shim.Error("ERR in json.Marshal(" + args[1] + "). err.Error()=" + err.Error())
	}

	APIstub.PutState(aKeyAsStr, aValAsBytes)
	fmt.Println("Changed KEY=" + aKeyAsStr + " VAL=" + string(aValAsBytes[:]))

	return shim.Success(nil)
}


func (s *SmartContract) deleteAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	aKeyAsStr        := args[0]

	APIstub.DelState(aKeyAsStr)
	fmt.Println("Deleted KEY=" + aKeyAsStr )

	return shim.Success(nil)
}


// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
