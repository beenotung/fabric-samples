/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

const KeyList = "_KEY_LIST_"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("exDemo Init")
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting no arguments.")
	}

	fmt.Printf("Created New Key-Value Map.\n")

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	fmt.Printf("%s(%q)\n", function, args)
	if function == "insert" {
		return t.insert(stub, args)
	} else if function == "update" {
		return t.update(stub, args)
	} else if function == "key_search" {
		return t.keySearch(stub, args)
	} else if function == "value_search" {
		return t.valueSearch(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting {insert, update, key_search, value_search}.")
}

func (t *SimpleChaincode) keys(stub shim.ChaincodeStubInterface) (keyMap map[string]byte, err error) {
	keyListBytes, err := stub.GetState(KeyList)
	if err != nil {
		return
	}
	keyMap = make(map[string]byte)
	if keyListBytes != nil {
		err = json.Unmarshal(keyListBytes, &keyMap)
		if err != nil {
			return
		}
	}
	return
}

func (t *SimpleChaincode) set(stub shim.ChaincodeStubInterface, key string, value []byte) pb.Response {
	keyMap, err := t.keys(stub)
	if err != nil {
		return shim.Error("Failed to decode key list: " + err.Error())
	}
	keyMap[key] = 0
	keyListBytes, err := json.Marshal(keyMap)
	if err != nil {
		return shim.Error("Failed to encode key list: " + err.Error())
	}
	err = stub.PutState(KeyList, keyListBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) insert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2.")
	}
	key := args[0]
	if key == KeyList {
		return shim.Error(fmt.Sprintf("Invalid key {%s} is preseved.", KeyList))
	}
	if len(key) == 0 {
		return shim.Error("Invalid key, expecting non-empty string.")
	}
	oldValue, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	newValue := args[1]
	var value []byte
	if oldValue == nil || len(oldValue) == 0 {
		value = []byte(newValue)
	} else {
		value = append(oldValue, ";"+newValue...)
	}
	fmt.Printf("new value for {%s} = {%s}\n", key, value)
	return t.set(stub, key, value)
}

func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2.")
	}
	key := args[0]
	if key == KeyList {
		return shim.Error(fmt.Sprintf("Invalid key {%s} is preseved.", KeyList))
	}
	if len(key) == 0 {
		return shim.Error("Invalid key, expecting non-empty string.")
	}
	value := args[1]
	fmt.Printf("new value for {%s} = {%s}\n", key, value)
	return t.set(stub, key, []byte(value))
}

func (t *SimpleChaincode) keySearch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}
	key := args[0]
	if key == KeyList {
		return shim.Error(fmt.Sprintf("Invalid key {%s} is preseved.", KeyList))
	}
	if len(key) == 0 {
		return shim.Error("Invalid key, expecting non-empty string.")
	}
	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("value of {%s} = {%s}\n", key, value)
	return shim.Success(value)
}

func equals(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func (t *SimpleChaincode) valueSearch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}
	keyMap, err := t.keys(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	targetValue := args[0]
	targetValueBytes := []byte(targetValue)
	keys := make([]string, 0)
	for key := range keyMap {
		value, err := stub.GetState(key)
		if err != nil {
			return shim.Error(err.Error())
		}
		if equals(value, targetValueBytes) {
			keys = append(keys, key)
		}
	}
	fmt.Printf("keys of {%s} = %q\n", targetValue, keys)
	keysBytes, err := json.Marshal(keys)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(keysBytes)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
