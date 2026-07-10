package hue

import "context"

// ResourceType identifies the type of a Hue CLIP v2 resource.
type ResourceType string

const (
	ResourceLight        ResourceType = "light"
	ResourceRoom         ResourceType = "room"
	ResourceZone         ResourceType = "zone"
	ResourceDevice       ResourceType = "device"
	ResourceGroupedLight ResourceType = "grouped_light"
)

// ResourceIdentifier references another CLIP v2 resource by id and type.
type ResourceIdentifier struct {
	RID   string       `json:"rid"`
	RType ResourceType `json:"rtype"`
}

// Metadata holds the user-facing name and icon archetype of a resource.
type Metadata struct {
	Name      string `json:"name"`
	Archetype string `json:"archetype,omitempty"`
}

// OnState is the on/off state shared by lights and grouped lights.
type OnState struct {
	On bool `json:"on"`
}

// Light is a single Hue bulb.
type Light struct {
	ID       string             `json:"id"`
	IDV1     string             `json:"id_v1,omitempty"`
	Owner    ResourceIdentifier `json:"owner"`
	Metadata Metadata           `json:"metadata"`
	On       OnState            `json:"on"`
}

// Zone groups a set of lights (across rooms) under one name. Its aggregate
// on/off state is reported separately, via the GroupedLight referenced in
// Services.
type Zone struct {
	ID       string               `json:"id"`
	IDV1     string               `json:"id_v1,omitempty"`
	Children []ResourceIdentifier `json:"children"` // the lights belonging to this zone
	Services []ResourceIdentifier `json:"services"` // includes the zone's grouped_light
	Metadata Metadata             `json:"metadata"`
}

// Room groups the devices installed in a physical space. Unlike Zone, whose
// Children reference lights directly, a Room's Children reference devices —
// resolve a room's lights by matching Light.Owner against those device ids.
type Room struct {
	ID       string               `json:"id"`
	IDV1     string               `json:"id_v1,omitempty"`
	Children []ResourceIdentifier `json:"children"` // the devices installed in this room
	Services []ResourceIdentifier `json:"services"` // includes the room's grouped_light
	Metadata Metadata             `json:"metadata"`
}

// GroupedLight is the on/off/dimming state of a zone (or room) as a whole.
type GroupedLight struct {
	ID    string             `json:"id"`
	IDV1  string             `json:"id_v1,omitempty"`
	Owner ResourceIdentifier `json:"owner"` // the zone/room this belongs to
	On    OnState            `json:"on"`
}

// response is the envelope the CLIP v2 API wraps every resource list in.
type response[T any] struct {
	Errors []struct {
		Description string `json:"description"`
	} `json:"errors"`
	Data []T `json:"data"`
}

// Lights returns every light known to the bridge.
func (c *Client) Lights(ctx context.Context) ([]Light, error) {
	var res response[Light]
	if err := c.get(ctx, "/clip/v2/resource/light", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// Zones returns every zone known to the bridge.
func (c *Client) Zones(ctx context.Context) ([]Zone, error) {
	var res response[Zone]
	if err := c.get(ctx, "/clip/v2/resource/zone", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// Rooms returns every room known to the bridge.
func (c *Client) Rooms(ctx context.Context) ([]Room, error) {
	var res response[Room]
	if err := c.get(ctx, "/clip/v2/resource/room", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// GroupedLights returns the aggregate on/off state for every zone/room.
func (c *Client) GroupedLights(ctx context.Context) ([]GroupedLight, error) {
	var res response[GroupedLight]
	if err := c.get(ctx, "/clip/v2/resource/grouped_light", &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
