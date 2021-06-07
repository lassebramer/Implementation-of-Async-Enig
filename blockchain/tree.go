package blockchain

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"

	"github.com/Hvassaa/bachelor/types"
)

type (
	BlockTree interface {
		GetBestBlockInSubTree(root types.Hash) types.Hash
		GetParent(hash types.Hash) *types.Block
		InsertBlock(block *types.Block)
		GetLedger(block *types.Block) Ledger
		IsDescendent(parent *types.Block, child *types.Block) bool
		IsDescendentNoLock(parent *types.Block, child *types.Block) bool
		// Contains(id Hash) bool
		GetBlock(id types.Hash) *types.Block
		GetBlockNoLock(id types.Hash) *types.Block
		GetBestBlock() *types.Block
		GetList() []types.Hash
		GetParentByBlock(block *types.Block) *types.Block
		GetChildren(parent types.Hash) []types.Hash
		GetCopyOfTree() map[types.Hash]types.Block
		AddFinal(roind int, hash types.Hash)
		GetFinals() map[types.Hash]int
		ReadTree(filename string)
		WriteTree()
		GetLastFinal() *types.Block

		// GetParameters(slot int, predecessor *BlockChain) *LeaderElectionParameters
		// AddFinalityProof(proof *FinalityProof)
		// GetLastFinalityProof() *FinalityProof
		// GetLastFinalityProofOnChain(chain *BlockChain) *FinalityProof
		// SubscribeAtHeight(height int) chan *BlockChain
		// SubscribeToHash(hash Hash) chan *BlockChain
	}

	Tree struct {
		latestFinal         types.Hash
		tree                map[types.Hash]*types.Block
		startingMoney       int
		initialStakeHolders []string
		treeLock            sync.RWMutex
		bestBlock           types.Hash
		finalBlocks         map[int]types.Hash
	}
)

func (t *Tree) GetLastFinal() *types.Block {
	return t.GetBlock(t.latestFinal)
}

func (t *Tree) AddFinal(round int, hash types.Hash) {
	t.latestFinal = hash
	t.finalBlocks[round] = hash
	finalBlock := t.GetBlock(hash)
	bestBlock := t.GetBestBlock()
	bestBlockInFinalSubtree := t.IsDescendent(finalBlock, bestBlock)
	if !bestBlockInFinalSubtree {
		t.bestBlock = t.GetBestBlockInSubTree(hash)
	}
}

func (t *Tree) GetFinals() map[types.Hash]int {
	reverseMap := make(map[types.Hash]int)
	for round, hash := range t.finalBlocks {
		mapRound, exsist := reverseMap[hash]
		if exsist {
			if round < mapRound {
				reverseMap[hash] = round
			}
		} else {
			reverseMap[hash] = round
		}
	}
	return reverseMap
}

// Function used to compare byte arrays
func GreaterThan(h1, h2 types.Hash) bool {
	for i := range h1 {
		if h1[i] > h2[i] {
			return true
		} else if h1[i] < h2[i] {
			return false
		}
	}
	return false
}

func (t *Tree) GetChildren(parent types.Hash) []types.Hash {
	childList := make([]types.Hash, 1)
	t.treeLock.RLock()
	// fmt.Println("LOCK: GetChildren RLock")
	for _, block := range t.tree {
		if block.Parent == parent {
			childList = append(childList, block.Hash())
		}
	}
	t.treeLock.RUnlock()
	// fmt.Println("LOCK: GetChildren RUnLock")
	return childList
}

func MakeBlockTree(initialStakeHolders []string, gBlock types.Block) BlockTree {
	t := new(Tree)
	t.tree = make(map[types.Hash]*types.Block)
	t.finalBlocks = make(map[int]types.Hash)
	gHash := gBlock.Hash()
	t.tree[gHash] = &gBlock
	t.bestBlock = gHash
	t.startingMoney = 1000
	t.latestFinal = gHash
	t.initialStakeHolders = initialStakeHolders

	return t
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Got error: ", err)
	}
}

func (t *Tree) WriteTree() {
	file, err := os.Create("../main/out/writtenTree.gob")
	CheckError(err)
	defer file.Close()

	encoder := gob.NewEncoder(file)
	tree := t.GetCopyOfTree()
	encoder.Encode(tree)
}

func (t *Tree) ReadTree(filename string) {
	file, err := os.Open(filename)
	CheckError(err)
	defer file.Close()

	decoder := gob.NewDecoder(file)
	decoder.Decode(&t.tree)
}

func (t *Tree) GetParent(hash types.Hash) *types.Block {
	block := t.GetBlock(hash)
	parentHash := block.Parent
	return t.GetBlock(parentHash)
}

func (t *Tree) GetParentNoLock(hash types.Hash) *types.Block {
	block := t.GetBlockNoLock(hash)
	parentHash := block.Parent
	return t.GetBlockNoLock(parentHash)
}

func (t *Tree) GetBlock(id types.Hash) *types.Block {
	t.treeLock.RLock()
	// fmt.Println("LOCK: GetBlock RLock")
	b := t.tree[id]
	t.treeLock.RUnlock()
	// fmt.Println("LOCK: GetBlock RUnlock")
	return b
}

func (t *Tree) GetBlockNoLock(id types.Hash) *types.Block {
	b := t.tree[id]
	return b
}

func (t *Tree) GetParentByBlock(block *types.Block) *types.Block {
	t.treeLock.RLock()
	// fmt.Println("LOCK: GetParentByBlock RLock")
	parent := t.tree[block.Parent]
	t.treeLock.RUnlock()
	// fmt.Println("LOCK: GetParentByBlock RUnlock")
	return parent
}

func (t *Tree) InsertBlock(block *types.Block) {
	t.treeLock.Lock()

	hash := block.Hash()
	bb := t.GetBlockNoLock(t.bestBlock)
	lastFinalHash := t.latestFinal
	lastFinalBlock := t.GetBlockNoLock(lastFinalHash)
	blockInFinalSubtree := t.IsDescendentNoLock(lastFinalBlock, block)
	if block.Height > bb.Height && blockInFinalSubtree {
		t.bestBlock = hash
	}

	if block.Height == bb.Height && GreaterThan(block.Hash(), bb.Hash()) && blockInFinalSubtree {
		t.bestBlock = hash
	}
	// find lenght to last final
	distanceToLastFinal := block.Height - lastFinalBlock.Height
	block.DistanceToLastFinal = distanceToLastFinal

	t.tree[hash] = block
	t.treeLock.Unlock()
}

// func (t *Tree) GetList() map[Hash]*Block {
// 	return t.tree
// }

func (t *Tree) GetBestBlock() *types.Block {
	return t.GetBlock(t.bestBlock)
}

func (t *Tree) GetBestBlockInSubTree(root types.Hash) types.Hash {
	bestBlock := types.Hash{}
	bestHeight := 0
	t.treeLock.RLock()
	rootBlock := t.GetBlockNoLock(root)
	// fmt.Println("LOCK: GetBestBlockInSubTree RLock")
	for _, currentBlock := range t.tree {
		if t.IsDescendentNoLock(rootBlock, currentBlock) {
			if currentBlock.Height > bestHeight {
				bestBlock = currentBlock.Hash()
				bestHeight = currentBlock.Height
			} else if currentBlock.Height == bestHeight && GreaterThan(currentBlock.Hash(), bestBlock) {
				bestBlock = currentBlock.Hash()
				bestHeight = currentBlock.Height
			}
		}
	}
	t.treeLock.RUnlock()
	// fmt.Println("LOCK: GetBestBlockInSubTree RUnlock")
	return bestBlock

}

func (t *Tree) IsDescendent(parent *types.Block, child *types.Block) bool {
	currentBlock := child

	for currentBlock != nil && parent != nil {
		if currentBlock.Hash() == parent.Hash() {
			return true
		}
		currentBlock = t.GetBlock(currentBlock.Parent)
	}
	return false
}

func (t *Tree) IsDescendentNoLock(parent *types.Block, child *types.Block) bool {
	currentBlock := child

	for currentBlock != nil && parent != nil {
		if currentBlock.Hash() == parent.Hash() {
			return true
		}
		currentBlock = t.GetBlockNoLock(currentBlock.Parent)
	}
	return false
}

func (t *Tree) GetLedger(block *types.Block) Ledger {
	// the ledger to update and return
	resultLedger := MakeLedger()

	// give starting money, to all initialStakeHolders
	for _, initialStakeHolder := range t.initialStakeHolders {
		resultLedger.Accounts[initialStakeHolder] = t.startingMoney
	}

	// run through all blocks on chain and run transactions
	currentBlock := block
	for currentBlock != nil {
		for _, transaction := range currentBlock.Transactions {
			resultLedger.ForceTransaction(transaction.From, transaction.To, transaction.Amount)
			// give/take gas money
			resultLedger.ForceTransaction(transaction.To, block.PublicKey, 1)
		}
		currentBlock = t.GetBlock(currentBlock.Parent)
	}
	return *resultLedger
}

func (t *Tree) GetList() []types.Hash {
	hashes := []types.Hash{}
	t.treeLock.RLock()
	defer t.treeLock.RUnlock()
	for k := range t.tree {
		hashes = append(hashes, k)
	}
	return hashes
}

func (t *Tree) GetCopyOfTree() map[types.Hash]types.Block {
	copyMap := make(map[types.Hash]types.Block)
	t.treeLock.Lock()
	defer t.treeLock.Unlock()
	for hash, block := range t.tree {
		copyMap[hash] = *block
	}
	return copyMap
}
