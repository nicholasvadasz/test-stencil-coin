package test

import (
	"Coin/pkg/blockchain"
	"testing"
	"time"
)

func TestBroadcastTransaction(t *testing.T) {
	// set up cluster
	cluster := NewCluster(3)
	chains := []*blockchain.BlockChain{cluster[0].BlockChain, cluster[1].BlockChain, cluster[2].BlockChain}
	defer CleanUp(chains)
	StartCluster(cluster)
	ConnectCluster(cluster)
	StartMiners(cluster)

	// create and broadcast transaction
	genNode := cluster[0]
	genBlock := genNode.BlockChain.LastBlock
	block := MakeBlockFromPrev(genBlock)
	tx := block.Transactions[0]
	genNode.BroadcastTransaction(tx)

	// wait for networking to occur
	time.Sleep(time.Second)

	// verify success
	CheckTransactionInTXPool(t, genNode, tx)
	CheckTransactionSeen(t, cluster, tx)
}

func TestHandleMinerBlock(t *testing.T) {
	// set up cluster
	cluster := NewCluster(3)
	chains := []*blockchain.BlockChain{cluster[0].BlockChain, cluster[1].BlockChain, cluster[2].BlockChain}
	defer CleanUp(chains)
	StartCluster(cluster)
	ConnectCluster(cluster)

	// first miner successfully mines block
	genBlock := cluster[0].BlockChain.LastBlock
	block := MakeBlockFromPrev(genBlock)
	cluster[0].HandleMinerBlock(block)

	// check that all other nodes heard about this block
	time.Sleep(time.Second)

	// make sure that all the chains are correct
	CheckMainChains(t, cluster)
}
