package hue_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/florentsorel/watchue/internal/hue"
)

// newTestClient spins up a TLS test server for mux and wraps it in a
// hue.Client, using the server's own client so the self-signed test
// certificate is trusted without relying on hue.Client's default
// InsecureSkipVerify behavior.
func newTestClient(t *testing.T, mux *http.ServeMux) *hue.Client {
	t.Helper()
	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	addr := strings.TrimPrefix(server.URL, "https://")
	return hue.NewClient(addr, "test-app-key", server.Client())
}

func TestClientLights(t *testing.T) {
	var gotKey string
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/light", func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("hue-application-key")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[],"data":[{
			"id":"light-1",
			"id_v1":"/lights/1",
			"owner":{"rid":"device-1","rtype":"device"},
			"metadata":{"name":"Salon"},
			"on":{"on":true}
		}]}`))
	})

	lights, err := newTestClient(t, mux).Lights(context.Background())
	if err != nil {
		t.Fatalf("Lights() error = %v", err)
	}
	if len(lights) != 1 {
		t.Fatalf("got %d lights, want 1", len(lights))
	}
	if got := lights[0]; got.ID != "light-1" || got.Metadata.Name != "Salon" || !got.On.On {
		t.Errorf("unexpected light: %+v", got)
	}
	if gotKey != "test-app-key" {
		t.Errorf("hue-application-key header = %q, want %q", gotKey, "test-app-key")
	}
}

func TestClientZones(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/zone", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[],"data":[{
			"id":"zone-1",
			"children":[{"rid":"light-1","rtype":"light"}],
			"services":[{"rid":"grouped-1","rtype":"grouped_light"}],
			"metadata":{"name":"Bureau"}
		}]}`))
	})

	zones, err := newTestClient(t, mux).Zones(context.Background())
	if err != nil {
		t.Fatalf("Zones() error = %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("got %d zones, want 1", len(zones))
	}
	got := zones[0]
	if got.Metadata.Name != "Bureau" || len(got.Children) != 1 || got.Children[0].RID != "light-1" {
		t.Errorf("unexpected zone: %+v", got)
	}
	if len(got.Services) != 1 || got.Services[0].RType != hue.ResourceGroupedLight {
		t.Errorf("unexpected zone services: %+v", got.Services)
	}
}

func TestClientRooms(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/room", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[],"data":[{
			"id":"room-1",
			"children":[{"rid":"device-1","rtype":"device"}],
			"services":[{"rid":"grouped-1","rtype":"grouped_light"}],
			"metadata":{"name":"Cuisine"}
		}]}`))
	})

	rooms, err := newTestClient(t, mux).Rooms(context.Background())
	if err != nil {
		t.Fatalf("Rooms() error = %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("got %d rooms, want 1", len(rooms))
	}
	got := rooms[0]
	if got.Metadata.Name != "Cuisine" || len(got.Children) != 1 || got.Children[0].RType != hue.ResourceDevice {
		t.Errorf("unexpected room: %+v", got)
	}
}

func TestClientGroupedLights(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/grouped_light", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[],"data":[{
			"id":"grouped-1",
			"owner":{"rid":"zone-1","rtype":"zone"},
			"on":{"on":false}
		}]}`))
	})

	groups, err := newTestClient(t, mux).GroupedLights(context.Background())
	if err != nil {
		t.Fatalf("GroupedLights() error = %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d grouped lights, want 1", len(groups))
	}
	if got := groups[0]; got.Owner.RID != "zone-1" || got.On.On {
		t.Errorf("unexpected grouped light: %+v", got)
	}
}

func TestClientLightsUnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/light", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	if _, err := newTestClient(t, mux).Lights(context.Background()); err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}

func TestClientLightsMalformedBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/clip/v2/resource/light", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})

	if _, err := newTestClient(t, mux).Lights(context.Background()); err == nil {
		t.Fatal("expected a decode error for a malformed response body")
	}
}
