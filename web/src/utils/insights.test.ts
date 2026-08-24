import { describe, it, expect } from "vitest"
import {
  bucketize,
  bucketSizeFor,
  formatDuration,
  perResource,
  nightlySwitches,
  formatAnchoredHour,
  formatSwitchTooltip,
  weeklyHeatmap,
  type Session,
} from "@/utils/insights"

// The backend hands out UTC strings; every assertion here is about what the
// browser's local calendar makes of them, so tests pin an explicit `now`.
function session(start: string, end: string | null, name = "Salon"): Session {
  return { resource_id: name, resource_type: "room", name, start, end }
}

const NOW = new Date(2026, 7, 24, 12, 0, 0).getTime() // Aug 24 2026, 12:00 local

function utc(local: Date): string {
  return local.toISOString().slice(0, 19).replace("T", " ")
}

describe("bucketize", () => {
  it("produces one bucket per day, oldest first, ending today", () => {
    const buckets = bucketize([], 7, NOW)
    expect(buckets).toHaveLength(7)
    expect(buckets[0].start.getDate()).toBe(18)
    expect(buckets[6].start.getDate()).toBe(24)
  })

  it("splits a session across the midnight it spans", () => {
    // 23:00 -> 01:00 local, so one hour each side of midnight.
    const start = utc(new Date(2026, 7, 22, 23, 0))
    const end = utc(new Date(2026, 7, 23, 1, 0))
    const buckets = bucketize([session(start, end)], 7, NOW)
    const aug22 = buckets.find((b) => b.start.getDate() === 22)!
    const aug23 = buckets.find((b) => b.start.getDate() === 23)!
    expect(aug22.onMs).toBe(3_600_000)
    expect(aug23.onMs).toBe(3_600_000)
  })

  it("counts a turn-on and its turn-off in the buckets they landed in", () => {
    const start = utc(new Date(2026, 7, 22, 23, 0))
    const end = utc(new Date(2026, 7, 23, 1, 0))
    const buckets = bucketize([session(start, end)], 7, NOW)
    const aug22 = buckets.find((b) => b.start.getDate() === 22)!
    const aug23 = buckets.find((b) => b.start.getDate() === 23)!
    expect([aug22.turnedOn, aug22.turnedOff]).toEqual([1, 0])
    expect([aug23.turnedOn, aug23.turnedOff]).toEqual([0, 1])
  })

  it("runs an open session up to now and no further", () => {
    const start = utc(new Date(2026, 7, 24, 10, 0))
    const buckets = bucketize([session(start, null)], 7, NOW)
    const today = buckets[buckets.length - 1]
    expect(today.onMs).toBe(2 * 3_600_000)
    expect(today.turnedOff).toBe(0)
  })

  it("switches to Monday-aligned weekly buckets past the daily cap", () => {
    expect(bucketSizeFor(92)).toBe("day")
    expect(bucketSizeFor(93)).toBe("week")
    const buckets = bucketize([], 365, NOW)
    expect(buckets.length).toBeLessThan(60)
    for (const b of buckets) expect(b.start.getDay()).toBe(1) // Monday
  })
})

describe("weeklyHeatmap", () => {
  it("reports a fully-on slot as 1 and an untouched one as 0", () => {
    // Every Saturday 20:00-21:00 across the window would be needed for a true
    // 1.0, so use a one-day window holding a single Sunday.
    const now = new Date(2026, 7, 23, 23, 59, 59).getTime() // a Sunday
    const start = utc(new Date(2026, 7, 23, 20, 0))
    const end = utc(new Date(2026, 7, 23, 21, 0))
    const grid = weeklyHeatmap([session(start, end)], 1, now)
    expect(grid[6][20]).toBeCloseTo(1, 5) // Sunday is the last row
    expect(grid[6][19]).toBe(0)
    expect(grid[0][20]).toBe(0) // Monday never occurs in a one-day window
  })

  it("normalises by how often a slot actually occurred in the window", () => {
    // Two Sundays in a 14-day window, lit on only one of them.
    const now = new Date(2026, 7, 23, 23, 59, 59).getTime()
    const start = utc(new Date(2026, 7, 23, 20, 0))
    const end = utc(new Date(2026, 7, 23, 21, 0))
    const grid = weeklyHeatmap([session(start, end)], 14, now)
    expect(grid[6][20]).toBeCloseTo(0.5, 5)
  })
})

describe("perResource", () => {
  it("totals on-time and changes per resource, busiest first", () => {
    const sessions = [
      session(utc(new Date(2026, 7, 23, 10, 0)), utc(new Date(2026, 7, 23, 11, 0)), "Bureau"),
      session(utc(new Date(2026, 7, 23, 10, 0)), utc(new Date(2026, 7, 23, 14, 0)), "Salon"),
      session(utc(new Date(2026, 7, 24, 9, 0)), utc(new Date(2026, 7, 24, 10, 0)), "Salon"),
    ]
    const totals = perResource(sessions, 7, NOW)
    expect(totals.map((t) => t.name)).toEqual(["Salon", "Bureau"])
    expect(totals[0].onMs).toBe(5 * 3_600_000)
    expect(totals[0].changes).toBe(4) // two on/off pairs
    expect(totals[1].changes).toBe(2)
  })

  it("clips a session that started before the window", () => {
    // On for 3h, of which only the last hour falls inside a 1-day window.
    const now = new Date(2026, 7, 24, 1, 0).getTime()
    const start = utc(new Date(2026, 7, 23, 22, 0))
    const end = utc(new Date(2026, 7, 24, 1, 0))
    const totals = perResource([session(start, end)], 1, now)
    expect(totals[0].onMs).toBe(3_600_000)
    // The turn-on happened outside the window, so only the turn-off counts.
    expect(totals[0].changes).toBe(1)
  })
})

describe("formatDuration", () => {
  it.each([
    [0, "0m"],
    [59_000, "0m"],
    [90 * 60_000, "1h 30m"],
    [12 * 3_600_000, "12h"],
    [45 * 60_000, "45m"],
  ])("formats %ims as %s", (ms, want) => {
    expect(formatDuration(ms)).toBe(want)
  })
})

describe("nightlySwitches", () => {
  // 12:00 anchor: a column runs noon to noon, so a night sits whole inside one.
  const NIGHT = 12

  it("keeps a night's first switch-on and its last switch-off", () => {
    const sessions = [
      session(utc(new Date(2026, 7, 23, 18, 0)), utc(new Date(2026, 7, 23, 19, 0))),
      session(utc(new Date(2026, 7, 23, 21, 0)), utc(new Date(2026, 7, 24, 1, 17))),
    ]
    const { labels, on, off } = nightlySwitches(sessions, 3, NIGHT, NOW)
    expect(labels).toEqual(["Aug 22", "Aug 23", "Aug 24"])

    const night = on[1]!
    expect(new Date(night.at).getHours()).toBe(18) // the earlier of the two
    expect(night.count).toBe(2)
    expect(new Date(off[1]!.at).getHours()).toBe(1) // the later of the two
  })

  it("files an after-midnight switch-off under the evening it belongs to", () => {
    const sessions = [session(utc(new Date(2026, 7, 23, 21, 0)), utc(new Date(2026, 7, 24, 1, 17)))]
    const { off } = nightlySwitches(sessions, 3, NIGHT, NOW)
    expect(off[1]).not.toBeNull() // Aug 23's column, not Aug 24's
    expect(off[2]).toBeNull()
    // 01:17 is 13h17 past a noon anchor, so it plots above the same night's 21:00.
    expect(off[1]!.value).toBeCloseTo(13.28, 1)
    expect(off[1]!.value).toBeGreaterThan(9) // 21:00
  })

  it("files it under the next column when days start at midnight instead", () => {
    const sessions = [session(utc(new Date(2026, 7, 23, 21, 0)), utc(new Date(2026, 7, 24, 1, 17)))]
    const { off } = nightlySwitches(sessions, 3, 0, NOW)
    expect(off[1]).toBeNull()
    expect(off[2]!.value).toBeCloseTo(1.28, 1)
  })

  it("leaves a day with no switch null so the line breaks there", () => {
    const sessions = [session(utc(new Date(2026, 7, 22, 21, 0)), utc(new Date(2026, 7, 22, 23, 0)))]
    const { on, off } = nightlySwitches(sessions, 3, NIGHT, NOW)
    expect(on[0]).not.toBeNull()
    expect([on[1], on[2], off[1], off[2]]).toEqual([null, null, null, null])
  })

  it("ignores switches older than the requested number of days", () => {
    const sessions = [session(utc(new Date(2026, 6, 1, 21, 0)), utc(new Date(2026, 6, 1, 23, 0)))]
    const { on, off } = nightlySwitches(sessions, 3, NIGHT, NOW)
    expect(on.every((p) => p === null)).toBe(true)
    expect(off.every((p) => p === null)).toBe(true)
  })
})

describe("formatAnchoredHour", () => {
  it("reads back as a clock hour whatever the anchor", () => {
    expect(formatAnchoredHour(0, 12)).toBe("12h")
    expect(formatAnchoredHour(12, 12)).toBe("00h")
    expect(formatAnchoredHour(13.28, 12)).toBe("01h")
    expect(formatAnchoredHour(0, 0)).toBe("00h")
  })
})

describe("formatSwitchTooltip", () => {
  const at = new Date(2026, 7, 24, 1, 17).getTime()

  it("leads with the switch's own date and time, then the curve it belongs to", () => {
    expect(formatSwitchTooltip({ value: 13.28, at, count: 1 }, "Switched off")).toBe(
      "Aug 24 01:17 · Switched off"
    )
  })

  it("says how many switches that night when the line only plots one of them", () => {
    expect(formatSwitchTooltip({ value: 13.28, at, count: 4 }, "Switched off")).toBe(
      "Aug 24 01:17 · Switched off · 4 that day"
    )
  })
})
