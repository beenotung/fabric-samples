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
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"time"
	"crypto/sha256"
	"encoding/hex"
)

const KeyList = "_KEY_LIST_"

var OK = []byte("OK")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const (
	DataType_       = iota // make it start from 1
	DataTypePublish = iota
	DataTypeRequest = iota
	DataTypeReply   = iota
)

// util types
type User struct {
	Name string `json:"name"`
}
type MultiHash struct {
	Method string `json:"method"`
	Digest string `json:"digest"`
}

// general container
type DataContainer struct {
	Type int
	TxId string
	Data interface{}
}

// data types
type DataPublish struct {
	Content []byte
	Date    time.Time
	Hash    MultiHash
	Owner   User
}
type DataRequest struct {
	PublishTxId string
	Requester   User
}
type DataReply struct {
	RequestTxId string
	Answer      bool
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	stub.GetTxID()
	fmt.Println("exDemo Init")
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting no arguments.")
	}

	fmt.Printf("Created New Data Sharing World.\n")

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("data-sharing Invoke")
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

func (t *SimpleChaincode) PublishData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Error: missing arguments for PublishData")
	}
	if len(args) > 2 {
		return shim.Error("Error: too mush arguments for PublishData")
	}
	owner := User{}
	err := json.Unmarshal([]byte(args[1]), &owner)
	if err != nil {
		return shim.Error(err.Error())
	}
	data := DataPublish{
		Content: []byte( args[0]),
		Date:    time.Now(),
		Owner:   owner,
	}
	hash, err := t.Hash(data)
	if err != nil {
		return shim.Error(err.Error())
	}
	data.Hash = hash
	txId := stub.GetTxID()
	container := DataContainer{
		Type: DataTypePublish,
		TxId: txId,
		Data: data,
	}
	// TODO update index
	err = t.PutState(stub, txId, container)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(OK)
}
func (t *SimpleChaincode) ShowMetaData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Error("not impl")
}
func (t *SimpleChaincode) ShowPendingRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Error("not impl")
}
func (t *SimpleChaincode) HandleRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Error("not impl")
}
func (t *SimpleChaincode) ShowInfoAboutData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Error("not impl")
}
func (t *SimpleChaincode) RequestData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Error("not impl")
}

func (t *SimpleChaincode) Hash(data interface{}) (hash MultiHash, err error) {
	bs, err := json.Marshal(data)
	if err != nil {
		return
	}
	algo := sha256.New()
	algo.Write(bs)
	hash = MultiHash{
		Method: "sha256",
		Digest: hex.EncodeToString(algo.Sum(nil)),
	}
	return
}
func (t *SimpleChaincode) PutState(stub shim.ChaincodeStubInterface, key string, value interface{}) error {
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = stub.PutState(key, bs)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
