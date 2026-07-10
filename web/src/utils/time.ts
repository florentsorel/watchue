// SQLite's CURRENT_TIMESTAMP is UTC but has no offset/'Z' suffix, so append
// one before parsing or the browser reads it as local time.
export function parseUtc(sqliteTimestamp: string): Date {
  return new Date(sqliteTimestamp.replace(" ", "T") + "Z")
}

export function formatAbsoluteTime(date: Date): string {
  const datePart = new Intl.DateTimeFormat("en-US", {
    month: "2-digit",
    day: "2-digit",
    year: "2-digit",
  }).format(date)
  const timePart = new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).format(date)
  return `${datePart} ${timePart}`
}

export function formatRelativeTime(sqliteTimestamp: string): string {
  const date = parseUtc(sqliteTimestamp)
  const minutes = Math.floor((Date.now() - date.getTime()) / 60000)
  if (minutes < 1) return "just now"
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return formatAbsoluteTime(date)
}
