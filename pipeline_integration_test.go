package redeye_test

// Integration tests for the video pipeline: build pipelines from real registered
// filter names, run frames through them, and assert observable transformations.
// Filters are registered via init() side-effects from the imports below.

import (
	"testing"

	"github.com/rustyeddy/redeye"
	_ "github.com/rustyeddy/redeye/filters/colors"
	_ "github.com/rustyeddy/redeye/filters/resize"
	_ "github.com/rustyeddy/redeye/filters/textzoom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocv.io/x/gocv"
)

// testFrame allocates a w×h BGR frame filled with a mid-gray value.
func testFrame(w, h int) *redeye.Frame {
	m := gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)
	return &redeye.Frame{Mat: &m}
}

func runPipeline(p *redeye.Pipeline, f *redeye.Frame) *redeye.Frame {
	for _, flt := range p.Filters {
		f = flt.Filter(f)
	}
	return f
}

// --- resize filter ---

func TestPipelineResizeHalvesDimensions(t *testing.T) {
	p, err := redeye.NewPipeline("resize:0.5")
	require.NoError(t, err)
	defer p.Close()

	f := testFrame(100, 100)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, 50, result.Cols(), "width after 0.5 scale")
	assert.Equal(t, 50, result.Rows(), "height after 0.5 scale")
}

func TestPipelineResizeIndependentXY(t *testing.T) {
	p, err := redeye.NewPipeline("resize:0.5:0.25")
	require.NoError(t, err)
	defer p.Close()

	f := testFrame(100, 100)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, 50, result.Cols(), "width after X=0.5")
	assert.Equal(t, 25, result.Rows(), "height after Y=0.25")
}

// --- color-detect (no-op passthrough) ---

func TestPipelineColorDetectPreservesDimensions(t *testing.T) {
	p, err := redeye.NewPipeline("color-detect")
	require.NoError(t, err)
	defer p.Close()

	f := testFrame(80, 60)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, 80, result.Cols())
	assert.Equal(t, 60, result.Rows())
}

// --- multi-filter pipeline ---

func TestPipelineTwoFiltersAppliedInOrder(t *testing.T) {
	// resize to half, then color-detect (no-op); final size must be 50x50.
	p, err := redeye.NewPipeline("resize:0.5:color-detect")
	require.NoError(t, err)
	defer p.Close()

	require.Len(t, p.Filters, 2, "should have two filters")
	assert.Equal(t, "resize", p.Filters[0].Name())
	assert.Equal(t, "color-detect", p.Filters[1].Name())

	f := testFrame(100, 100)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, 50, result.Cols())
	assert.Equal(t, 50, result.Rows())
}

// --- text-zoom filter ---

func TestPipelineTextZoomPreservesDimensions(t *testing.T) {
	p, err := redeye.NewPipeline("text-zoom")
	require.NoError(t, err)
	defer p.Close()

	const w, h = 160, 120
	f := testFrame(w, h)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, w, result.Cols(), "text-zoom must preserve width")
	assert.Equal(t, h, result.Rows(), "text-zoom must preserve height")
}

// --- unknown filter returns error ---

func TestPipelineUnknownFilterReturnsError(t *testing.T) {
	// "not-a-filter" is in filter-name position (no preceding filter), so it errors.
	_, err := redeye.NewPipeline("not-a-filter")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not-a-filter")
}

// --- empty pipeline is a no-op ---

func TestPipelineEmptyPassesFrameUnchanged(t *testing.T) {
	p, err := redeye.NewPipeline("")
	require.NoError(t, err)
	defer p.Close()

	f := testFrame(64, 48)
	defer f.Close()

	result := runPipeline(p, f)
	assert.Equal(t, 64, result.Cols())
	assert.Equal(t, 48, result.Rows())
}
