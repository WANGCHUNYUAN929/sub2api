export const HEADER_OVERRIDE_ENABLED_KEY = 'header_override_enabled'
export const HEADER_OVERRIDES_KEY = 'header_overrides'

export const REQUEST_HEADERS_OVERRIDE_PLACEHOLDER =
  '{\n  "User-Agent": "claude-cli/2.1.196 (external, claude-vscode, agent-sdk/0.3.196)"\n}'

export type RequestHeadersOverrideError =
  | 'invalid_json'
  | 'must_be_object'
  | 'only_user_agent'
  | 'value_must_be_string'
  | 'value_required'

export type RequestHeadersOverrideParseResult =
  | { ok: true; headers: Record<string, string>; formatted: string }
  | { ok: false; error: RequestHeadersOverrideError }

const smartQuoteMap: Record<string, string> = {
  '\u201c': '"',
  '\u201d': '"',
  '\u201e': '"',
  '\u201f': '"',
  '\u301d': '"',
  '\u301e': '"',
  '\u301f': '"',
  '\uff02': '"'
}

export const normalizeJsonQuotes = (value: string): string =>
  value.replace(/[\u201c\u201d\u201e\u201f\u301d\u301e\u301f\uff02]/g, (char) => smartQuoteMap[char] || char)

export const canUseRequestHeadersOverride = (platform?: string, type?: string): boolean =>
  (platform === 'openai' || platform === 'anthropic') && type === 'apikey'

export const parseRequestHeadersOverrideInput = (value: string): RequestHeadersOverrideParseResult => {
  const input = normalizeJsonQuotes(value).trim()
  if (!input) {
    return { ok: true, headers: {}, formatted: '' }
  }

  let parsed: unknown
  try {
    parsed = JSON.parse(input)
  } catch {
    return { ok: false, error: 'invalid_json' }
  }

  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    return { ok: false, error: 'must_be_object' }
  }

  const out: Record<string, string> = {}
  for (const [key, value] of Object.entries(parsed as Record<string, unknown>)) {
    if (key.toLowerCase() !== 'user-agent') {
      return { ok: false, error: 'only_user_agent' }
    }
    if (typeof value !== 'string') {
      return { ok: false, error: 'value_must_be_string' }
    }
    const trimmed = value.trim()
    if (!trimmed) {
      return { ok: false, error: 'value_required' }
    }
    out['User-Agent'] = trimmed
  }

  return {
    ok: true,
    headers: out,
    formatted: Object.keys(out).length > 0 ? JSON.stringify(out, null, 2) : ''
  }
}

export const formatRequestHeadersOverride = (credentials?: Record<string, unknown>): string => {
  if (credentials?.[HEADER_OVERRIDE_ENABLED_KEY] !== true) {
    return ''
  }
  const raw = credentials?.[HEADER_OVERRIDES_KEY]
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return ''
  }
  const result = parseRequestHeadersOverrideInput(JSON.stringify(raw))
  return result.ok ? result.formatted : ''
}
