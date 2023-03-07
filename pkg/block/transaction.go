package block

import (
	"Coin/pkg/id"
	"Coin/pkg/pro"
	"Coin/pkg/utils"
	"crypto/sha256"
	"fmt"
	"google.golang.org/protobuf/proto"
	"strconv"
)

// TransactionInput is used as the input to create a TransactionOutput.
// Recall that TransactionInputs generate TransactionOutputs which in turn
// generate new TransactionInputs and so forth.
// ReferenceTransactionHash is the hash of the parent TransactionOutput's Transaction.
// OutputIndex is the index of the parent TransactionOutput's Transaction.
// UnlockingScript is the signature that verifies that the payer can spend the referenced TransactionOutput.
type TransactionInput struct {
	ReferenceTransactionHash string
	OutputIndex              uint32
	UnlockingScript          string
}

// TransactionOutput is an output created from a TransactionInput.
// Recall that TransactionOutputs generate TransactionInputs which in turn
// generate new TransactionOutputs and so forth.
// Amount is how much this TransactionOutput is worth.
// LockingScript is the publicKey of the payee which can be verified by the payee's signature.
type TransactionOutput struct {
	Amount        uint32
	LockingScript string
}

// Transaction contains information about a transaction.
// Version is the version of this transaction.
// Inputs is a slice of TransactionInputs.
// Outputs is a slice of TransactionOutputs.
// LockTime is the future time after which the Transaction is valid.
type Transaction struct {
	Version  uint32
	Inputs   []*TransactionInput
	Outputs  []*TransactionOutput
	LockTime uint32
}

// EncodeTransactionInput returns a pro.TransactionInput input
// given a TransactionInput.
func EncodeTransactionInput(txi *TransactionInput) *pro.TransactionInput {
	return &pro.TransactionInput{
		ReferenceTransactionHash: txi.ReferenceTransactionHash,
		OutputIndex:              txi.OutputIndex,
		UnlockingScript:          txi.UnlockingScript,
	}
}

// DecodeTransactionInput returns a TransactionInput given
// a pro.TransactionInput.
func DecodeTransactionInput(ptxi *pro.TransactionInput) *TransactionInput {
	return &TransactionInput{
		ReferenceTransactionHash: ptxi.GetReferenceTransactionHash(),
		OutputIndex:              ptxi.GetOutputIndex(),
		UnlockingScript:          ptxi.GetUnlockingScript(),
	}
}

// EncodeTransactionOutput returns a pro.TransactionOutput given
// a TransactionOutput.
func EncodeTransactionOutput(txo *TransactionOutput) *pro.TransactionOutput {
	return &pro.TransactionOutput{
		Amount:        txo.Amount,
		LockingScript: txo.LockingScript,
	}
}

// DecodeTransactionOutput returns a TransactionOutput given
// a pro.TransactionOutput.
func DecodeTransactionOutput(ptxo *pro.TransactionOutput) *TransactionOutput {
	return &TransactionOutput{
		Amount:        ptxo.GetAmount(),
		LockingScript: ptxo.GetLockingScript(),
	}
}

// EncodeTransaction returns a pro.Transaction given a Transaction.
func EncodeTransaction(tx *Transaction) *pro.Transaction {
	var ptxis []*pro.TransactionInput
	for _, txi := range tx.Inputs {
		ptxis = append(ptxis, EncodeTransactionInput(txi))
	}
	var ptxos []*pro.TransactionOutput
	for _, txo := range tx.Outputs {
		ptxos = append(ptxos, EncodeTransactionOutput(txo))
	}
	return &pro.Transaction{
		Version:  tx.Version,
		Inputs:   ptxis,
		Outputs:  ptxos,
		LockTime: tx.LockTime,
	}
}

// DecodeTransaction returns a Transaction given a pro.Transaction.
func DecodeTransaction(ptx *pro.Transaction) *Transaction {
	var txis []*TransactionInput
	for _, ptxi := range ptx.GetInputs() {
		txis = append(txis, DecodeTransactionInput(ptxi))
	}
	var txos []*TransactionOutput
	for _, ptxo := range ptx.GetOutputs() {
		txos = append(txos, DecodeTransactionOutput(ptxo))
	}
	return &Transaction{
		Version:  ptx.GetVersion(),
		Inputs:   txis,
		Outputs:  txos,
		LockTime: ptx.GetLockTime(),
	}
}

// Hash returns the hash of the transaction
func (tx *Transaction) Hash() string {
	h := sha256.New()
	pt := EncodeTransaction(tx)
	bytes, err := proto.Marshal(pt)
	if err != nil {
		fmt.Errorf("[tx.Hash()] Unable to marshal transaction")
	}
	h.Write(bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsCoinbase returns whether the
// transaction is a coinbase transaction.
// Returns:
// bool	true if the transaction is a coinbase.
// False otherwise.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 0
}

// Size returns the size of the
// underlying protobuf transaction
func (tx *Transaction) Size() uint32 {
	return pro.SizeOfTransaction(EncodeTransaction(tx))
}

// SumOutputs returns the sum of the outputs.
// Returns:
// uint32	the sum of the amounts on each
// output
func (tx *Transaction) SumOutputs() uint32 {
	var r uint32 = 0
	for _, v := range tx.Outputs {
		r += v.Amount
	}
	return r
}

// NameTag is used to log transactions while debugging
func (tx *Transaction) NameTag() string {
	i, _ := strconv.ParseInt(tx.Hash()[:10], 16, 64)
	return fmt.Sprintf("%v", utils.Colorize(fmt.Sprintf("tx-%v", tx.Hash()[:6]), int(i)))
}

// MakeSignature generates
// an unlocking script (a.k.a. signature) for the
// transaction output based on a private key.
// Inputs:
// id	id.ID	the id of the person wanting to
// unlock the particular transaction output.
// Returns:
// string	The signature represented as a hex string.
// error	Errors if the signature could not be
// produced or there was a decoding error.
func (txo *TransactionOutput) MakeSignature(id id.ID) (string, error) {
	sk := id.GetPrivateKey()
	// convert txo to bytes
	ptxo := EncodeTransactionOutput(txo)
	bytes, err := proto.Marshal(ptxo)
	if err != nil {
		fmt.Errorf("[tx.MakeSignature()] Unable to marshal transaction")
	}
	sig, err := utils.Sign(sk, bytes)
	if err != nil {
		fmt.Printf("ERROR {TransactionOutput.MakeSignature}: " +
			"The signature could not be formed.\n")
		return "", nil
	}
	return sig, nil
}
