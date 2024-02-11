package datastructure

type trieNode struct {
	children    map[rune]*trieNode // Children of this node
	failure     *trieNode          // Failure link
	outputs     []string           // Output strings at this node
	isEndOfWord bool               // Is this node the end of a word
}

func newTrieNode() *trieNode {
	return &trieNode{children: make(map[rune]*trieNode), outputs: []string{}}
}

func newTrie() *trieNode {
	return newTrieNode()
}

func (t *trieNode) addPattern(pattern string) {
	currentNode := t
	for _, ch := range pattern {
		if _, ok := currentNode.children[ch]; !ok {
			currentNode.children[ch] = newTrieNode()
		}
		currentNode = currentNode.children[ch]
	}
	currentNode.isEndOfWord = true // Mark the end of a word
	currentNode.outputs = append(currentNode.outputs, pattern)
}

func (t *trieNode) buildFailureLinks() {
	queue := []*trieNode{}
	for _, child := range t.children {
		child.failure = t // Root's children fail back to the root
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for key, child := range current.children {
			queue = append(queue, child)
			failNode := current.failure
			for failNode != nil && failNode.children[key] == nil {
				failNode = failNode.failure
			}
			if failNode != nil {
				child.failure = failNode.children[key]
				child.outputs = append(child.outputs, child.failure.outputs...)
			} else {
				child.failure = t
			}
		}
	}
}

func (t *trieNode) search(text string) (foundPatterns []string) {
	currentState := t
	for _, ch := range text {
		for currentState != nil && currentState.children[ch] == nil {
			currentState = currentState.failure
		}
		if currentState == nil {
			currentState = t // Reset to root if no valid state is found
			continue
		}
		currentState = currentState.children[ch]
		if currentState.isEndOfWord {
			foundPatterns = append(foundPatterns, currentState.outputs...)
		}
	}
	return foundPatterns
}
