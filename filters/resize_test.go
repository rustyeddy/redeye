package filters

import (
	"bytes"
	"encoding/json"
	"image"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

// --- Init config parsing ---

func TestResizeInitDefaults(t *testing.T) {
	r := &Resize{}
	r.Init("")
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	if x != 1.0 || y != 1.0 {
		t.Errorf("empty config: want X=1.0 Y=1.0, got X=%v Y=%v", x, y)
	}
}

func TestResizeInitUniformScale(t *testing.T) {
	r := &Resize{}
	r.Init("0.5")
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	if x != 0.5 || y != 0.5 {
		t.Errorf("uniform scale: want X=0.5 Y=0.5, got X=%v Y=%v", x, y)
	}
}

func TestResizeInitXYScale(t *testing.T) {
	r := &Resize{}
	r.Init("0.5:0.75")
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	if x != 0.5 || y != 0.75 {
		t.Errorf("x:y scale: want X=0.5 Y=0.75, got X=%v Y=%v", x, y)
	}
}

func TestResizeInitInvalidScaleFallsBackToDefault(t *testing.T) {
	r := &Resize{}
	r.Init("notanumber")
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	// Bad parse: X and Y stay at the defaults set at the top of Init.
	if x != 1.0 || y != 1.0 {
		t.Errorf("invalid scale: want X=1.0 Y=1.0, got X=%v Y=%v", x, y)
	}
}

// --- Filter ---

func TestResizeFilterAppliesScale(t *testing.T) {
	m := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
	defer m.Close()

	r := &Resize{}
	r.Init("0.5")

	frame := &redeye.Frame{Mat: &m}
	result := r.Filter(frame)

	if result != frame {
		t.Error("Filter should return the same frame pointer")
	}

	got := m.Size()
	want := image.Point{X: 50, Y: 50}
	if got[0] != want.Y || got[1] != want.X {
		t.Errorf("after 0.5 scale: want 50×50, got %v", got)
	}
}

// --- ServeHTTP ---

func TestResizeServeHTTPBadJSONReturns400(t *testing.T) {
	r := &Resize{}
	r.Init("1.0")

	req := httptest.NewRequest(http.MethodPost, "/filters/resize", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("bad JSON: want 400, got %d", rr.Code)
	}
	// Fields must be unchanged.
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	if x != 1.0 || y != 1.0 {
		t.Errorf("fields should be unchanged after bad request; got X=%v Y=%v", x, y)
	}
}

func TestResizeServeHTTPValidJSONUpdatesFields(t *testing.T) {
	r := &Resize{}
	r.Init("1.0")

	body, _ := json.Marshal(map[string]float64{"x": 0.5, "y": 0.75})
	req := httptest.NewRequest(http.MethodPost, "/filters/resize", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("valid JSON: want 200, got %d", rr.Code)
	}
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	if x != 0.5 || y != 0.75 {
		t.Errorf("want X=0.5 Y=0.75, got X=%v Y=%v", x, y)
	}
}

func TestResizeServeHTTPResponseIsJSON(t *testing.T) {
	r := &Resize{}
	r.Init("1.0")

	body, _ := json.Marshal(map[string]float64{"x": 0.25, "y": 0.25})
	req := httptest.NewRequest(http.MethodPost, "/filters/resize", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var resp struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if resp.X != 0.25 || resp.Y != 0.25 {
		t.Errorf("response body: want x=0.25 y=0.25, got x=%v y=%v", resp.X, resp.Y)
	}
}

// --- Concurrency (most useful under `go test -race`) ---

func TestResizeFilterAndServeHTTPConcurrent(t *testing.T) {
	r := &Resize{}
	r.Init("1.0")

	m := gocv.NewMatWithSize(20, 20, gocv.MatTypeCV8UC3)
	defer m.Close()
	frame := &redeye.Frame{Mat: &m}

	var wg sync.WaitGroup

	// goroutine 1: read X/Y via Filter
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			r.Filter(frame)
		}
	}()

	// goroutine 2: write X/Y via ServeHTTP
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			body, _ := json.Marshal(map[string]float64{"x": 1.0, "y": 1.0})
			req := httptest.NewRequest(http.MethodPost, "/filters/resize", bytes.NewReader(body))
			r.ServeHTTP(httptest.NewRecorder(), req)
		}
	}()

	wg.Wait()
}
