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

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/db"
	"github.com/hyperledger/fabric/protos"
	"github.com/spf13/viper"
	"github.com/tecbot/gorocksdb"
)

const (
	//MaxValueSize is used to compare the size of the key-value in db.
	//If a key-value is more than this size, it's details are printed for further analysis.
	MaxValueSize = 1024 * 1024
)

type detailPrinter func(data []byte)

func main() {
	flagSetName := os.Args[0]
	flagSet := flag.NewFlagSet(flagSetName, flag.ExitOnError)
	dbDirPtr := flagSet.String("dbDir", "", "path to db dump")
	flagSet.Parse(os.Args[1:])

	dbDir := *dbDirPtr

	if dbDir == "" {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", flagSetName)
		flagSet.PrintDefaults()
		os.Exit(3)
	}
	viper.Set("peer.fileSystemPath", dbDir)
	fmt.Printf("dbDir = [%s]\n", dbDir)

	// check that dbDir exists
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "dbDir does not exist")
		os.Exit(4)
	}

	if _, err := os.Stat(dbDir + "/db"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "dbDir does not contain a sub-dir named 'db'")
		os.Exit(5)
	}

	openchainDB := db.GetDBHandle()
	defer openchainDB.Close()
	fmt.Println()
	// scan(openchainDB, "blockchainCF", openchainDB.BlockchainCF, blockDetailPrinter)
	scan(openchainDB, "blockchainCF", openchainDB.BlockchainCF, blockDetailJsonPrinter)
	fmt.Println()
	scan(openchainDB, "persistCF", openchainDB.PersistCF, nil)
	fmt.Println()
	scan(openchainDB, "stateCF", openchainDB.StateCF, nil)
	fmt.Println()
	scan(openchainDB, "stateDeltaCF", openchainDB.StateDeltaCF, nil)
	fmt.Println()
	scan(openchainDB, "IndexesCF", openchainDB.IndexesCF, nil)
	fmt.Println()
	printLiveFilesMetaData(openchainDB)
	fmt.Println()
	printProperties(openchainDB)
	fmt.Println()
}

func printLiveFilesMetaData(openchainDB *db.OpenchainDB) {
	fmt.Println("------ Details of LiveFilesMetaData ---")
	db := openchainDB.DB
	liveFileMetadata := db.GetLiveFilesMetaData()
	for _, file := range liveFileMetadata {
		fmt.Printf("file.Name=[%s], file.Level=[%d], file.Size=[%d]\n",
			file.Name, file.Level, file.Size)
	}
}

func printProperties(openchainDB *db.OpenchainDB) {
	fmt.Println("------ Details of Properties ---")
	db := openchainDB.DB
	fmt.Printf("rocksdb.estimate-live-data-size:- BlockchainCF:%s, StateCF:%s, StateDeltaCF:%s, IndexesCF:%s, PersistCF:%s\n\n",
		db.GetPropertyCF("rocksdb.estimate-live-data-size", openchainDB.BlockchainCF),
		db.GetPropertyCF("rocksdb.estimate-live-data-size", openchainDB.StateCF),
		db.GetPropertyCF("rocksdb.estimate-live-data-size", openchainDB.StateDeltaCF),
		db.GetPropertyCF("rocksdb.estimate-live-data-size", openchainDB.IndexesCF),
		db.GetPropertyCF("rocksdb.estimate-live-data-size", openchainDB.PersistCF))
	fmt.Printf("Default:%s\n", db.GetProperty("rocksdb.estimate-live-data-size"))

	fmt.Printf("rocksdb.num-live-versions:- BlockchainCF:%s, StateCF:%s, StateDeltaCF:%s, IndexesCF:%s, PersistCF:%s\n\n",
		db.GetPropertyCF("rocksdb.num-live-versions", openchainDB.BlockchainCF),
		db.GetPropertyCF("rocksdb.num-live-versions", openchainDB.StateCF),
		db.GetPropertyCF("rocksdb.num-live-versions", openchainDB.StateDeltaCF),
		db.GetPropertyCF("rocksdb.num-live-versions", openchainDB.IndexesCF),
		db.GetPropertyCF("rocksdb.num-live-versions", openchainDB.PersistCF))

	fmt.Printf("rocksdb.cfstats:\n %s %s %s %s %s\n\n",
		db.GetPropertyCF("rocksdb.cfstats", openchainDB.BlockchainCF),
		db.GetPropertyCF("rocksdb.cfstats", openchainDB.StateCF),
		db.GetPropertyCF("rocksdb.cfstats", openchainDB.StateDeltaCF),
		db.GetPropertyCF("rocksdb.cfstats", openchainDB.IndexesCF),
		db.GetPropertyCF("rocksdb.cfstats", openchainDB.PersistCF))
}

func scan(openchainDB *db.OpenchainDB, cfName string, cf *gorocksdb.ColumnFamilyHandle, printer detailPrinter) (int, int) {
	fmt.Printf("------- Printing Key-values larger than [%d] bytes in Column family [%s]--------\n", MaxValueSize, cfName)
	itr := openchainDB.GetIterator(cf)
	totalKVs := 0
	overSizeKVs := 0
	itr.SeekToFirst()
	for ; itr.Valid(); itr.Next() {
		k := itr.Key()
		v := itr.Value()
		keyBytes := k.Data()
		valueSize := v.Size()
		totalKVs++
		fmt.Printf("key=[%x], valueSize=[%d]\n", keyBytes, valueSize)
		if printer != nil {
			fmt.Println("=== KV Details === ")
			printer(v.Data())
			fmt.Println("")
		}
		if valueSize >= MaxValueSize {
			overSizeKVs++
		}
		k.Free()
		v.Free()
	}
	itr.Close()
	fmt.Printf("totalKVs=[%d], overSizeKVs=[%d]\n", totalKVs, overSizeKVs)
	return totalKVs, overSizeKVs
}

/*
func loadAllBucketNodesFromDB(openchainDB *db.OpenchainDB) {
	itr := openchainDB.GetStateCFIterator()
	defer itr.Close()
	itr.Seek([]byte{byte(0)})
	count := 0
	for ; itr.Valid(); itr.Next() {
		key := itr.Key().Data()
		if key[0] != byte(0) {
			itr.Key().Free()
			itr.Value().Free()
			break
		}

		bKey := decodeBucketKey(statemgmt.Copy(itr.Key().Data()))
		nodeBytes := statemgmt.Copy(itr.Value().Data())
		bucketNode := buckettree.unmarshalBucketNode(&bKey, nodeBytes)
		size := bKey.size() + bucketNode.size()
		itr.Key().Free()
		itr.Value().Free()
		count++
		fmt.Printf("\nbKey:[], nodeBytes[], bucketNode[]", bKey, nodeBytes, bucketNode)
	}
	fmt.Printf("Loaded buckets data. Total buckets in DB = [%d].", count)
}

func decodeBucketKey(keyBytes []byte) buckettree.bucketKey {
	level, numBytesRead := proto.DecodeVarint(keyBytes[1:])
	bucketNumber, _ := proto.DecodeVarint(keyBytes[numBytesRead+1:])
	return bucketKey{int(level), int(bucketNumber)}
}
*/

func blockDetailJsonPrinter(blockBytes []byte) {
	fmt.Printf("\n----Block data----\n")
	b := new(bytes.Buffer)
	encode := json.NewEncoder(b)
	block, err := protos.UnmarshallBlock(blockBytes)
	if err != nil {
		return
	}
	if len(blockBytes) < MaxValueSize {
		encode.Encode(block)
		b.WriteTo(os.Stdout)
	}
	trxDetailJsonPrinter(block.GetTransactions())
	nonHashJsonPrinter(block.NonHashData)
	fmt.Printf("\n----End Block----\n")
}

func trxDetailJsonPrinter(txs []*protos.Transaction) {
	b := new(bytes.Buffer)
	encode := json.NewEncoder(b)
	fmt.Printf("\n    ----Number of transactions = [%d]----\n", len(txs))
	for _, tx := range txs {
		if tx.Type == protos.Transaction_CHAINCODE_INVOKE {
			encode.Encode(tx)
			b.WriteTo(os.Stdout)
			ci := &protos.ChaincodeInvocationSpec{}
			err := proto.Unmarshal(tx.Payload, ci)
			if err != nil {
				fmt.Print("\n    Error unmarshal transaction payload", err)
			} else {
				encode.Encode(ci)
				fmt.Print("\n    --transaction payload:CHAINCODE_INVOKE")
				b.WriteTo(os.Stdout)
			}
		} else {
			cIDBytes := tx.ChaincodeID
			cID := &protos.ChaincodeID{}
			proto.Unmarshal(cIDBytes, cID)
			fmt.Print("\n    --transaction payload:CHAINCODE_DEPLOY")
			fmt.Printf("TxDetails: payloadSize=[%d], tx.Type=[%s], cID.Name=[%s], cID.Path=[%s]\n", len(tx.Payload), tx.Type, cID.Name, cID.Path)
		}
	}
	fmt.Printf("\n    ----End transactions = [%d]----\n", len(txs))
}

func nonHashJsonPrinter(nonhashdata *protos.NonHashData) {
	fmt.Print("\n--Block nonHashData")
	b := new(bytes.Buffer)
	encode := json.NewEncoder(b)
	encode.Encode(nonhashdata)
	b.WriteTo(os.Stdout)
}

func blockDetailPrinter(blockBytes []byte) {
	block, err := protos.UnmarshallBlock(blockBytes)
	if err == nil && len(blockBytes) < MaxValueSize {
		fmt.Printf("{\n")
		fmt.Printf("  version:[%d]\n", block.Version)
		fmt.Printf("  timestamp:[%s]\n", block.Timestamp.String())
		fmt.Printf("  transactions:{\n")
		trxDetailPrinter(block.GetTransactions(), "  ")
		fmt.Printf("\n  }\n")
		fmt.Printf("  stateHash:[%x]\n", block.StateHash)
		fmt.Printf("  previousBlockHash:[%x]\n", block.PreviousBlockHash)
		fmt.Printf("  consensusMetadata:[%x]\n", block.ConsensusMetadata)
		fmt.Printf("  nonHashData:{\n")
		nonHashPrinter(block.NonHashData, "  ")
		fmt.Printf("  }\n")
		fmt.Printf("}\n")
	}
}

func nonHashPrinter(nonhashdata *protos.NonHashData, ident string) {
	fmt.Printf("\n%s  %s", ident, nonhashdata.String())
}

func trxDetailPrinter(txs []*protos.Transaction, ident string) {
	fmt.Printf("\n%s----Number of transactions = [%d]----\n", ident, len(txs))
	for _, tx := range txs {
		cIDBytes := tx.ChaincodeID
		cID := &protos.ChaincodeID{}
		proto.Unmarshal(cIDBytes, cID)
		fmt.Printf("\n%s{", ident)
		fmt.Printf("\n%s  type:[%s]", ident, tx.Type)
		fmt.Printf("\n%s  chaincodeID:Name=[%s],Path=[%s]", ident, cID.Name, cID.Path)
		fmt.Printf("\n%s  payload:size[%d]{", ident, len(tx.Payload))

		if tx.Type == protos.Transaction_CHAINCODE_INVOKE {
			ci := &protos.ChaincodeInvocationSpec{}
			err := proto.Unmarshal(tx.Payload, ci)
			if err != nil {
				fmt.Print("\n  Error unmarshal", err)
			} else {
				fmt.Printf("\n%s    [%s] \n", ident, ci.String())
			}
		} else {
			fmt.Printf("\n%s   deploy code ignoe", ident)
		}

		fmt.Printf("\n%s  }", ident)
		fmt.Printf("\n%s  metadata:[%x]", ident, tx.Metadata)
		fmt.Printf("\n%s  uuid:[%s]", ident, tx.Uuid)
		fmt.Printf("\n%s  timestamp:[%s]", ident, tx.Timestamp)
		fmt.Printf("\n%s  confidentialityLevel:[%s]", ident, tx.ConfidentialityLevel)
		fmt.Printf("\n%s  confidentialityProtocolVersion:[%s]", ident, tx.ConfidentialityProtocolVersion)
		fmt.Printf("\n%s  nonce:[%x]", ident, tx.Nonce)
		fmt.Printf("\n%s  toValidators:[%x]", ident, tx.ToValidators)
		fmt.Printf("\n%s  cert:[%x]", ident, tx.Cert)
		fmt.Printf("\n%s  signature:[%x]", ident, tx.Signature)
		// fmt.Printf("  --TxDetails: payloadSize=[%d], tx.Type=[%s], cID.Name=[%s], cID.Path=[%s]\n", len(tx.Payload), tx.Type, cID.Name, cID.Path)
		fmt.Printf("\n%s}", ident)
		fmt.Printf("\n%s----next transaction ----\n", ident)
	}
	fmt.Printf("\n%s----end trx----\n", ident)

}
