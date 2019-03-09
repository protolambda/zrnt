package main

import (
	"fmt"
	"github.com/protolambda/go-beacon-transition/eth2/beacon"
	"github.com/protolambda/go-beacon-transition/eth2/beacon/genesis"
	"github.com/protolambda/go-beacon-transition/eth2/util/debug_json"
	"os"
)

type DataSrc func() interface{}

func main() {

	data := map[string]DataSrc{
		"genesis_state": CreateGenesisState,
		"empty_block": CreateEmptyBlock,
		// TODO: more data
	}

	for k, v := range data {
		// create the data, encode it, and write it to a file
		if err := writeDebugJson(k, v()); err != nil {
			panic(err)
		}
	}

}

func writeDebugJson(name string, data interface{}) error {
	encodedData, err := debug_json.EncodeToTreeRootJSON(data, "    ")
	if err != nil {
		fmt.Println("Failed to encode for", name)
		return err
	}
	f, err := os.Create(name+".json")
	defer f.Close()
	if err != nil {
		return err
	}
	if _, err := f.Write(encodedData); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println("encoded and written file for ", name)
	return nil
}

func CreateEmptyBlock() interface{} {
	return beacon.GetEmptyBlock()
}

func CreateGenesisState() interface{} {
	return genesis.GetGenesisBeaconState([]beacon.Deposit{}, 0, beacon.Eth1Data{})
}