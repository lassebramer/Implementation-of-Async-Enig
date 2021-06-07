package blockchain

import (
	"crypto/rsa"
	"strconv"
	"sync"
	"time"

	"github.com/Hvassaa/bachelor/network"
	"github.com/Hvassaa/bachelor/signature"
	"github.com/Hvassaa/bachelor/types"
)

type Blockchain interface {
	GetBlockTree() *BlockTree
	HandleTransaction(transaction types.SignedTransaction)
	HandleBlock(block *types.Block)
	GetTransactions() []types.SignedTransaction
	GetLedger() Ledger
	RegisterClient(client *network.ClientInterface)
	GetPublickey() string
	MakeBlock()
	MakeTransaction(to rsa.PublicKey, amount int)
	GetHardness() int
	StartLottery()
	GetGenesisBlockHash() types.Hash
	GetNumberOfPeers() int
	SetFinalizer(f *Finalizer)
	GetFinalizer() *Finalizer
	ReadTree(filename string)
	WriteTree()
	GetLastFinal() *types.Block
}

type Protocol struct {
	slot       int
	blockTree  BlockTree
	privateKey rsa.PrivateKey
	publicKeys []rsa.PublicKey
	// Transactions not seen in a block, map from id to transaction
	queuedTransactions     map[string]types.SignedTransaction
	queuedTransactionsLock sync.RWMutex
	waitingBlocks          map[types.Hash]*types.Block
	waitingBlockLock       sync.RWMutex
	client                 *network.ClientInterface
	transactionId          int
	InitialStakeHolders    []string
	lottery                *Lottery
	hardness               int
	gBlock                 types.Hash
	finalizer              *Finalizer
	handleBlockChannel     chan *types.Block
	UseFinalizer           bool
}

func (bc *Protocol) GetLastFinal() *types.Block {
	return bc.blockTree.GetLastFinal()
}

func MakeProtocol(privateKey rsa.PrivateKey, publicKeys []rsa.PublicKey, hardness int) Blockchain {
	bc := new(Protocol)
	bc.UseFinalizer = false
	bc.slot = 1
	InitialStakeHolders := make([]string, 0, len(publicKeys))
	for _, pk := range publicKeys {
		pkString := signature.PublicKeyToString(pk)
		InitialStakeHolders = append(InitialStakeHolders, pkString)
	}
	gBlock := types.Block{}
	gBlock.Slot = 0
	gBlock.Height = 0
	tree := MakeBlockTree(InitialStakeHolders, gBlock)
	bc.gBlock = gBlock.Hash()
	bc.blockTree = tree
	bc.privateKey = privateKey
	bc.publicKeys = publicKeys
	bc.queuedTransactions = make(map[string]types.SignedTransaction)
	//transaction id should be fixed length
	bc.transactionId = 1000000
	bc.hardness = hardness
	// TODO maybe this should be a parameter
	lottery := MakeFakeLottery(bc)
	bc.lottery = &lottery
	bc.waitingBlocks = make(map[types.Hash]*types.Block)
	bc.handleBlockChannel = make(chan *types.Block)
	go bc.ActuallyHandleBlock()
	return bc
}

func (bc *Protocol) ReadTree(filename string) {
	bc.blockTree.ReadTree(filename)
}

func (bc *Protocol) WriteTree() {
	bc.blockTree.WriteTree()
}

func (bc *Protocol) SetFinalizer(f *Finalizer) {
	bc.finalizer = f
	bc.UseFinalizer = true
}

func (bc *Protocol) GetFinalizer() *Finalizer {
	return bc.finalizer
}

func (bc *Protocol) GetBlockTree() *BlockTree {
	return &bc.blockTree
}

func (bc *Protocol) GetGenesisBlockHash() types.Hash {
	return bc.gBlock
}

func (bc *Protocol) HandleTransaction(transaction types.SignedTransaction) {
	bc.queuedTransactionsLock.Lock()
	defer bc.queuedTransactionsLock.Unlock()
	if _, exists := bc.queuedTransactions[transaction.ID]; !exists {
		bc.queuedTransactions[transaction.ID] = transaction
		(*bc.client).Send(transaction)
	}

}

func (bc *Protocol) HandleBlock(block *types.Block) {
	// fmt.Println("got block")
	go func() {
		bc.handleBlockChannel <- block
	}()

}

func (bc *Protocol) ActuallyHandleBlock() {
	for {
		block := <-bc.handleBlockChannel

		receivedBlockHash := block.Hash()

		// check that we do not have it
		blockInTree := bc.blockTree.GetBlock(receivedBlockHash) != nil
		if blockInTree {
			continue
		}

		// check if parent is unknown
		parent := bc.blockTree.GetParentByBlock(block)
		if parent == nil {
			// Parent not present in tree so we cannot add block yet
			bc.waitingBlocks[receivedBlockHash] = block
			continue
		}
		delete(bc.waitingBlocks, receivedBlockHash)
		// check slot number
		if parent.Slot > block.Slot {
			continue
		}

		// check height
		if parent.Height != block.Height-1 {
			continue
		}

		// check that block is valid
		verified := block.VerifyBlock()
		if !verified {
			continue
		}

		// _, blockIsWaiting := bc.waitingBlocks[receivedBlockHash]
		// if blockIsWaiting {
		// 	return
		// }

		// // check that transactions are valid
		// ledger := bc.GetLedger(bc.blockTree.GetParent(block))
		// for _, transaction := range block.Transactions {
		// 	if !ledger.SignedTransaction(&transaction) {
		// 		// TODO maybe add to discarded blocks?
		// 		return
		// 	}
		// }
		bc.blockTree.InsertBlock(block)

		//rejustification
		if bc.UseFinalizer {
			bc.finalizer.RejustifyPreVotes(true)
			bc.finalizer.RejustifyAllVotes()
			bc.finalizer.RejustifyAllPreCommits()
		}

		for _, waitingBlock := range bc.waitingBlocks {
			if waitingBlock.Parent == block.Hash() {
				bc.HandleBlock(waitingBlock)
			}
		}
	}
}

func (bc *Protocol) GetTransactions() []types.SignedTransaction {
	return nil
}
func (bc *Protocol) GetLedger() Ledger {
	bestBlock := bc.blockTree.GetBestBlock()
	return bc.blockTree.GetLedger(bestBlock)
}
func (bc *Protocol) GetPublickey() string {
	return signature.PublicKeyToString(bc.privateKey.PublicKey)
}
func (bc *Protocol) MakeBlock() {
	block := new(types.Block)
	block.Slot = bc.slot
	bc.queuedTransactionsLock.Lock()
	for _, trans := range bc.queuedTransactions {
		block.Transactions = append(block.Transactions, trans)
	}
	bc.queuedTransactions = make(map[string]types.SignedTransaction)
	bc.queuedTransactionsLock.Unlock()
	block.PublicKey = bc.GetPublickey()
	parent := bc.blockTree.GetBestBlock()
	block.Parent = parent.Hash()
	// increase height
	block.Height = parent.Height + 1

	block.Sign(&bc.privateKey)

	(*bc.client).Send(block)
}

func (bc *Protocol) MakeTransaction(to rsa.PublicKey, amount int) {
	toString := signature.PublicKeyToString(to)
	transaction := types.NewSignedTransaction(strconv.Itoa(bc.transactionId), bc.privateKey, toString, amount)
	bc.transactionId++
	bc.HandleTransaction(transaction)
}

func (bc *Protocol) RegisterClient(client *network.ClientInterface) {
	bc.client = client
}

func (bc *Protocol) GetHardness() int {
	return bc.hardness
}

// Starts a thread which will run the lottery
// for this instance of the protocol
func (bc *Protocol) StartLottery() {
	go func() {
		for {
			(*bc.lottery).TryWin()
			bc.slot++
			time.Sleep(2000 * time.Millisecond)
		}
	}()
}

func (bc *Protocol) GetNumberOfPeers() int {
	return len(bc.publicKeys)
}
