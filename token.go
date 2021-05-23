package main

import (
	"encoding/binary"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"strconv"
)

type TokenCC struct {
}

var balances map[string]uint64
var exists map[string]bool
var allowed map[string]map[string]uint64
var totalSupply uint64

var functions = map[string]func(args []string, stub shim.ChaincodeStubInterface) peer.Response{
	"balanceOf": func(args []string, stub shim.ChaincodeStubInterface) peer.Response {
		address := args[0]

		if exists[address] == false {
			message := fmt.Sprint("%s doesn't exist", address)
			return shim.Error(message)
		}

		payload := make([]byte, 8)
		binary.LittleEndian.PutUint64(payload, balances[address])

		return shim.Success(payload)
	},
	"transfer": func(args []string, stub shim.ChaincodeStubInterface) peer.Response {
		addressFrom := args[0]
		addressTo := args[1]

		if exists[addressFrom] == false {
			message := fmt.Sprint("%s doesn't exist", addressFrom)
			return shim.Error(message)
		}
		if exists[addressTo] == false {
			message := fmt.Sprint("%s doesn't exist", addressTo)
			return shim.Error(message)
		}

		amount, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}

		if amount > balances[addressFrom] {
			message := fmt.Sprint("%s doesn't have %d tokens", addressFrom, amount)
			return shim.Error(message)
		}

		balances[addressFrom] -= amount
		balances[addressTo] += amount

		return shim.Success(nil)
	},
	"transferFrom": func(args []string, stub shim.ChaincodeStubInterface) peer.Response {
		addressFrom := args[0]
		addressTo := args[1]
		addressSpender := args[2]

		if exists[addressFrom] == false {
			message := fmt.Sprint("%s doesn't exist", addressFrom)
			return shim.Error(message)
		}
		if exists[addressTo] == false {
			message := fmt.Sprint("%s doesn't exist", addressTo)
			return shim.Error(message)
		}

		amount, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}
		if allowed[addressSpender][addressFrom] < amount {
			message := fmt.Sprint("%d more than allowed amount", amount)
			return shim.Error(message)
		}

		if amount > balances[addressFrom] {
			message := fmt.Sprint("%s doesn't have %d tokens", addressFrom, amount)
			return shim.Error(message)
		}

		allowed[addressSpender][addressFrom] -= amount
		balances[addressFrom] -= amount
		balances[addressTo] += amount

		return shim.Success(nil)
	},
	"approve": func(args []string, stub shim.ChaincodeStubInterface) peer.Response {
		addressSpender := args[0]
		addressFrom := args[1]

		if exists[addressFrom] == false {
			message := fmt.Sprint("%s doesn't exist", addressFrom)
			return shim.Error(message)
		}

		amount, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return shim.Error(err.Error())
		}

		if exists[addressSpender] == false {
			exists[addressSpender] = true
			balances[addressSpender] = 0
		}

		allowed[addressSpender][addressFrom] = amount

		return shim.Success(nil)
	},
}

func (p *TokenCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	balances = make(map[string]uint64)
	exists = make(map[string]bool)
	allowed = make(map[string]map[string]uint64)

	totalSupply = 100000000000

	fmt.Println("TokenCC has been initialized")
	return shim.Success(nil)
}

func (p *TokenCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	functionName, args := stub.GetFunctionAndParameters()

	f, ok := functions[functionName]
	if !ok {
		return shim.Error("unknown function name for chaincode PersonCC")
	}

	return f(args, stub)
}

func main() {
	err := shim.Start(new(TokenCC))
	if err != nil {
		fmt.Printf("Error starting token chaincode: %s", err)
	}
}
