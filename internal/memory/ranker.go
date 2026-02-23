package memory

import "sort"

// Ranker ranks retrieval results by combining similarity score and importance.
type Ranker struct{}

// NewRanker creates a new Ranker.
func NewRanker() *Ranker { return &Ranker{} }

// RankedChunk pairs a Chunk with a retrieval score.
type RankedChunk struct {
	Chunk
	FinalScore float64
}

// RankedMemory pairs a Memory with a retrieval score.
type RankedMemory struct {
	Memory
	FinalScore float64
}

// RankChunks scores and sorts chunks by similarity, highest first.
// similarityByID maps chunk ID → cosine similarity (0-1).
func (r *Ranker) RankChunks(chunks []Chunk, similarityByID map[string]float64) []RankedChunk {
	ranked := make([]RankedChunk, 0, len(chunks))
	for _, c := range chunks {
		sim := similarityByID[c.ID]
		// Test files are deprioritised by default (importance 0.3).
		importance := 1.0
		if c.ChunkType == "test" {
			importance = 0.3
		}
		ranked = append(ranked, RankedChunk{
			Chunk:      c,
			FinalScore: sim * importance,
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].FinalScore > ranked[j].FinalScore
	})
	return ranked
}

// RankMemories scores and sorts memories by similarity × importance, highest first.
func (r *Ranker) RankMemories(memories []Memory, similarityByID map[string]float64) []RankedMemory {
	ranked := make([]RankedMemory, 0, len(memories))
	for _, m := range memories {
		sim := similarityByID[m.ID]
		// Importance is already 0-1 from the DB; use it as a multiplier.
		importance := m.Importance
		if importance == 0 {
			importance = 0.5
		}
		ranked = append(ranked, RankedMemory{
			Memory:     m,
			FinalScore: sim * importance,
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].FinalScore > ranked[j].FinalScore
	})
	return ranked
}
