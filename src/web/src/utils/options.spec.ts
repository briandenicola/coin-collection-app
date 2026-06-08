import { describe, it, expect } from 'vitest'
import { parseOptionList, formatOptionList } from './options'

describe('parseOptionList', () => {
  const DEFAULT_CATEGORIES = ['Roman', 'Greek', 'Byzantine', 'Modern', 'Other']
  const DEFAULT_ERAS = ['ancient', 'medieval', 'modern']

  describe('default behavior', () => {
    it('returns defaults when value is null', () => {
      expect(parseOptionList(null, DEFAULT_CATEGORIES)).toEqual(DEFAULT_CATEGORIES)
    })

    it('returns defaults when value is undefined', () => {
      expect(parseOptionList(undefined, DEFAULT_CATEGORIES)).toEqual(DEFAULT_CATEGORIES)
    })

    it('returns defaults when value is empty string', () => {
      expect(parseOptionList('', DEFAULT_CATEGORIES)).toEqual(DEFAULT_CATEGORIES)
    })

    it('returns defaults when value is only whitespace', () => {
      expect(parseOptionList('   \n  \t  ', DEFAULT_CATEGORIES)).toEqual(DEFAULT_CATEGORIES)
    })

    it('returns a copy of defaults, not a reference', () => {
      const result = parseOptionList(null, DEFAULT_CATEGORIES)
      result.push('Modified')
      expect(DEFAULT_CATEGORIES).not.toContain('Modified')
    })
  })

  describe('parsing behavior', () => {
    it('parses newline-delimited list', () => {
      const input = 'Imperial\nRepublican\nProvincial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('trims whitespace from each line', () => {
      const input = '  Imperial  \n Republican\n\tProvincial\t'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('drops blank lines', () => {
      const input = 'Imperial\n\n\nRepublican\n\nProvincial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('drops lines with only whitespace', () => {
      const input = 'Imperial\n   \nRepublican\n\t\t\nProvincial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('handles mixed line endings (CRLF)', () => {
      const input = 'Imperial\r\nRepublican\r\nProvincial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('handles single value without newline', () => {
      expect(parseOptionList('OnlyOne', DEFAULT_CATEGORIES)).toEqual(['OnlyOne'])
    })
  })

  describe('deduplication', () => {
    it('removes duplicate values', () => {
      const input = 'Imperial\nRepublican\nImperial\nProvincial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial', 'Republican', 'Provincial'])
    })

    it('preserves first occurrence order', () => {
      const input = 'C\nA\nB\nA\nC'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['C', 'A', 'B'])
    })

    it('treats whitespace variations as duplicates after trimming', () => {
      const input = 'Imperial\n  Imperial  \n\tImperial'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Imperial'])
    })
  })

  describe('edge cases', () => {
    it('handles very long lists', () => {
      const longList = Array.from({ length: 100 }, (_, i) => `Option${i}`).join('\n')
      const result = parseOptionList(longList, DEFAULT_CATEGORIES)
      expect(result).toHaveLength(100)
      expect(result[0]).toBe('Option0')
      expect(result[99]).toBe('Option99')
    })

    it('handles values with special characters', () => {
      const input = 'Roman (Imperial)\nGreek & Hellenistic\nByzantine/Medieval'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual([
        'Roman (Imperial)',
        'Greek & Hellenistic',
        'Byzantine/Medieval',
      ])
    })

    it('handles Unicode characters', () => {
      const input = 'Römisch\n中国\nΕλληνικά'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Römisch', '中国', 'Ελληνικά'])
    })

    it('handles numeric strings', () => {
      const input = '100\n200\n300'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['100', '200', '300'])
    })
  })

  describe('real-world scenarios', () => {
    it('parses default category list', () => {
      const input = 'Roman\nGreek\nByzantine\nModern\nOther'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(DEFAULT_CATEGORIES)
    })

    it('parses default era list', () => {
      const input = 'ancient\nmedieval\nmodern'
      expect(parseOptionList(input, DEFAULT_ERAS)).toEqual(DEFAULT_ERAS)
    })

    it('parses custom category list from admin UI', () => {
      const input = 'Imperial Roman\nRepublican Roman\nGreek\nProvincial\nOther'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual([
        'Imperial Roman',
        'Republican Roman',
        'Greek',
        'Provincial',
        'Other',
      ])
    })

    it('handles messy user input from textarea', () => {
      const input = '\n  Roman  \n\n\n  Greek\t\n   \n  Byzantine\n\n'
      expect(parseOptionList(input, DEFAULT_CATEGORIES)).toEqual(['Roman', 'Greek', 'Byzantine'])
    })
  })
})

describe('formatOptionList', () => {
  it('converts array to newline-delimited string', () => {
    expect(formatOptionList(['A', 'B', 'C'])).toBe('A\nB\nC')
  })

  it('handles empty array', () => {
    expect(formatOptionList([])).toBe('')
  })

  it('handles single item', () => {
    expect(formatOptionList(['OnlyOne'])).toBe('OnlyOne')
  })

  it('preserves order', () => {
    expect(formatOptionList(['C', 'A', 'B'])).toBe('C\nA\nB')
  })

  it('handles values with special characters', () => {
    expect(formatOptionList(['Roman (Imperial)', 'Greek & Hellenistic'])).toBe(
      'Roman (Imperial)\nGreek & Hellenistic',
    )
  })
})

describe('parseOptionList and formatOptionList roundtrip', () => {
  it('roundtrips default categories', () => {
    const defaults = ['Roman', 'Greek', 'Byzantine', 'Modern', 'Other']
    const formatted = formatOptionList(defaults)
    const parsed = parseOptionList(formatted, defaults)
    expect(parsed).toEqual(defaults)
  })

  it('roundtrips custom list', () => {
    const custom = ['A', 'B', 'C']
    const formatted = formatOptionList(custom)
    const parsed = parseOptionList(formatted, custom)
    expect(parsed).toEqual(custom)
  })

  it('roundtrips with Unicode', () => {
    const unicode = ['Römisch', '中国', 'Ελληνικά']
    const formatted = formatOptionList(unicode)
    const parsed = parseOptionList(formatted, unicode)
    expect(parsed).toEqual(unicode)
  })
})
