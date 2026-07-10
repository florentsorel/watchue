// Package catalog assembles the bridge's lights, zones and rooms into the
// nested shape the web client renders for resource selection.
package catalog

import (
	"context"

	"github.com/florentsorel/watchue/internal/hue"
)

// HueClient is the subset of hue.Client operations Build needs.
type HueClient interface {
	Lights(ctx context.Context) ([]hue.Light, error)
	Zones(ctx context.Context) ([]hue.Zone, error)
	Rooms(ctx context.Context) ([]hue.Room, error)
	GroupedLights(ctx context.Context) ([]hue.GroupedLight, error)
}

// Light is one bulb as exposed to the client, nested under its zone/room.
type Light struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	On        bool   `json:"on"`
	Archetype string `json:"archetype"`
}

// Group is a zone or room together with its own on/off state and lights.
type Group struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	On        bool    `json:"on"`
	Archetype string  `json:"archetype"`
	Lights    []Light `json:"lights"`
}

// Catalog is the bridge's resources grouped for display/selection.
type Catalog struct {
	Zones []Group `json:"zones"`
	Rooms []Group `json:"rooms"`
}

// Build fetches lights, zones, rooms and grouped-light states from client
// and assembles them into a Catalog.
func Build(ctx context.Context, client HueClient) (Catalog, error) {
	lights, err := client.Lights(ctx)
	if err != nil {
		return Catalog{}, err
	}
	zones, err := client.Zones(ctx)
	if err != nil {
		return Catalog{}, err
	}
	rooms, err := client.Rooms(ctx)
	if err != nil {
		return Catalog{}, err
	}
	groupedLights, err := client.GroupedLights(ctx)
	if err != nil {
		return Catalog{}, err
	}

	lightByID := make(map[string]hue.Light, len(lights))
	for _, l := range lights {
		lightByID[l.ID] = l
	}
	onByGroupedLightID := make(map[string]bool, len(groupedLights))
	for _, g := range groupedLights {
		onByGroupedLightID[g.ID] = g.On.On
	}

	// zone/room on/off state lives in the grouped_light it references
	groupOn := func(services []hue.ResourceIdentifier) bool {
		for _, s := range services {
			if s.RType == hue.ResourceGroupedLight {
				return onByGroupedLightID[s.RID]
			}
		}
		return false
	}
	toLight := func(l hue.Light) Light {
		return Light{ID: l.ID, Name: l.Metadata.Name, On: l.On.On, Archetype: l.Metadata.Archetype}
	}

	zoneGroups := make([]Group, 0, len(zones))
	for _, z := range zones {
		var zoneLights []Light
		for _, child := range z.Children {
			if child.RType != hue.ResourceLight {
				continue
			}
			if l, ok := lightByID[child.RID]; ok {
				zoneLights = append(zoneLights, toLight(l))
			}
		}
		zoneGroups = append(zoneGroups, Group{
			ID:        z.ID,
			Name:      z.Metadata.Name,
			On:        groupOn(z.Services),
			Archetype: z.Metadata.Archetype,
			Lights:    zoneLights,
		})
	}

	roomGroups := make([]Group, 0, len(rooms))
	for _, r := range rooms {
		deviceIDs := make(map[string]struct{}, len(r.Children))
		for _, child := range r.Children {
			if child.RType == hue.ResourceDevice {
				deviceIDs[child.RID] = struct{}{}
			}
		}
		var roomLights []Light
		for _, l := range lights {
			if l.Owner.RType != hue.ResourceDevice {
				continue
			}
			if _, ok := deviceIDs[l.Owner.RID]; ok {
				roomLights = append(roomLights, toLight(l))
			}
		}
		roomGroups = append(roomGroups, Group{
			ID:        r.ID,
			Name:      r.Metadata.Name,
			On:        groupOn(r.Services),
			Archetype: r.Metadata.Archetype,
			Lights:    roomLights,
		})
	}

	return Catalog{Zones: zoneGroups, Rooms: roomGroups}, nil
}

// Resolve looks up id among the bridge's zones, rooms, and lights, returning
// its type and current name. Checks Rooms for lights, not Zones: every light
// belongs to exactly one room, while zone membership is only a subset.
func (c Catalog) Resolve(id string) (resourceType, name string, ok bool) {
	for _, z := range c.Zones {
		if z.ID == id {
			return string(hue.ResourceZone), z.Name, true
		}
	}
	for _, r := range c.Rooms {
		if r.ID == id {
			return string(hue.ResourceRoom), r.Name, true
		}
		for _, l := range r.Lights {
			if l.ID == id {
				return string(hue.ResourceLight), l.Name, true
			}
		}
	}
	return "", "", false
}
