package blockchain

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/Hvassaa/bachelor/types"
)

type TreeDrawer struct {
	ID         string
	Bc         *Blockchain
	CopyOfTree map[types.Hash]types.Block
}

func (t *TreeDrawer) DrawTree() {
	os.Mkdir("out", os.ModePerm)
	blockTree := (*t.Bc).GetBlockTree()
	tree := (*blockTree).GetCopyOfTree()
	finalBlocks := (*blockTree).GetFinals()

	uniqueHash := make(map[types.Hash]struct{})

	treeString := "digraph G {\n"

	for hash, block := range tree {
		uniqueHash[hash] = struct{}{}
		if (block.Parent == types.Hash{}) {
			continue
		}
		parentHash := block.Parent
		treeString = treeString + "a" + parentHash.ToString() + " -> " + "a" + hash.ToString() + "\n"
	}

	for uniqueH := range uniqueHash {
		identifier := uniqueH.ToString()[0:4]
		treeString = treeString + "a" + uniqueH.ToString() + " [label=\"" + identifier + "\"]\n"
	}

	for finalHash := range finalBlocks {
		treeString = treeString + "a" + finalHash.ToString() + " [color=red]\n"
	}

	treeString = treeString + "}"

	ioutil.WriteFile("out/"+t.ID+".dot", []byte(treeString), 0644)
}

func (t *TreeDrawer) getLeafs() map[types.Hash]types.Block {
	blockTree := (*t.Bc).GetBlockTree()
	tree1 := t.CopyOfTree
	tree2 := t.CopyMap(tree1)
	for hash := range tree1 {
		parent := (*blockTree).GetParent(hash)
		if parent != nil {
			delete(tree2, parent.Hash())
		}
	}
	return tree2
}

func (t *TreeDrawer) getBestChain() map[types.Hash]types.Block {
	blockTree := (*t.Bc).GetBlockTree()
	bestBLock := (*blockTree).GetBestBlock()
	chain := make(map[types.Hash]types.Block)
	chain[bestBLock.Hash()] = *bestBLock
	parent := (*blockTree).GetParent(bestBLock.Hash())
	for parent != nil {
		chain[parent.Hash()] = *parent
		parent = (*blockTree).GetParent(parent.Hash())
	}
	return chain
}

func (t *TreeDrawer) branchCounter() []string {
	bestChain := t.getBestChain()
	leafs := t.getLeafs()
	blockTree := (*t.Bc).GetBlockTree()
	lengths := []string{}
	for _, block := range leafs {
		currentBlock := block
		length := 0
		for {
			if _, exist := bestChain[currentBlock.Hash()]; exist {
				break
			}
			length += 1
			currentBlock = *(*blockTree).GetParent(currentBlock.Hash())
		}
		if length != 0 {
			lengths = append(lengths, strconv.Itoa(length))
		}
	}
	return lengths
}

func (t *TreeDrawer) WriteLengths(fname string, params []string) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	lengths := append(params, strconv.Itoa((len(t.getBestChain()))))
	lengths = append(lengths, strconv.Itoa(len(t.CopyOfTree)))
	lengths = append(lengths, strconv.Itoa((*t.Bc).GetLastFinal().Height))

	lengths = append(lengths, t.branchCounter()...)

	w := csv.NewWriter(f)
	w.Write(lengths)
	w.Flush()
	f.Close()
}

func (t *TreeDrawer) CopyMap(originalMap map[types.Hash]types.Block) map[types.Hash]types.Block {
	copyMap := make(map[types.Hash]types.Block)
	for hash, block := range originalMap {
		copyMap[hash] = block
	}
	return copyMap
}

func (t *TreeDrawer) WriteDistancesToLastFinals(fname string, params []string) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fields := []string{"Height", "FinalRatio", "Distance", "FinalHeights"}

	w := csv.NewWriter(f)
	w.Write(fields)

	// add distance to lasta final
	for _, b := range t.getBestChain() {
		output := []string{}

		distance := strconv.Itoa(b.DistanceToLastFinal)
		height := strconv.Itoa(b.Height)
		finalHeight := b.Height - b.DistanceToLastFinal
		output = append(output, height)
		output = append(output, distance)
		output = append(output, strconv.Itoa(finalHeight))
		w.Write(output)
	}

	w.Flush()
	f.Close()
}

func WriteSliceToCSV(fname string, slice []string) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	w := csv.NewWriter(f)
	w.Write(slice)
	w.Flush()
	f.Close()
}
