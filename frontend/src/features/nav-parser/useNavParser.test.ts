import { describe, expect, it } from 'vitest'
import { useNavParser } from './useNavParser'

describe('useNavParser', () => {
  it('starts in an idle state while retaining the manual launch API', () => {
    const parser = useNavParser()

    expect(parser.isNavParsing.value).toBe(false)
    expect(parser.navParserProgress.value.stage).toBe('Ожидание')
    expect(typeof parser.runNavParser).toBe('function')
    expect(parser.navParserSourceLabel('manual')).toBe('Ручной запуск')
    expect(parser.navParserSourceLabel('scheduled')).toBe('По расписанию')
  })

  it('presents progress log entries newest first without mutating API data', () => {
    const parser = useNavParser()
    parser.navParserProgress.value.logs = [
      { time: '2026-07-21T10:00:00Z', level: 'info', message: 'Первый' },
      { time: '2026-07-21T10:01:00Z', level: 'info', message: 'Второй' },
    ]

    expect(parser.navParserLogsNewestFirst.value.map((entry) => entry.message)).toEqual(['Второй', 'Первый'])
    expect(parser.navParserProgress.value.logs.map((entry) => entry.message)).toEqual(['Первый', 'Второй'])
  })
})
