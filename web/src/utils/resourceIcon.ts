// Hue's own CLIP v2 `metadata.archetype` values, mapped to SVG sprite symbols
// (see IconSprite.vue, `i-hue-*`) sourced from arallsopp/hass-hue-icons. Not
// every archetype has a match — unmapped ones fall back to the generic guess
// below.
const LIGHT_ARCHETYPE_ICONS: Record<string, string> = {
  classic_bulb: "hue-bulb-classic",
  sultan_bulb: "hue-bulb-sultan",
  spot_bulb: "hue-bulb-spot",
  flood_bulb: "hue-bulb-flood",
  candle_bulb: "hue-bulb-candle",
  hue_lightstrip: "hue-lightstrip",
  hue_iris: "hue-iris",
  hue_bloom: "hue-bloom",
  hue_go: "hue-go",
  hue_play: "hue-play-bar",
  pendant_round: "hue-pendant-round",
  pendant_long: "hue-pendant-long",
  ceiling_round: "hue-ceiling-round",
  ceiling_square: "hue-ceiling-square",
  floor_shade: "hue-floor-shade",
  floor_lantern: "hue-floor-lantern",
  floor_spot: "hue-floor-spot",
  table_shade: "hue-table-shade",
  table_wash: "hue-table-wash",
  recessed_ceiling: "hue-recessed-ceiling",
  recessed_floor: "hue-recessed-floor",
  single_spot: "hue-single-spot",
  double_spot: "hue-double-spot",
  wall_lantern: "hue-wall-lantern",
  wall_shade: "hue-wall-shade",
  wall_spot: "hue-wall-spot",
  bollard: "hue-bollard",
  desk_lamp: "hue-desk-lamp",
}

// Hue's room/zone `metadata.archetype` enum (identical for both resource
// types). A handful of values (garden, top_floor, man_cave, music, tv,
// reading) have no matching icon in the source set and fall back to the
// name guess below.
const ROOM_ARCHETYPE_ICONS: Record<string, string> = {
  living_room: "hue-room-living",
  kitchen: "hue-room-kitchen",
  dining: "hue-room-dining",
  bedroom: "hue-room-bedroom",
  kids_bedroom: "hue-room-kids",
  bathroom: "hue-room-bathroom",
  nursery: "hue-room-nursery",
  recreation: "hue-room-recreation",
  office: "hue-room-office",
  gym: "hue-room-gym",
  hallway: "hue-room-hallway",
  toilet: "hue-room-toilet",
  front_door: "hue-room-front-door",
  garage: "hue-room-garage",
  terrace: "hue-room-terrace",
  driveway: "hue-room-driveway",
  carport: "hue-room-carport",
  home: "hue-home",
  downstairs: "hue-downstairs",
  upstairs: "hue-upstairs",
  attic: "hue-room-attic",
  guest_room: "hue-room-guestroom",
  staircase: "hue-room-stairs",
  lounge: "hue-room-lounge",
  computer: "hue-room-computer",
  studio: "hue-room-studio",
  closet: "hue-room-closet",
  storage: "hue-room-storage",
  laundry_room: "hue-room-laundry",
  balcony: "hue-room-balcony",
  porch: "hue-room-porch",
  barbecue: "hue-room-bbq",
  pool: "hue-room-pool",
  other: "hue-room-other",
}

// Falls back to a keyword guess when the archetype is missing or unmapped.
const ROOM_KEYWORDS: Array<[string, string]> = [
  ["bedroom", "room-bedroom"],
  ["kitchen", "room-kitchen"],
  ["office", "room-office"],
  ["bath", "room-bathroom"],
  ["hall", "room-hallway"],
  ["living", "room-living"],
]

function guessRoomIconFromName(name: string): string {
  const lower = name.toLowerCase()
  for (const [keyword, icon] of ROOM_KEYWORDS) {
    if (lower.includes(keyword)) return icon
  }
  return "room-living"
}

export function roomIcon(archetype: string | undefined, name: string): string {
  if (archetype && ROOM_ARCHETYPE_ICONS[archetype]) return ROOM_ARCHETYPE_ICONS[archetype]
  return guessRoomIconFromName(name)
}

// Zones use the shared archetype map too, but fall back to the generic zone
// symbol instead of a room guess — a zone's name is less likely to hint at a
// single room type.
export function zoneIcon(archetype: string | undefined): string {
  if (archetype && ROOM_ARCHETYPE_ICONS[archetype]) return ROOM_ARCHETYPE_ICONS[archetype]
  return "zone"
}

export function lightIcon(archetype: string | undefined, name: string): string {
  if (archetype && LIGHT_ARCHETYPE_ICONS[archetype]) return LIGHT_ARCHETYPE_ICONS[archetype]

  const lower = name.toLowerCase()
  if (lower.includes("strip")) return "lightstrip"
  if (lower.includes("spot") || lower.includes("ceiling")) return "spot"
  return "bulb"
}
