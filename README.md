## Wallet RPC Client

The ```go-monero/walletrpc``` package is a RPC client with all the methods of the v0.11.0.0 release.
It does support digest authentication, [however I don't recommend using it alone (without https).](https://en.wikipedia.org/wiki/Digest_access_authentication#Disadvantages) If there is a need to split the RPC client and server into separate instances, you could put a proxy on the instance that contains the RPC server and check the authenticity of the requests using https + X-API-KEY headers between the proxy and this RPC client (there is an example about this implementation below)

### Usage

The simplest way to use walletrpc is if you have both the server (monero-wallet-rpc) and the client on the same machine.

Go:

```Go
package main

import (
	"fmt"
	"os"

	"github.com/ibclabs/go-monero/walletrpc"
)

func main() {
	// Start a wallet client instance
	client := walletrpc.New(walletrpc.Config{
		Address: "http://127.0.0.1:18082/json_rpc",
	})

	// check wallet balance
	balance, unlocked, err := client.GetBalance()

	// there are two types of error that can happen:
	//   connection errors
	//   monero wallet errors
	// connection errors are pretty much unicorns if everything is on the
	// same instance (unless your OS hit an open files limit or something)
	if err != nil {
		if iswerr, werr := walletrpc.GetWalletError(err); iswerr {
			// it is a monero wallet error
			fmt.Printf("Wallet error (id:%v) %v\n", werr.Code, werr.Message)
			os.Exit(1)
		}
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Balance:", walletrpc.XMRToDecimal(balance))
	fmt.Println("Unlocked balance:", walletrpc.XMRToDecimal(unlocked))

	// Make a transfer
	res, err := client.Transfer(walletrpc.TransferRequest{
		Destinations: []walletrpc.Destination{
			{
				Address: "45eoXYNHC4LcL2Hh42T9FMPTmZHyDEwDbgfBEuNj3RZUek8A4og4KiCfVL6ZmvHBfCALnggWtHH7QHF8426yRayLQq7MLf5",
				Amount:  10000000000, // 0.01 XMR
			},
		},
		Priority: walletrpc.PriorityUnimportant,
		Mixin:    1,
	})
	if err != nil {
		if iswerr, werr := walletrpc.GetWalletError(err); iswerr {
			// insufficient funds return a monero wallet error
			// walletrpc.ErrGenericTransferError
			fmt.Printf("Wallet error (id:%v) %v\n", werr.Code, werr.Message)
			os.Exit(1)
		}
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Transfer success! Fee:", walletrpc.XMRToDecimal(res.Fee), "Hash:", res.TxHash)
}
```