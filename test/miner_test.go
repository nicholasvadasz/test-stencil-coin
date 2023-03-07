package test

import (
	"Coin/pkg/blockchain"
	"bytes"
	"context"
	"testing"
	"time"
)

func TestGenerateCoinbaseTransaction(t *testing.T) {
	// Set up node
	n := NewGenesisNode()
	defer CleanUp([]*blockchain.BlockChain{n.BlockChain})
	n.Start()

	// create transactions and store them
	txs := GenerateTransactions(nil)
	n.BlockChain.CoinDB.StoreBlock(txs)

	// build new transactions off of txs
	newTransactions := GenerateTransactions(txs)
	coinbase := n.Miner.GenerateCoinbaseTransaction(newTransactions)
	if coinbase.Inputs == nil {
		t.Errorf("coinbase inputs should be an empty slice, not nil")
	}
	AssertSize(t, 1, len(coinbase.Outputs))
	if coinbase.Outputs[0].Amount != 150 {
		t.Errorf("output amount is incorrect, Actual amount:%v\n", coinbase.Outputs[0].Amount)
	}
}

func TestCalculateNonce(t *testing.T) {
	// Set up node
	n := NewGenesisNode()
	defer CleanUp([]*blockchain.BlockChain{n.BlockChain})
	n.Start()

	difficulty := CreateDifficultyTarget(3)
	n.Miner.DifficultyTarget = difficulty

	b := MockedBlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	success := n.Miner.CalculateNonce(ctx, b)
	if !success {
		t.Errorf("Calculating the nonce should have succeeded!")
	}
	if bytes.Compare([]byte(b.Hash()), difficulty) != -1 {
		t.Errorf("Hash must be less than difficulty")
	}
}

func TestCalculateNonceContext(t *testing.T) {
	// Set up node
	n := NewGenesisNode()
	defer CleanUp([]*blockchain.BlockChain{n.BlockChain})
	n.Start()

	difficulty := CreateHardestDifficultyTarget()
	n.Miner.DifficultyTarget = difficulty

	b := MockedBlock()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	success := n.Miner.CalculateNonce(ctx, b)
	if success {
		t.Errorf("Calculating the nonce should NOT have succeeded!")
	}
	if bytes.Compare([]byte(b.Hash()), difficulty) == -1 {
		t.Errorf("Hash should not be less than difficulty!")
	}
}

func TestMine(t *testing.T) {
	// fill the coin DB with coins
	n := NewGenesisNode()
	defer CleanUp([]*blockchain.BlockChain{n.BlockChain})
	n.Start()

	// create transactions and store them
	txs := GenerateTransactions(nil)
	n.BlockChain.CoinDB.StoreBlock(txs)
	newTxs := GenerateTransactions(txs)
	for _, tx := range newTxs {
		n.Miner.HandleTransaction(tx)
	}
	AssertSize(t, int(n.Miner.TxPool.Count.Load()), len(newTxs))

	// make difficulty target easy so that miner instantly finds a block
	difficulty := CreateDifficultyTarget(2)
	n.Miner.DifficultyTarget = difficulty

	// send the miner off to the races
	b := n.Miner.Mine()
	if b == nil {
		t.Errorf("block should not be nil")
	}
	// give node time to handle the miner block
	time.Sleep(2 * time.Second)

	AssertSize(t, int(n.BlockChain.Length), 2)
	if n.BlockChain.LastHash != b.Hash() {
		t.Errorf("Blockchain last hash should be equal to that of the mined block")
	}
	minedBlock := n.BlockChain.LastBlock
	if bytes.Compare([]byte(minedBlock.Hash()), n.Miner.DifficultyTarget) != -1 {
		t.Errorf("Hash should be less than difficulty!")
	}
	// coinbase needs to be in there
	AssertSize(t, len(minedBlock.Transactions), len(newTxs)+1)
	// coinbase should not have inputs
	coinbase := minedBlock.Transactions[0]
	AssertSize(t, len(coinbase.Inputs), 0)
	if coinbase.Outputs[0].Amount < n.Miner.CalculateMintingReward() {
		t.Errorf("Minting reward should have been greater")
	}
	AssertSize(t, int(n.Miner.TxPool.Count.Load()), 0)
	if n.Miner.Mining.Load() != false {
		t.Errorf("Miner shouldn't be mining anymore, since there is nothing to mine.")
	}
}
