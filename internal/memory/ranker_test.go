package memory

import "testing"

func TestRankChunks_SortsBySimilarity(t *testing.T) {
	chunks := []Chunk{
		{ID: "a", Content: "alpha", ChunkType: "code"},
		{ID: "b", Content: "beta", ChunkType: "code"},
		{ID: "c", Content: "gamma", ChunkType: "code"},
	}
	simMap := map[string]float64{"a": 0.5, "b": 0.9, "c": 0.7}

	ranker := NewRanker()
	ranked := ranker.RankChunks(chunks, simMap)

	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked chunks, got %d", len(ranked))
	}
	if ranked[0].ID != "b" {
		t.Errorf("expected highest ranked to be 'b', got %q", ranked[0].ID)
	}
	if ranked[1].ID != "c" {
		t.Errorf("expected second ranked to be 'c', got %q", ranked[1].ID)
	}
}

func TestRankChunks_TestFilesDeprioritised(t *testing.T) {
	chunks := []Chunk{
		{ID: "code1", Content: "impl", ChunkType: "code"},
		{ID: "test1", Content: "test", ChunkType: "test"},
	}
	simMap := map[string]float64{"code1": 0.5, "test1": 0.5}

	ranker := NewRanker()
	ranked := ranker.RankChunks(chunks, simMap)

	// Code chunk should rank higher because test files get 0.3 importance.
	if ranked[0].ID != "code1" {
		t.Errorf("expected code chunk to rank higher than test chunk")
	}
	// Test chunk: 0.5 * 0.3 = 0.15, Code chunk: 0.5 * 1.0 = 0.5
	if ranked[1].FinalScore >= ranked[0].FinalScore {
		t.Errorf("test chunk score %f should be less than code chunk score %f",
			ranked[1].FinalScore, ranked[0].FinalScore)
	}
}

func TestRankChunks_Empty(t *testing.T) {
	ranker := NewRanker()
	ranked := ranker.RankChunks(nil, nil)
	if len(ranked) != 0 {
		t.Errorf("expected empty result, got %d", len(ranked))
	}
}

func TestRankMemories_SortsBySimilarityTimesImportance(t *testing.T) {
	memories := []Memory{
		{ID: "m1", Content: "use React", Importance: 0.8},
		{ID: "m2", Content: "always test", Importance: 1.0},
		{ID: "m3", Content: "a note", Importance: 0.5},
	}
	simMap := map[string]float64{"m1": 0.9, "m2": 0.5, "m3": 0.8}

	ranker := NewRanker()
	ranked := ranker.RankMemories(memories, simMap)

	// m1: 0.9*0.8=0.72, m2: 0.5*1.0=0.5, m3: 0.8*0.5=0.4
	if ranked[0].ID != "m1" {
		t.Errorf("expected m1 first, got %q (score=%f)", ranked[0].ID, ranked[0].FinalScore)
	}
}

func TestRankMemories_ZeroImportanceUsesDefault(t *testing.T) {
	memories := []Memory{
		{ID: "m1", Content: "something", Importance: 0},
	}
	simMap := map[string]float64{"m1": 0.8}

	ranker := NewRanker()
	ranked := ranker.RankMemories(memories, simMap)

	// Zero importance defaults to 0.5, so score = 0.8 * 0.5 = 0.4
	expected := 0.4
	if ranked[0].FinalScore != expected {
		t.Errorf("expected score %f, got %f", expected, ranked[0].FinalScore)
	}
}
