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

type Round struct {
	// prevoteLock          sync.Mutex
	prevotes             map[string]*types.Message
	uniquePrevotingPeers map[string]struct{}
	// voteLock             sync.Mutex
	uniqueVotingPeers map[string]struct{}
	votes             map[string]*types.Message
	bufferedVotes     map[string]*types.Message
	// precommitLock        sync.Mutex
	preCommits          map[string]*types.Message
	bufferedPreCommmits map[string]*types.Message
	final               types.Hash
	steeringBlock       types.Hash
	round               int
	voted               bool
	precommited         bool
}

type Finalizer struct {
	bufferedPrevotes     map[string]*types.Message
	finalizerLock        sync.Mutex
	getRoundLock         sync.RWMutex
	privateKey           rsa.PrivateKey
	round                map[int]*Round
	roundLocks           map[int]*sync.Mutex
	client               *network.ClientInterface
	blockchain           *Blockchain
	twoThirds            int
	oneThird             int
	handleMessageChannel chan *types.Message
	FinalAndSTBDistances []string
}

func MakeFinalizer(bc *Blockchain, privateKey rsa.PrivateKey) *Finalizer {
	f := Finalizer{}
	f.privateKey = privateKey
	f.twoThirds = (2 * (*bc).GetNumberOfPeers()) / 3
	f.oneThird = (*bc).GetNumberOfPeers() / 3
	f.blockchain = bc
	f.round = make(map[int]*Round)
	f.roundLocks = make(map[int]*sync.Mutex)
	f.bufferedPrevotes = make(map[string]*types.Message)
	f.handleMessageChannel = make(chan *types.Message)
	f.FinalAndSTBDistances = make([]string, 0)
	go f.ActuallyHandleMessage()
	r := Round{
		prevotes:             make(map[string]*types.Message),
		votes:                make(map[string]*types.Message),
		preCommits:           make(map[string]*types.Message),
		uniquePrevotingPeers: make(map[string]struct{}),
		uniqueVotingPeers:    make(map[string]struct{}),
		steeringBlock:        (*bc).GetGenesisBlockHash(),
		// prevoteLock:          sync.Mutex{},
		// voteLock:             sync.Mutex{},
		// precommitLock:        sync.Mutex{},
	}
	var mutex = &sync.Mutex{}
	f.roundLocks[0] = mutex
	f.round[0] = &r

	return &f
}

func (f *Finalizer) RegisterClient(c *network.ClientInterface) {
	f.client = c
}

func (f *Finalizer) StartFinalizer() {
	go f.PreVote(1)
}

func (f *Finalizer) GetRound(round int) *Round {
	// if the round is already created, return
	f.getRoundLock.Lock()
	defer f.getRoundLock.Unlock()
	if r, alreadyExists := f.round[round]; alreadyExists {
		return r
	} else {
		f.addRound(round)
		return f.round[round]
	}
}

func (f *Finalizer) addRound(round int) {
	r := Round{
		prevotes:             make(map[string]*types.Message),
		votes:                make(map[string]*types.Message),
		preCommits:           make(map[string]*types.Message),
		uniquePrevotingPeers: make(map[string]struct{}),
		uniqueVotingPeers:    make(map[string]struct{}),
		steeringBlock:        types.Hash{},
		voted:                false,
		precommited:          false,
		bufferedVotes:        make(map[string]*types.Message),
		bufferedPreCommmits:  make(map[string]*types.Message),
		round:                round,
		// prevoteLock:          sync.Mutex{},
		// voteLock:             sync.Mutex{},
		// precommitLock:        sync.Mutex{},
	}
	var mutex = &sync.Mutex{}
	f.roundLocks[round] = mutex
	f.round[round] = &r
}

func (f *Finalizer) PreVote(round int) {
	// fmt.Println(runtime.NumGoroutine())
	//Old sleep
	//time.Sleep(2 * time.Second)
	lastRound := f.GetRound(round - 1)
	lastSTB := lastRound.steeringBlock
	if (lastSTB == types.Hash{}) {
		return
	}
	tree := (*f.blockchain).GetBlockTree()
	lastFinal := (*tree).GetLastFinal()
	bb := (*tree).GetBestBlock()
	newBlocks := bb.Height - lastFinal.Height
	for newBlocks < 5 {
		time.Sleep(1 * time.Second)
		bb = (*tree).GetBestBlock()
		newBlocks = bb.Height - lastFinal.Height
	}

	bestBlock := (*tree).GetBestBlockInSubTree(lastSTB)

	preVote := types.Message{
		Identifier: "prevote",
		Round:      round,
		Block:      bestBlock,
		Peer:       (*f.blockchain).GetPublickey()}

	preVoteByte := preVote.ToBytes()
	preVote.Signature = signature.Sign(&f.privateKey, preVoteByte)
	(*f.client).Send(preVote)
}

func (f *Finalizer) Vote(round int) {

	r := f.GetRound(round)
	voteBLock := f.ghost(f.oneThird, r.prevotes)
	vote := types.Message{
		Identifier: "vote",
		Round:      round,
		Block:      voteBLock,
		Peer:       (*f.blockchain).GetPublickey()}

	voteByte := vote.ToBytes()
	vote.Signature = signature.Sign(&f.privateKey, voteByte)
	(*f.client).Send(vote)
}

func (f *Finalizer) preCommit(round int) {

	r := f.GetRound(round)
	preCommitBlock := f.ghost(f.twoThirds, r.votes)

	preCommit := types.Message{
		Identifier: "precommit",
		Round:      round,
		Block:      preCommitBlock,
		Peer:       (*f.blockchain).GetPublickey()}

	preCommitByte := preCommit.ToBytes()
	preCommit.Signature = signature.Sign(&f.privateKey, preCommitByte)
	(*f.client).Send(preCommit)
}
func (f *Finalizer) ghost(threshold int, voteMap map[string]*types.Message) types.Hash {
	tMap := make(map[types.Hash]int)
	uniqueVotesOnBlocks := make(map[string]struct{})
	tree := (*f.blockchain).GetBlockTree()
	lastFinal := (*tree).GetLastFinal()
	bestHash := types.Hash{}
	bestHeight := 0
	// for _, p := range voteMap {
	for _, hash := range (*tree).GetList() {
		tMap[hash] = 0
	}
	for _, m := range voteMap {
		currentBlock := (*tree).GetBlock(m.Block)
		for currentBlock != nil && currentBlock.Hash() != lastFinal.Parent {
			currentHash := currentBlock.Hash()
			uniqueVoteString := string(currentHash[:]) + m.Peer
			if _, exist := uniqueVotesOnBlocks[uniqueVoteString]; !exist {
				tMap[currentHash] += 1
				uniqueVotesOnBlocks[uniqueVoteString] = struct{}{}
			}
			currentBlock = (*tree).GetParent(currentHash)
		}
	}
	for h := range tMap {
		if threshold < tMap[h] {
			height := (*tree).GetBlock(h).Height
			if bestHeight < height || (bestHeight == height && GreaterThan(h, bestHash)) {
				// fmt.Println("Height | ", h.ToString()[:4], " | ", height, " | Old best height: ", bestHeight)

				bestHeight = height
				bestHash = h
			}
		}

	}
	// bestBlock := (*tree).GetBlock(bestHash)

	return bestHash
}

// rewrite
func (f *Finalizer) RejustifyPreVotes(lock bool) {
	go func() {
		f.finalizerLock.Lock()
		defer f.finalizerLock.Unlock()
		tree := (*f.blockchain).GetBlockTree()

		for _, bufferedVote := range f.bufferedPrevotes {
			lastRound := f.GetRound(bufferedVote.Round - 1)
			round := f.GetRound(bufferedVote.Round)
			f.roundLocks[round.round].Lock()
			lastSTB := (*tree).GetBlock(lastRound.steeringBlock)
			votedBlock := (*tree).GetBlock(bufferedVote.Block)
			justified := (*tree).IsDescendent(lastSTB, votedBlock)
			if justified {
				messageString := bufferedVote.ToString()
				round.prevotes[messageString] = bufferedVote
				round.uniquePrevotingPeers[bufferedVote.Peer] = struct{}{}
				delete(f.bufferedPrevotes, messageString)
				if len(round.uniquePrevotingPeers) > f.twoThirds {
					f.Vote(round.round)
				}
			}
			//delete old votes
			f.roundLocks[round.round].Unlock()

		}
	}()
}

func (f *Finalizer) JustifyVote(round int, root types.Hash) bool {
	//Er det et problem at excessvotesInSubstree kun er unikke i hvert subtræ og noOfPrevotes er helt unikke? Flere stemmer fra advesary kan tælle i votesInSubtree og ikke i noOfPrevots...
	r := f.GetRound(round)
	tree := (*f.blockchain).GetBlockTree()
	lastRound := f.GetRound(round - 1)
	lastSTB := (*tree).GetBlock(lastRound.steeringBlock)
	votedBlock := (*tree).GetBlock(root)
	justifiedSubtreeOfSTB := (*tree).IsDescendent(lastSTB, votedBlock)
	noOfPrevotes := len(r.uniquePrevotingPeers)
	if noOfPrevotes < f.twoThirds {
		return false
	}
	votesInSubtree := f.UniqueVotesInSubtree(root, r.prevotes)
	children := (*tree).GetChildren(root)
	voteOnBlock := f.UniqueVotesInBlock(root, r.prevotes)
	excessVotesInChildren := 0
	totalVotesInChildren := 0
	for _, childHash := range children {
		votesInChild := f.UniqueVotesInSubtree(childHash, r.prevotes)
		totalVotesInChildren += votesInChild
		if votesInChild > f.oneThird {
			excessVotesInChild := votesInChild - f.oneThird
			excessVotesInChildren += excessVotesInChild
		}
	}
	// lol this is ugly
	uniqueExcessVotesInSubtree := excessVotesInChildren - (totalVotesInChildren - (votesInSubtree - voteOnBlock))
	possibleVotesSeen := noOfPrevotes - uniqueExcessVotesInSubtree
	enoughVotes := possibleVotesSeen > f.twoThirds && votesInSubtree > f.oneThird
	return enoughVotes && justifiedSubtreeOfSTB
}

func (f *Finalizer) RejustifyVote(round int) {

	r := f.GetRound(round)
	for ms, m := range r.bufferedVotes {
		isJustified := f.JustifyVote(round, m.Block)
		if isJustified {
			r.votes[ms] = m
			r.uniqueVotingPeers[m.Peer] = struct{}{}
			delete(r.bufferedVotes, ms)
			if len(r.uniqueVotingPeers) > f.twoThirds {
				f.preCommit(r.round)
			}
		}
	}
}

func (f *Finalizer) RejustifyAllVotes() {
	go func() {
		copyOfRounds := f.GetCopyOfRounds()
		for i := range copyOfRounds {
			f.roundLocks[i].Lock()
			f.RejustifyVote(i)
			f.roundLocks[i].Unlock()

		}
	}()
}

func (f *Finalizer) JustifyPrecommit(round int, root types.Hash) bool {
	r := f.GetRound(round)
	votesInSubtree := f.UniqueVotesInSubtree(root, r.votes)
	enoughVotesInSubtree := votesInSubtree > f.twoThirds
	return enoughVotesInSubtree
}

func (f *Finalizer) RejustifyPreCommit(round int) {
	r := f.GetRound(round)
	for ms, m := range r.bufferedPreCommmits {
		isJustified := f.JustifyPrecommit(round, m.Block)
		if isJustified {
			r.preCommits[ms] = m
			delete(r.bufferedPreCommmits, ms)
			f.UpdateSTB(r.round)
			f.UpdateFinal(r.round)
		}
	}
}

func (f *Finalizer) RejustifyAllPreCommits() {
	go func() {
		copyOfRounds := f.GetCopyOfRounds()
		for i := range copyOfRounds {
			f.roundLocks[i].Lock()
			f.RejustifyPreCommit(i)
			f.roundLocks[i].Unlock()
		}
	}()
}

func (f *Finalizer) UniqueVotesInSubtree(root types.Hash, voteMap map[string]*types.Message) int {
	// votesInSubtree := 0
	uniquePeers := make(map[string]struct{})
	tree := (*(*f.blockchain).GetBlockTree())
	lastFinal := tree.GetLastFinal()
	for _, m := range voteMap {
		hash := m.Block
		currentBlock := tree.GetBlock(hash)
		for currentBlock != nil && currentBlock.Hash() != lastFinal.Parent {
			currentHash := currentBlock.Hash()
			if currentHash == root {
				// votesInSubtree++
				uniquePeers[m.Peer] = struct{}{}
				break
			}
			currentBlock = tree.GetParent(currentHash)
		}
	}
	return len(uniquePeers)
}

func (f *Finalizer) UniqueVotesInBlock(block types.Hash, voteMap map[string]*types.Message) int {
	uniquePeers := make(map[string]struct{})
	for _, m := range voteMap {
		hash := m.Block
		if hash == block {
			uniquePeers[m.Peer] = struct{}{}
		}
	}
	return len(uniquePeers)
}

func (f *Finalizer) UpdateFinal(round int) {
	r := f.GetRound(round)
	tree := (*(*f.blockchain).GetBlockTree())
	currentBestSlot := -1
	currentBestHash := types.Hash{}
	for _, mesu := range r.preCommits {
		var currentVotes int
		// r.precommitLock.Lock()
		currentVotes = f.UniqueVotesInSubtree(mesu.Block, r.preCommits)
		// r.precommitLock.Unlock()
		if currentVotes > f.twoThirds {
			currentBlock := tree.GetBlock(mesu.Block)
			if currentBlock.Slot > currentBestSlot || currentBestSlot == -1 {
				currentBestSlot = currentBlock.Slot
				currentBestHash = currentBlock.Hash()
			}
		}
	}
	if (currentBestHash != types.Hash{} && currentBestHash != r.final) {
		tree.AddFinal(round, currentBestHash)
		r.final = currentBestHash
		// find distance between stb and final
		if (r.steeringBlock != types.Hash{}) {
			finalBlockHeight := tree.GetBlock(currentBestHash).Height
			steeringBlockHeight := tree.GetBlock(r.steeringBlock).Height
			difference := strconv.Itoa(finalBlockHeight - steeringBlockHeight)
			f.FinalAndSTBDistances = append(f.FinalAndSTBDistances, difference)
		}
	}
}

func (f *Finalizer) UpdateSTB(round int) {
	r := f.GetRound(round)
	tree := (*(*f.blockchain).GetBlockTree())
	currentBestSlot := -1
	currentBestHash := types.Hash{}
	for _, mesu := range r.preCommits {
		var currentVotes int
		currentVotes = f.UniqueVotesInChain(round, mesu.Block)
		if currentVotes > f.twoThirds {
			currentBlock := tree.GetBlock(mesu.Block)
			if currentBlock.Slot < currentBestSlot || currentBestSlot == -1 {
				currentBestSlot = currentBlock.Slot
				currentBestHash = currentBlock.Hash()
			}
		}
	}

	if (currentBestHash != types.Hash{}) {
		if (r.steeringBlock == types.Hash{}) {
			r.steeringBlock = currentBestHash

			go f.PreVote(round + 1)
		}
		r.steeringBlock = currentBestHash

		if r.steeringBlock != currentBestHash {
			//might not be nessecary
			f.RejustifyPreVotes(false)
			f.RejustifyVote(round + 1)
		}
	}
}

func (f *Finalizer) UniqueVotesInChain(round int, lowestBlock types.Hash) int {
	uniquePeers := make(map[string]struct{})
	r := f.GetRound(round)
	tree := (*(*f.blockchain).GetBlockTree())
	lastFinal := tree.GetLastFinal()
	currentBlock := tree.GetBlock(lowestBlock)
	for currentBlock != nil && currentBlock.Hash() != lastFinal.Parent {
		currentHash := currentBlock.Hash()
		for _, mes := range r.preCommits {
			if mes.Block == currentHash {
				uniquePeers[mes.Peer] = struct{}{}
			}
		}
		currentBlock = tree.GetParent(currentHash)
	}
	return len(uniquePeers)
}

func (f *Finalizer) CheckBufferedVotes(round int) {
	r := f.GetRound(round)
	for _, message := range r.bufferedVotes {
		isJustified := f.JustifyVote(round, message.Block)
		if isJustified {
			messageString := message.ToString()
			r.votes[messageString] = message
			delete(r.bufferedVotes, messageString)
		}
	}
}

func (f *Finalizer) CheckBufferedPreCommits(round int) {
	r := f.GetRound(round)
	for _, message := range r.bufferedPreCommmits {
		isJustified := f.JustifyPrecommit(round, message.Block)
		if isJustified {
			messageString := message.ToString()
			r.preCommits[messageString] = message
			delete(r.bufferedPreCommmits, messageString)
		}
	}
}

func (f *Finalizer) HandleMessage(message *types.Message) {
	go func() {
		f.handleMessageChannel <- message
	}()
}

// TODO check buffered votes/messages
func (f *Finalizer) ActuallyHandleMessage() {
	for {
		message := <-f.handleMessageChannel
		round := message.Round
		r := f.GetRound(round)
		f.roundLocks[round].Lock() // DANGER
		tree := (*f.blockchain).GetBlockTree()
		//verify message
		messageByte := message.ToBytes()
		publicKey := signature.PublicKeyFromString(message.Peer)
		messageSignature := message.Signature
		valid := signature.Verify(&publicKey, messageByte, messageSignature)
		if !valid {
			f.roundLocks[round].Unlock()
			continue
		}
		// block received in some vote
		votedBlock := (*tree).GetBlock(message.Block)
		messageString := message.ToString()
		switch message.Identifier {
		case "prevote":
			// r.prevoteLock.Lock()
			// defer r.prevoteLock.Unlock()
			if _, in := r.prevotes[messageString]; in {
				f.roundLocks[round].Unlock()
				continue
			}
			lastRound := f.GetRound(round - 1)
			lastSTB := (*tree).GetBlock(lastRound.steeringBlock)
			justified := (*tree).IsDescendent(lastSTB, votedBlock)
			// fmt.Println("prevote | ", justified, " | ", message.block.ToString()[:4])

			if !justified {
				f.roundLocks[round].Unlock()
				f.finalizerLock.Lock()
				f.bufferedPrevotes[messageString] = message
				f.finalizerLock.Unlock()
				continue
			}
			// (*f.client).Send(message)
			r.prevotes[messageString] = message
			r.uniquePrevotingPeers[message.Peer] = struct{}{}
			if len(r.uniquePrevotingPeers) > f.twoThirds {
				if !r.voted {
					f.Vote(round)
					r.voted = true
				}
				f.RejustifyVote(round)
			}
		case "vote":

			// r.voteLock.Lock()
			// defer r.voteLock.Unlock()
			if _, in := r.votes[messageString]; in {
				f.roundLocks[round].Unlock()
				continue
			}
			justifiedVote := f.JustifyVote(round, message.Block)
			// fmt.Println("Vote | ", justifiedVote, " | ", message.block.ToString()[:4])
			if justifiedVote {
				r.votes[messageString] = message
				// (*f.client).Send(message)
				r.uniqueVotingPeers[message.Peer] = struct{}{}
				if len(r.uniqueVotingPeers) > f.twoThirds {
					if !r.precommited {
						f.preCommit(round)
						r.precommited = true
					}
					f.RejustifyPreCommit(round)
				}
			} else {
				r.bufferedVotes[messageString] = message
			}
			//Rejustify
		case "precommit":

			// r.precommitLock.Lock()
			// defer r.precommitLock.Unlock()``
			if _, in := r.preCommits[messageString]; in {
				f.roundLocks[round].Unlock()
				continue
			}
			justifiedPreCommit := f.JustifyPrecommit(round, message.Block)
			// fmt.Println("Precommit | ", justifiedPreCommit, " | ", message.block.ToString()[:4])
			if justifiedPreCommit {
				r.preCommits[messageString] = message
				// (*f.client).Send(message)
				// update steeringBlock and final block
				f.UpdateSTB(round)
				f.UpdateFinal(round)
			} else {
				r.bufferedPreCommmits[messageString] = message
			}
		}
		f.roundLocks[round].Unlock()
	}
}

func (f *Finalizer) GetCopyOfRounds() map[int]*Round {
	copyMap := make(map[int]*Round)
	f.getRoundLock.Lock()
	defer f.getRoundLock.Unlock()
	for hash, block := range f.round {
		copyMap[hash] = block
	}
	return copyMap
}
