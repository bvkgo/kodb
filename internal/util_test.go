package internal

import "testing"

func TestIndexRuneN(t *testing.T) {
	if p := indexRuneN("/a/b/c", '/', 0); p != -1 {
		t.Fatalf("want -1 got %d", p)
	}
	if p := indexRuneN("/a/b/c", '@', 1); p != -1 {
		t.Fatalf("want -1 got %d", p)
	}
	if p := indexRuneN("/a/b/c", '/', 4); p != -1 {
		t.Fatalf("want -1 got %d", p)
	}

	if p := indexRuneN("/a/b/c", '/', 1); p != 0 {
		t.Fatalf("want 0 got %d", p)
	}
	if p := indexRuneN("/a/b/c/d/e/f", '/', 2); p != 2 {
		t.Fatalf("want 2 got %d", p)
	}
	if p := indexRuneN("a/b/c/d/e/f", '/', 3); p != 5 {
		t.Fatalf("want 5 got %d", p)
	}
}
