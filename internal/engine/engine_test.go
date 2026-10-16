package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPipeline(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "test.csv")
	data := "Latitude,Longitude,Status,ServiceName,CreatedDate\n39.30,-76.61,Open,Pothole,2026-02-10\n0,0,Closed,Trash,2026-02-11\n"
	if e := os.WriteFile(p, []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
	rs, skip, e := ParseFile(context.Background(), "requests", p, 2)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 1 || skip != 1 {
		t.Fatalf("records=%d skipped=%d", len(rs), skip)
	}
	r := Analyze(rs, skip)
	if r.Requests != 1 || r.Areas[0].Infrastructure != 1 || r.Areas[0].Open != 1 {
		t.Fatalf("unexpected result: %+v", r)
	}
	if !strings.Contains(Report(r), "Q5:") {
		t.Fatal("missing fifth question")
	}
}
