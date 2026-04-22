package engine

type MemoryRelationshipStore struct {
	lovers *LoverPair
}

func NewMemoryRelationshipStore() *MemoryRelationshipStore {
	return &MemoryRelationshipStore{}
}

func (s *MemoryRelationshipStore) SetLovers(pair *LoverPair) {
	s.lovers = pair
}

func (s *MemoryRelationshipStore) Lovers() *LoverPair {
	return s.lovers
}
