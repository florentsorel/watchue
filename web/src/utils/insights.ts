import { parseUtc } from "@/utils/time"

export interface Session {
  resource_id: string
  resource_type: "zone" | "room" | "light"
  name: string
  start: string
  end: string | null // null while the resource is still on
}

export interface Bucket {
  start: Date
  label: string
  onMs: number
  turnedOn: number
  turnedOff: number
}

export interface ResourceTotal {
  id: string
  name: string
  type: "zone" | "room" | "light"
  onMs: number
  changes: number
}

const HOUR_MS = 3_600_000

// Past this many days the window switches to weekly buckets — a bar per day
// beyond ~3 months is unreadable at any card width.
const MAX_DAILY_BUCKETS = 92

export type BucketSize = "day" | "week"

interface Interval {
  from: number
  to: number
}

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

// Local-midnight arithmetic, not `+ n * 86400000`: DST days are 23 or 25 hours
// long, so adding a fixed span drifts off midnight twice a year.
function addDays(d: Date, n: number): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate() + n)
}

// addDays normalises to midnight, which the day buckets want and a viewport
// panning by whole days does not — this keeps the time of day.
function shiftDays(d: Date, n: number): Date {
  const shifted = new Date(d)
  shifted.setDate(shifted.getDate() + n)
  return shifted
}

function overlapMs(a: Interval, b: Interval): number {
  return Math.max(0, Math.min(a.to, b.to) - Math.max(a.from, b.from))
}

function formatDayLabel(d: Date): string {
  return new Intl.DateTimeFormat("en-US", { month: "short", day: "numeric" }).format(d)
}

/** Session bounds in epoch ms, an open session running until `now`. */
export function toInterval(s: Session, now: number): Interval {
  return {
    from: parseUtc(s.start).getTime(),
    to: s.end ? parseUtc(s.end).getTime() : now,
  }
}

export function bucketSizeFor(days: number): BucketSize {
  return days <= MAX_DAILY_BUCKETS ? "day" : "week"
}

/** Local-midnight start of the `days`-long window ending today. */
export function windowStart(days: number, now: number): number {
  return addDays(startOfDay(new Date(now)), -(days - 1)).getTime()
}

/**
 * Splits the window into day or week buckets — oldest first, last one holding
 * `now` — and fills each with the on-time overlapping it plus the switch-ons
 * and switch-offs that landed inside it.
 *
 * Boundaries are local midnights, so a session running past midnight is split
 * across both days rather than credited entirely to the one it started in.
 */
export function bucketize(sessions: Session[], days: number, now = Date.now()): Bucket[] {
  const step = bucketSizeFor(days) === "week" ? 7 : 1
  const today = startOfDay(new Date(now))

  let first = addDays(today, -(days - 1))
  if (step === 7) {
    // Align to Monday so weekly bars stay comparable week to week.
    first = addDays(first, -((first.getDay() + 6) % 7))
  }

  const buckets: Bucket[] = []
  for (let start = first; start.getTime() <= now; start = addDays(start, step)) {
    buckets.push({ start, label: formatDayLabel(start), onMs: 0, turnedOn: 0, turnedOff: 0 })
  }

  for (const s of sessions) {
    const iv = toInterval(s, now)
    const endedAt = s.end ? parseUtc(s.end).getTime() : null
    for (const b of buckets) {
      const bounds = { from: b.start.getTime(), to: addDays(b.start, step).getTime() }
      b.onMs += overlapMs(iv, bounds)
      if (iv.from >= bounds.from && iv.from < bounds.to) b.turnedOn++
      if (endedAt !== null && endedAt >= bounds.from && endedAt < bounds.to) b.turnedOff++
    }
  }
  return buckets
}

/**
 * Share of each weekday/hour slot spent on, as a 7×24 grid of 0..1 — rows
 * Monday-first, columns 0h..23h, local time.
 *
 * Normalised per slot rather than left as a raw total: a 30-day window holds
 * four Mondays but five Tuesdays, so raw sums would make Tuesday look busier
 * by calendar accident alone.
 */
export function weeklyHeatmap(sessions: Session[], days: number, now = Date.now()): number[][] {
  const onMs: number[][] = Array.from({ length: 7 }, () => new Array<number>(24).fill(0))
  const slotMs: number[][] = Array.from({ length: 7 }, () => new Array<number>(24).fill(0))

  const from = windowStart(days, now)
  const window = { from, to: now }

  for (let day = new Date(from); day.getTime() <= now; day = addDays(day, 1)) {
    const row = (day.getDay() + 6) % 7
    for (let h = 0; h < 24; h++) {
      slotMs[row][h] += overlapMs(hourBounds(day, h), window)
    }
  }

  // Walked per session rather than per slot: a slot-major loop would rescan
  // every session 24 × `days` times, which at a year's window is minutes of work.
  for (const s of sessions) {
    const iv = toInterval(s, now)
    const start = Math.max(iv.from, from)
    for (let t = start; t < iv.to; ) {
      const cursor = new Date(t)
      const bounds = hourBounds(cursor, cursor.getHours())
      onMs[(cursor.getDay() + 6) % 7][cursor.getHours()] += overlapMs(iv, bounds)
      // The fall-back DST hour happens twice, so its reconstructed end can land
      // at or before the cursor — step a plain hour rather than spin forever.
      t = bounds.to > t ? bounds.to : t + HOUR_MS
    }
  }

  return onMs.map((row, r) => row.map((ms, h) => (slotMs[r][h] > 0 ? ms / slotMs[r][h] : 0)))
}

function hourBounds(day: Date, hour: number): Interval {
  return {
    from: new Date(day.getFullYear(), day.getMonth(), day.getDate(), hour).getTime(),
    to: new Date(day.getFullYear(), day.getMonth(), day.getDate(), hour + 1).getTime(),
  }
}

/** On-time and change count per resource, busiest first. */
export function perResource(sessions: Session[], days: number, now = Date.now()): ResourceTotal[] {
  const from = windowStart(days, now)
  const window = { from, to: now }
  const byId = new Map<string, ResourceTotal>()

  for (const s of sessions) {
    let total = byId.get(s.resource_id)
    if (!total) {
      total = { id: s.resource_id, name: s.name, type: s.resource_type, onMs: 0, changes: 0 }
      byId.set(s.resource_id, total)
    }
    const iv = toInterval(s, now)
    total.onMs += overlapMs(iv, window)
    if (iv.from >= from) total.changes++
    if (s.end && parseUtc(s.end).getTime() >= from) total.changes++
  }

  return [...byId.values()].sort((a, b) => b.onMs - a.onMs)
}

export interface NightPoint {
  /** Hours since the column's anchor, so the axis can place it. */
  value: number
  /** The switch itself, for the tooltip's real clock time. */
  at: number
  /** How many switches of this kind that night — the line only plots one. */
  count: number
}

export interface NightlySwitches {
  labels: string[]
  on: Array<NightPoint | null>
  off: Array<NightPoint | null>
}

/**
 * Start of the column an instant belongs to, given the hour a "day" opens at.
 *
 * With the default noon anchor a night is one column: a light switched off at
 * 01:17 lands in the evening it belongs to, above the 23:05 of the night
 * before, instead of dropping to the bottom of the next calendar column.
 */
export function anchoredDayStart(at: number, anchorHour: number): Date {
  const d = new Date(at)
  const base = new Date(d.getFullYear(), d.getMonth(), d.getDate(), anchorHour)
  return at < base.getTime() ? shiftDays(base, -1) : base
}

/**
 * One column per day, each holding the evening's first switch-on and its last
 * switch-off — the two that bound the lit stretch. Extra switches within a
 * column are counted but not plotted: the line is the night's shape, not every
 * flick of the switch.
 *
 * Days with no switch stay null so the line breaks there rather than drawing
 * straight through a day nothing happened.
 */
export function nightlySwitches(
  sessions: Session[],
  days: number,
  anchorHour: number,
  now = Date.now()
): NightlySwitches {
  const last = anchoredDayStart(now, anchorHour)
  const columns: Date[] = []
  for (let i = days - 1; i >= 0; i--) columns.push(shiftDays(last, -i))

  const indexOf = new Map<number, number>()
  columns.forEach((c, i) => indexOf.set(c.getTime(), i))

  const on: Array<NightPoint | null> = columns.map(() => null)
  const off: Array<NightPoint | null> = columns.map(() => null)

  function place(slot: Array<NightPoint | null>, at: number, keep: "first" | "last"): void {
    const column = anchoredDayStart(at, anchorHour)
    const i = indexOf.get(column.getTime())
    if (i === undefined) return
    const value = (at - column.getTime()) / HOUR_MS
    const existing = slot[i]
    if (!existing) {
      slot[i] = { value, at, count: 1 }
      return
    }
    existing.count++
    if (keep === "first" ? at < existing.at : at > existing.at) {
      existing.value = value
      existing.at = at
    }
  }

  for (const s of sessions) {
    place(on, parseUtc(s.start).getTime(), "first")
    if (s.end) place(off, parseUtc(s.end).getTime(), "last")
  }

  return { labels: columns.map(formatDayLabel), on, off }
}

/**
 * One tooltip row: value first, series name second — the reader already knows
 * which curve they are on and wants the number. Dated rather than just the
 * hour, since a switch past midnight falls on the next calendar day.
 */
export function formatSwitchTooltip(point: NightPoint, seriesLabel: string): string {
  const extra = point.count > 1 ? ` · ${point.count} that day` : ""
  return `${formatDateTime(point.at)} · ${seriesLabel}${extra}`
}

/** Clock label for a Y position, which counts hours from the column's anchor. */
export function formatAnchoredHour(value: number, anchorHour: number): string {
  return `${String(Math.round(anchorHour + value) % 24).padStart(2, "0")}h`
}

export function formatDateTime(ms: number): string {
  const d = new Date(ms)
  return `${formatDayLabel(d)} ${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
}

export function formatDuration(ms: number): string {
  if (ms < 60_000) return "0m"
  const hours = Math.floor(ms / HOUR_MS)
  const minutes = Math.round((ms % HOUR_MS) / 60_000)
  if (hours === 0) return `${minutes}m`
  if (hours < 10 && minutes > 0) return `${hours}h ${minutes}m`
  return `${hours}h`
}

export function toHours(ms: number): number {
  return Math.round((ms / HOUR_MS) * 100) / 100
}

export const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
