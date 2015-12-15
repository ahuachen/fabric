/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package buckettree

import (
	"testing"

	"github.com/openblockchain/obc-peer/openchain/ledger/statemgmt"
	"github.com/openblockchain/obc-peer/openchain/ledger/testutil"
)

func TestStateImpl_ComputeHash_AllInMemory_NoContents(t *testing.T) {
	testDBWrapper.CreateFreshDB(t)
	stateImplTestWrapper := newStateImplTestWrapper(t)
	hash := stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(statemgmt.NewStateDelta())
	testutil.AssertEquals(t, hash, nil)
}

func TestStateImpl_ComputeHash_AllInMemory_1(t *testing.T) {
	// number of buckets at each level 26,9,3,1
	testHasher, stateImplTestWrapper, stateDelta := createFreshDBAndInitTestStateImplWithCustomHasher(t, 26, 3)
	testHasher.populate("chaincodeID1", "key1", 0)
	testHasher.populate("chaincodeID2", "key2", 0)
	testHasher.populate("chaincodeID3", "key3", 0)
	testHasher.populate("chaincodeID4", "key4", 3)

	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID2", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID3", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID4", "key4", []byte("value4"))

	rootHash := stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)

	expectedHashBucket3_1 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID1", "key1"), []byte("value1"),
		statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"),
		statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3"))
	expectedHashBucket3_4 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID4", "key4"), []byte("value4"))

	expectedHash := testutil.ComputeCryptoHash(expectedHashBucket3_1, expectedHashBucket3_4)
	testutil.AssertEquals(t, rootHash, expectedHash)
}

func TestStateImpl_ComputeHash_AllInMemory_2(t *testing.T) {
	// number of buckets at each level 26,13,7,4,2,1
	testHasher, stateImplTestWrapper, stateDelta := createFreshDBAndInitTestStateImplWithCustomHasher(t, 26, 2)
	// first two buckets - meet at next level
	testHasher.populate("chaincodeID1", "key1", 0)
	testHasher.populate("chaincodeID2", "key2", 1)

	// middle two buckets
	testHasher.populate("chaincodeID3", "key3", 5)
	testHasher.populate("chaincodeID4", "key4", 9)

	// last two buckets - meet at next level
	testHasher.populate("chaincodeID5", "key5", 24)
	testHasher.populate("chaincodeID6", "key6", 25)

	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID2", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID3", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID4", "key4", []byte("value4"))
	stateDelta.Set("chaincodeID5", "key5", []byte("value5"))
	stateDelta.Set("chaincodeID6", "key6", []byte("value6"))

	rootHash := stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)

	expectedHashBucket5_1 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID1", "key1"), []byte("value1"))
	expectedHashBucket5_2 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"))
	expectedHashBucket5_6 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3"))
	expectedHashBucket5_10 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID4", "key4"), []byte("value4"))
	expectedHashBucket5_25 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID5", "key5"), []byte("value5"))
	expectedHashBucket5_26 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID6", "key6"), []byte("value6"))

	expectedHashBucket4_1 := testutil.ComputeCryptoHash(expectedHashBucket5_1, expectedHashBucket5_2)
	expectedHashBucket4_13 := testutil.ComputeCryptoHash(expectedHashBucket5_25, expectedHashBucket5_26)

	expectedHashBucket2_1 := testutil.ComputeCryptoHash(expectedHashBucket4_1, expectedHashBucket5_6)

	expectedHashBucket1_1 := testutil.ComputeCryptoHash(expectedHashBucket2_1, expectedHashBucket5_10)

	expectedHash := testutil.ComputeCryptoHash(expectedHashBucket1_1, expectedHashBucket4_13)
	testutil.AssertEquals(t, rootHash, expectedHash)
}

func TestStateImpl_ComputeHash_DB_1(t *testing.T) {
	// number of buckets at each level 26,9,3,1
	testHasher, stateImplTestWrapper, stateDelta := createFreshDBAndInitTestStateImplWithCustomHasher(t, 26, 3)
	// populate hash fucntion such that
	// all keys belong to a single bucket so as to test overwrite/delete scenario
	testHasher.populate("chaincodeID1", "key1", 3)
	testHasher.populate("chaincodeID2", "key2", 3)
	testHasher.populate("chaincodeID3", "key3", 3)
	testHasher.populate("chaincodeID4", "key4", 3)
	testHasher.populate("chaincodeID5", "key5", 3)
	testHasher.populate("chaincodeID6", "key6", 3)
	testHasher.populate("chaincodeID7", "key7", 3)

	stateDelta.Set("chaincodeID2", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID3", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID5", "key5", []byte("value5"))
	stateDelta.Set("chaincodeID6", "key6", []byte("value6"))
	rootHash := stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()

	expectedHash1 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"),
		statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3"),
		statemgmt.ConstructCompositeKey("chaincodeID5", "key5"), []byte("value5"),
		statemgmt.ConstructCompositeKey("chaincodeID6", "key6"), []byte("value6"))
	testutil.AssertEquals(t, rootHash, expectedHash1)

	// modify boundary keys and a middle key
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID2", "key2", []byte("value2_new"))
	stateDelta.Set("chaincodeID3", "key3", []byte("value3_new"))
	stateDelta.Set("chaincodeID6", "key6", []byte("value6_new"))
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	expectedHash2 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2_new"),
		statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3_new"),
		statemgmt.ConstructCompositeKey("chaincodeID5", "key5"), []byte("value5"),
		statemgmt.ConstructCompositeKey("chaincodeID6", "key6"), []byte("value6_new"))
	testutil.AssertEquals(t, rootHash, expectedHash2)
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID2", "key2"), []byte("value2_new"))
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID3", "key3"), []byte("value3_new"))
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID6", "key6"), []byte("value6_new"))

	// insert keys at boundary and in the middle
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID4", "key4", []byte("value4"))
	stateDelta.Set("chaincodeID7", "key7", []byte("value7"))
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	expectedHash3 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID1", "key1"), []byte("value1"),
		statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2_new"),
		statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3_new"),
		statemgmt.ConstructCompositeKey("chaincodeID4", "key4"), []byte("value4"),
		statemgmt.ConstructCompositeKey("chaincodeID5", "key5"), []byte("value5"),
		statemgmt.ConstructCompositeKey("chaincodeID6", "key6"), []byte("value6_new"),
		statemgmt.ConstructCompositeKey("chaincodeID7", "key7"), []byte("value7"))
	testutil.AssertEquals(t, rootHash, expectedHash3)
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID1", "key1"), []byte("value1"))
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID4", "key4"), []byte("value4"))
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID7", "key7"), []byte("value7"))

	// delete keys at a boundary and in the middle
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("chaincodeID1", "key1")
	stateDelta.Delete("chaincodeID4", "key4")
	stateDelta.Delete("chaincodeID7", "key7")
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	testutil.AssertEquals(t, rootHash, expectedHash2)
	testutil.AssertNil(t, stateImplTestWrapper.get("chaincodeID1", "key1"))
	testutil.AssertNil(t, stateImplTestWrapper.get("chaincodeID4", "key4"))
	testutil.AssertNil(t, stateImplTestWrapper.get("chaincodeID7", "key7"))
}

func TestStateImpl_ComputeHash_DB_2(t *testing.T) {
	// number of buckets at each level 26,13,7,4,2,1
	testHasher, stateImplTestWrapper, stateDelta := createFreshDBAndInitTestStateImplWithCustomHasher(t, 26, 2)
	testHasher.populate("chaincodeID1", "key1", 0)
	testHasher.populate("chaincodeID2", "key2", 1)
	testHasher.populate("chaincodeID3", "key3", 5)
	testHasher.populate("chaincodeID4", "key4", 9)
	testHasher.populate("chaincodeID5", "key5", 24)
	testHasher.populate("chaincodeID6", "key6", 25)

	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID2", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID3", "key3", []byte("value3"))
	stateDelta.Set("chaincodeID4", "key4", []byte("value4"))
	stateDelta.Set("chaincodeID5", "key5", []byte("value5"))
	stateDelta.Set("chaincodeID6", "key6", []byte("value6"))
	stateImplTestWrapper.prepareWorkingSet(stateDelta)
	// Populate DB
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()

	//////////// Test - constrcuting a new state tree simulates starting state tree when db already has some data ////////
	stateImplTestWrapper.constructNewStateImpl()
	rootHash := stateImplTestWrapper.computeCryptoHash()

	/*************************** bucket-tree-structure ***************
	1		1		1	1	1	1
	2		1		1	1	1	1
	6		3		2	1	1	1
	10	5		3	2	1	1
	25	13	7	4	2	1
	26	13	7	4	2	1
	*******************************************************************/
	expectedHashBucket5_1 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID1", "key1"), []byte("value1"))
	expectedHashBucket5_2 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"))
	expectedHashBucket5_6 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID3", "key3"), []byte("value3"))
	expectedHashBucket5_10 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID4", "key4"), []byte("value4"))
	expectedHashBucket5_25 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID5", "key5"), []byte("value5"))
	expectedHashBucket5_26 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID6", "key6"), []byte("value6"))
	expectedHashBucket4_1 := testutil.ComputeCryptoHash(expectedHashBucket5_1, expectedHashBucket5_2)
	expectedHashBucket4_13 := testutil.ComputeCryptoHash(expectedHashBucket5_25, expectedHashBucket5_26)
	expectedHashBucket2_1 := testutil.ComputeCryptoHash(expectedHashBucket4_1, expectedHashBucket5_6)
	expectedHashBucket1_1 := testutil.ComputeCryptoHash(expectedHashBucket2_1, expectedHashBucket5_10)
	expectedHash := testutil.ComputeCryptoHash(expectedHashBucket1_1, expectedHashBucket4_13)
	testutil.AssertEquals(t, rootHash, expectedHash)

	//////////////	Test - Add a few more keys (include keys in the existing buckes and new buckets) /////////////////////
	stateDelta = statemgmt.NewStateDelta()
	testHasher.populate("chaincodeID7", "key7", 1)
	testHasher.populate("chaincodeID8", "key8", 7)
	testHasher.populate("chaincodeID9", "key9", 9)
	testHasher.populate("chaincodeID10", "key10", 20)

	stateDelta.Set("chaincodeID7", "key7", []byte("value7"))
	stateDelta.Set("chaincodeID8", "key8", []byte("value8"))
	stateDelta.Set("chaincodeID9", "key9", []byte("value9"))
	stateDelta.Set("chaincodeID10", "key10", []byte("value10"))

	/*************************** bucket-tree-structure after adding keys ***************
	1		1		1	1	1	1
	2		1		1	1	1	1
	6		3		2	1	1	1
	8		4		2	1	1	1
	10	5		3	2	1	1
	21	11	6	3	2	1
	25	13	7	4	2	1
	26	13	7	4	2	1
	***********************************************************************************/
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedHashBucket5_2 = testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"),
		statemgmt.ConstructCompositeKey("chaincodeID7", "key7"), []byte("value7"))
	expectedHashBucket5_8 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID8", "key8"), []byte("value8"))
	expectedHashBucket5_10 = testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID4", "key4"), []byte("value4"),
		statemgmt.ConstructCompositeKey("chaincodeID9", "key9"), []byte("value9"))
	expectedHashBucket5_21 := testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID10", "key10"), []byte("value10"))
	expectedHashBucket4_1 = testutil.ComputeCryptoHash(expectedHashBucket5_1, expectedHashBucket5_2)
	expectedHashBucket3_2 := testutil.ComputeCryptoHash(expectedHashBucket5_6, expectedHashBucket5_8)
	expectedHashBucket2_1 = testutil.ComputeCryptoHash(expectedHashBucket4_1, expectedHashBucket3_2)

	expectedHashBucket1_1 = testutil.ComputeCryptoHash(expectedHashBucket2_1, expectedHashBucket5_10)
	expectedHashBucket1_2 := testutil.ComputeCryptoHash(expectedHashBucket5_21, expectedHashBucket4_13)
	expectedHash = testutil.ComputeCryptoHash(expectedHashBucket1_1, expectedHashBucket1_2)
	testutil.AssertEquals(t, rootHash, expectedHash)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()

	//////////////	Test - overwrite an existing key /////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Set("chaincodeID7", "key7", []byte("value7_new"))
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedHashBucket5_2 = testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key2"), []byte("value2"),
		statemgmt.ConstructCompositeKey("chaincodeID7", "key7"), []byte("value7_new"))
	expectedHashBucket4_1 = testutil.ComputeCryptoHash(expectedHashBucket5_1, expectedHashBucket5_2)
	expectedHashBucket2_1 = testutil.ComputeCryptoHash(expectedHashBucket4_1, expectedHashBucket3_2)
	expectedHashBucket1_1 = testutil.ComputeCryptoHash(expectedHashBucket2_1, expectedHashBucket5_10)
	expectedHash = testutil.ComputeCryptoHash(expectedHashBucket1_1, expectedHashBucket1_2)
	testutil.AssertEquals(t, rootHash, expectedHash)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID7", "key7"), []byte("value7_new"))

	//////////////	Test - delete an existing key /////////////////////
	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("chaincodeID2", "key2")
	rootHash = stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	expectedHashBucket5_2 = testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID7", "key7"), []byte("value7_new"))
	expectedHashBucket4_1 = testutil.ComputeCryptoHash(expectedHashBucket5_1, expectedHashBucket5_2)
	expectedHashBucket2_1 = testutil.ComputeCryptoHash(expectedHashBucket4_1, expectedHashBucket3_2)
	expectedHashBucket1_1 = testutil.ComputeCryptoHash(expectedHashBucket2_1, expectedHashBucket5_10)
	expectedHash = testutil.ComputeCryptoHash(expectedHashBucket1_1, expectedHashBucket1_2)
	testutil.AssertEquals(t, rootHash, expectedHash)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	testutil.AssertNil(t, stateImplTestWrapper.get("chaincodeID2", "key2"))
}

func TestStateImpl_ComputeHash_DB_3(t *testing.T) {
	// simple test... not using custom hasher
	conf = initConfig(DefaultNumBuckets, DefaultMaxGroupingAtEachLevel, fnvHash)
	testDBWrapper.CreateFreshDB(t)
	stateImplTestWrapper := newStateImplTestWrapper(t)
	stateImpl := stateImplTestWrapper.stateImpl
	stateDelta := statemgmt.NewStateDelta()
	stateDelta.Set("chaincode1", "key1", []byte("value1"))
	stateDelta.Set("chaincode2", "key2", []byte("value2"))
	stateDelta.Set("chaincode3", "key3", []byte("value3"))
	stateImpl.PrepareWorkingSet(stateDelta)
	hash1 := stateImplTestWrapper.computeCryptoHash()
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()

	stateDelta = statemgmt.NewStateDelta()
	stateDelta.Delete("chaincode1", "key1")
	stateDelta.Delete("chaincode2", "key2")
	stateDelta.Delete("chaincode3", "key3")
	stateImpl.PrepareWorkingSet(stateDelta)
	hash2 := stateImplTestWrapper.computeCryptoHash()
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()
	testutil.AssertNotEquals(t, hash1, hash2)
	testutil.AssertNil(t, hash2)
}

func TestStateImpl_DB_Changes(t *testing.T) {
	// number of buckets at each level 26,9,3,1
	testHasher, stateImplTestWrapper, stateDelta := createFreshDBAndInitTestStateImplWithCustomHasher(t, 26, 3)
	// populate hash fucntion such that
	// ["chaincodeID1", "key1"] is bucketized to bucket 1
	testHasher.populate("chaincodeID1", "key1", 0)
	testHasher.populate("chaincodeID1", "key2", 0)
	testHasher.populate("chaincodeID2", "key1", 1)
	testHasher.populate("chaincodeID2", "key3", 3)
	testHasher.populate("chaincodeID10", "key10", 24)

	// prepare stateDelta
	stateDelta.Set("chaincodeID1", "key1", []byte("value1"))
	stateDelta.Set("chaincodeID1", "key2", []byte("value2"))
	stateDelta.Set("chaincodeID2", "key1", []byte("value3"))
	stateDelta.Set("chaincodeID2", "key3", []byte("value4"))

	stateImplTestWrapper.prepareWorkingSetAndComputeCryptoHash(stateDelta)
	stateImplTestWrapper.persistChangesAndResetInMemoryChanges()

	// Read state from DB
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID1", "key1"), []byte("value1"))
	testutil.AssertEquals(t, stateImplTestWrapper.get("chaincodeID2", "key1"), []byte("value3"))

	// fetch datanode from DB
	dataNodeFromDB, _ := fetchDataNodeFromDB(newDataKey("chaincodeID2", "key1"))
	testutil.AssertEquals(t, dataNodeFromDB, newDataNode(newDataKey("chaincodeID2", "key1"), []byte("value3")))

	//fetch non-existing data node from DB
	dataNodeFromDB, _ = fetchDataNodeFromDB(newDataKey("chaincodeID10", "key10"))
	t.Logf("isNIL...[%t]", dataNodeFromDB == nil)
	testutil.AssertNil(t, dataNodeFromDB)

	// fetch all data nodes from db that belong to bucket 1 at lowest level
	dataNodesFromDB, _ := fetchDataNodesFromDBFor(newBucketKeyAtLowestLevel(1))
	testutil.AssertContainsAll(t, dataNodesFromDB,
		dataNodes{newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value1")),
			newDataNode(newDataKey("chaincodeID1", "key2"), []byte("value2"))})

	// fetch all data nodes from db that belong to bucket 2 at lowest level
	dataNodesFromDB, _ = fetchDataNodesFromDBFor(newBucketKeyAtLowestLevel(2))
	testutil.AssertContainsAll(t, dataNodesFromDB,
		dataNodes{newDataNode(newDataKey("chaincodeID2", "key1"), []byte("value3"))})

	// fetch first bucket at second level
	bucketNodeFromDB, _ := fetchBucketNodeFromDB(newBucketKey(2, 1))
	testutil.AssertEquals(t, bucketNodeFromDB.bucketKey, newBucketKey(2, 1))
	//check childrenCryptoHash entries in the bucket node from DB
	testutil.AssertEquals(t, bucketNodeFromDB.childrenCryptoHash[0],
		testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID1", "key1"), []byte("value1"),
			statemgmt.ConstructCompositeKey("chaincodeID1", "key2"), []byte("value2")))

	testutil.AssertEquals(t, bucketNodeFromDB.childrenCryptoHash[1],
		testutil.ComputeCryptoHash(statemgmt.ConstructCompositeKey("chaincodeID2", "key1"), []byte("value3")))

	testutil.AssertNil(t, bucketNodeFromDB.childrenCryptoHash[2])

	// third bucket at second level should be nil
	bucketNodeFromDB, _ = fetchBucketNodeFromDB(newBucketKey(2, 3))
	testutil.AssertNil(t, bucketNodeFromDB)
}
