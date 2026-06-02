package filters

import (
	"testing"

	"github.com/rustyeddy/redeye"
)

// mockFilter is a lightweight Filter implementation used only in tests.
// It records the config string passed to Init so tests can inspect it.
type mockFilter struct {
	Flt
	initConfig string
}

func (m *mockFilter) Init(config string) { m.initConfig = config }
func (m *mockFilter) Filter(f *redeye.Frame) *redeye.Frame { return f }

// registerMock adds a named mockFilter to the global Filters map and returns
// both the filter and a cleanup function that removes it afterwards.
func registerMock(name string) (*mockFilter, func()) {
	mf := &mockFilter{Flt: Flt{name: name}}
	Filters.Add(name, mf)
	return mf, func() { delete(Filters, name) }
}

func TestNewPipelineEmpty(t *testing.T) {
	p := NewPipeline("")
	if len(p.Filters) != 0 {
		t.Fatalf("expected 0 filters, got %d", len(p.Filters))
	}
}

func TestNewPipelineSingleFilter(t *testing.T) {
	mf, cleanup := registerMock("testA")
	defer cleanup()

	p := NewPipeline("testA")
	if len(p.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(p.Filters))
	}
	if mf.initConfig != "" {
		t.Errorf("expected empty config, got %q", mf.initConfig)
	}
}

func TestNewPipelineTwoFilters(t *testing.T) {
	mfA, cleanupA := registerMock("testB")
	mfB, cleanupB := registerMock("testC")
	defer cleanupA()
	defer cleanupB()

	p := NewPipeline("testB:testC")
	if len(p.Filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(p.Filters))
	}
	if mfA.initConfig != "" {
		t.Errorf("testB: expected empty config, got %q", mfA.initConfig)
	}
	if mfB.initConfig != "" {
		t.Errorf("testC: expected empty config, got %q", mfB.initConfig)
	}
}

func TestNewPipelineSingleFilterWithConfig(t *testing.T) {
	mf, cleanup := registerMock("testD")
	defer cleanup()

	p := NewPipeline("testD:0.5")
	if len(p.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(p.Filters))
	}
	if mf.initConfig != "0.5" {
		t.Errorf("expected config %q, got %q", "0.5", mf.initConfig)
	}
}

func TestNewPipelineFilterWithConfigThenAnother(t *testing.T) {
	mfA, cleanupA := registerMock("testE")
	mfB, cleanupB := registerMock("testF")
	defer cleanupA()
	defer cleanupB()

	p := NewPipeline("testE:0.5:testF")
	if len(p.Filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(p.Filters))
	}
	if mfA.initConfig != "0.5" {
		t.Errorf("testE: expected config %q, got %q", "0.5", mfA.initConfig)
	}
	if mfB.initConfig != "" {
		t.Errorf("testF: expected empty config, got %q", mfB.initConfig)
	}
}

func TestNewPipelineMultipleConfigTokens(t *testing.T) {
	mf, cleanup := registerMock("testG")
	defer cleanup()

	// Two config tokens for a single filter: joined with ":"
	p := NewPipeline("testG:0.5:0.8")
	if len(p.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(p.Filters))
	}
	if mf.initConfig != "0.5:0.8" {
		t.Errorf("expected config %q, got %q", "0.5:0.8", mf.initConfig)
	}
}

func TestNewPipelineUnknownFilterLogged(t *testing.T) {
	// An unknown name at the start should not panic or add a filter.
	p := NewPipeline("nonexistent")
	if len(p.Filters) != 0 {
		t.Fatalf("expected 0 filters, got %d", len(p.Filters))
	}
}

func TestNewPipelineFilterOrder(t *testing.T) {
	mfA, cleanupA := registerMock("testH")
	mfB, cleanupB := registerMock("testI")
	defer cleanupA()
	defer cleanupB()

	p := NewPipeline("testH:testI")
	if p.Filters[0].Name() != "testH" {
		t.Errorf("expected first filter %q, got %q", "testH", p.Filters[0].Name())
	}
	if p.Filters[1].Name() != "testI" {
		t.Errorf("expected second filter %q, got %q", "testI", p.Filters[1].Name())
	}
}
