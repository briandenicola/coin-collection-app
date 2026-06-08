/**
 * Parse a newline-delimited list of options into an array.
 * - Trim whitespace from each line
 * - Drop blank lines
 * - De-duplicate while preserving order
 * - Fallback to defaults if input is empty/invalid
 */
export function parseOptionList(value: string | null | undefined, defaults: string[]): string[] {
  if (!value || value.trim() === '') {
    return [...defaults]
  }

  const lines = value.split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0)

  // De-duplicate using a Set, preserving first occurrence order
  const seen = new Set<string>()
  const unique: string[] = []
  for (const line of lines) {
    if (!seen.has(line)) {
      seen.add(line)
      unique.push(line)
    }
  }

  return unique.length > 0 ? unique : [...defaults]
}

/**
 * Convert an array of options back to a newline-delimited string.
 */
export function formatOptionList(options: string[]): string {
  return options.join('\n')
}
