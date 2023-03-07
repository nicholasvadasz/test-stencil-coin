package test

import (
	"Coin/pkg/block"
	"testing"
)

func TestBasicHandleBlock(t *testing.T) {
	w := CreateMockedWallet()
	b := MockedBlock()
	w.HandleBlock(b.Transactions)
}

func TestHandleBlock(t *testing.T) {
	// create a wallet with the genesis coin
	w := CreateMockedGenesisWallet()
	genCoin := GenesisBlock().Transactions[0].Outputs[0]
	n := uint32(10)
	amt := uint32(100)
	FillWalletWithCoins(w, n, amt)
	// Our wallet should now have 11 coins, since we just filled it with 10 additional coins (we already had genesis)
	if len(w.CoinCollection) != 11 {
		t.Errorf("Expected Number of coins: %v\n Actual number of coins: %v", 11, len(w.CoinCollection))
	}
	balance := genCoin.Amount + n*amt
	AssertBalance(t, w, balance)
	amt2 := uint32(20)
	b := MockedBlockWithNCoins(w, 10, amt2)
	// our unconfirmed received coins should be empty to start
	AssertSize(t, len(w.UnconfirmedReceivedCoins), 0)
	w.HandleBlock(b.Transactions)
	// Check the sizes of our mappings
	AssertSize(t, len(w.CoinCollection), 11)
	AssertSize(t, len(w.UnconfirmedReceivedCoins), 10)
	// handle enough blocks to update receiving mappings
	for i := uint32(0); i < w.Config.SafeBlockAmount; i++ {
		w.HandleBlock(MockedBlock().Transactions)
	}
	// Check the sizes of our mappings
	AssertSize(t, len(w.CoinCollection), 21)
	AssertSize(t, len(w.UnconfirmedReceivedCoins), 0)
	updatedBalance := balance + n*amt2
	AssertBalance(t, w, updatedBalance)
}

func TestBasicTransactionRequest(t *testing.T) {
	w := CreateMockedGenesisWallet()
	genCoin := GenesisBlock().Transactions[0].Outputs[0]
	if len(w.UnconfirmedReceivedCoins) != 1 {
		t.Errorf("Error: should have added coin to unconfirmed received coins")
	}
	fee := uint32(20)
	// should not be able to spend the coin right away
	tx := w.RequestTransaction(50, fee, []byte("hello"))
	if tx != nil {
		t.Errorf("Wallet should not have been able to create transaction")
	}
	// handle enough blocks so that we can spend the coin
	for i := uint32(0); i < w.Config.SafeBlockAmount; i++ {
		w.HandleBlock(MockedBlock().Transactions)
	}
	AssertBalance(t, w, genCoin.Amount)
	// now requesting the transaction SHOULD work
	tx = w.RequestTransaction(50, fee, []byte("hello"))
	if tx == nil {
		t.Errorf("Wallet was unable to create transaction")
		return
	}
	// since we only have one coin, the sum of the outputs should equal that
	// of the genesis block
	if tx.SumOutputs() != GenesisBlock().Transactions[0].SumOutputs()-fee {
		t.Errorf("Expected sum: %v\n Actual sum %v", 50, tx.SumOutputs())
	}
}

func TestMultipleTransactionRequests(t *testing.T) {
	w := CreateMockedWallet()
	n := uint32(10)
	amt := uint32(100)
	FillWalletWithCoins(w, n, amt)
	AssertBalance(t, w, 10*100)
	sendingAmt := uint32(90)
	fee := uint32(5)
	var txs []*block.Transaction
	for i := uint32(0); i < n; i++ {
		tx := w.RequestTransaction(sendingAmt, fee, []byte("hello"))
		txs = append(txs, tx)
	}
	AssertBalance(t, w, 0)
	b := MockedBlock()
	b.Transactions = txs
	w.HandleBlock(b.Transactions)
	for i := uint32(0); i < w.Config.SafeBlockAmount; i++ {
		w.HandleBlock(MockedBlock().Transactions)
	}
	change := amt - (sendingAmt + fee)
	AssertBalance(t, w, change*n)
}

func TestRequestTransactionFails(t *testing.T) {
	w := CreateMockedWallet()
	n := uint32(2)
	amt := uint32(50)
	FillWalletWithCoins(w, n, amt)
	AssertBalance(t, w, n*amt)
	publicKey := []byte("hello")

	// First fail: amount exceeds balance
	tx := w.RequestTransaction(2*amt+1, 0, publicKey)
	if tx != nil {
		t.Errorf("Request should have failed due to insufficient balance")
	}
	AssertBalance(t, w, n*amt)

	// Send a valid request using the 2 coins that we have
	w.RequestTransaction(2*amt-20, 0, publicKey)
	AssertSize(t, len(w.CoinCollection), 0)
	AssertBalance(t, w, 0)

	// Second fail: don't have any coins to send the transaction, and balance should temporarily be 0
	tx = w.RequestTransaction(20, 0, publicKey)
	if tx != nil {
		t.Errorf("Request should have failed due to lack of available coins and balance")
	}
}
