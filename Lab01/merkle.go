package Lab01

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"hash"
)

// hash region
var DefaultShaHasher = NewHasher(sha256.New)

type HashFn func() hash.Hash
type Hasher struct{ Imp HashFn }

func NewHasher(imp HashFn) Hasher {
	return Hasher{imp}
}

func (hr *Hasher) Hash(data ...[]byte) []byte {
	h := hr.Imp()
	for _, d := range data {
		h.Write(d)
	}
	return h.Sum(nil)
}

// --- --- ---

// MerkleTree region
type MerkleTree struct {
	Root   *MerkleNode
	Leaves []*MerkleNode
	hasher Hasher
}

type MerkleNode struct {
	Parent *MerkleNode
	Left   *MerkleNode
	Right  *MerkleNode
	Data   []byte
}

// NewMerkleNode return a new Merkle node
func NewMerkleNode(left, right *MerkleNode, data []byte, h Hasher) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := h.Hash(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := h.Hash(prevHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

// NewMerkleTree return a new Merkle tree built on list of data with hasher function
func NewMerkleTree(data [][]byte, h Hasher) *MerkleTree {
	tree := &MerkleTree{
		Leaves: make([]*MerkleNode, 0, len(data)),
		hasher: h,
	}

	// add leaf nodes.
	for _, d := range data {
		tree.Leaves = append(tree.Leaves, &MerkleNode{Data: d})
	}

	tree.Root = tree.buildRoot()
	return tree
}

// VerifyProof verifies the integrity of the given value
func (tree *MerkleTree) VerifyProof(value []byte, proofs [][]byte, idxs []int, h Hasher) bool {
	prevHash := value
	for i := 0; i < len(proofs); i++ {
		if idxs[i] == 0 {
			prevHash = h.Hash(proofs[i], prevHash)
		} else {
			prevHash = h.Hash(prevHash, proofs[i])
		}
	}

	return bytes.Equal(tree.Root.Data, prevHash)
}

// GetProof returns the Merkle path proof to verify the integrity of the given data.
func (tree *MerkleTree) GetProof(data []byte) ([][]byte, []int, error) {
	var (
		path [][]byte
		idxs []int
	)

	// find the leaf node for the specific hash with the leaves list in tree.
	for _, currentNode := range tree.Leaves {
		if bytes.Equal(currentNode.Data, data) {
			// after finding the node, using the relationship of the nodes to find path.
			parent := currentNode.Parent
			for parent != nil {
				// if the current node is the left child, then need the right child to calculate the parent hash
				// for the proof and vice versa.
				// i.e:
				// if CurrentNode == Left ; ParentHash = (CurrentNode.Hash, RightChild.Hash)
				// if CurrentNode == Right ; ParentHash = (LeftChild.Hash, CurrentNode.Hash)
				// so we have to add the corresponding hash to the path, and in idxs, we save the hash's position 0
				// for left and 1 for right. In this way, when we want to verify the proof, we can know if
				// the given hash is the left o right child.
				if bytes.Equal(currentNode.Data, parent.Left.Data) {
					path = append(path, parent.Right.Data)
					idxs = append(idxs, 1)
				} else {
					path = append(path, parent.Left.Data)
					idxs = append(idxs, 0)
				}
				// continue the loop
				currentNode = parent
				parent = currentNode.Parent
			}
			return path, idxs, nil
		}
	}
	return path, idxs, errors.New("hash does not belong to the tree")
}

// verify by rebuild the tree and compare new tree hash with current hash.
func (tree *MerkleTree) Verify() bool {
	if len(tree.Leaves) == 0 || tree.Root == nil {
		return false
	}

	cr := tree.buildRoot()
	return bytes.Equal(tree.Root.Data, cr.Data)
}

// build tree with appended child leaves
func (tree *MerkleTree) buildRoot() *MerkleNode {
	nodes := tree.Leaves
	// iterating until reach a single node, which will be root.
	for len(nodes) > 1 {
		var parents []*MerkleNode

		// handle case number of node odd
		// duplicate the last node to concatenate it with itself.
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		// pairing nodes to build a parent from the pair
		for i := 0; i < len(nodes); i += 2 {
			n := &MerkleNode{
				Left:  nodes[i],
				Right: nodes[i+1],

				// compute the hash of the new node, which will be the combination of its children's hashes.
				Data: tree.hasher.Hash(nodes[i].Data, nodes[i+1].Data),
			}

			parents = append(parents, n)
			nodes[i].Parent, nodes[i+1].Parent = n, n
		}
		// once all possible pairs are processed, the parents become the children, start all over again.
		nodes = parents
	}

	return nodes[0]
}

// --- --- ---
