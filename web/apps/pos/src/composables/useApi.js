
import { useSessionStore } from '../stores/session.js'

export function useApi() {
  const session = useSessionStore()
  let _refreshInFlight = null

  async function _refresh() {
    if (!_refreshInFlight) {
      _refreshInFlight = session.refresh().finally(() => { _refreshInFlight = null })
    }
    return _refreshInFlight
  }

  async function request(path, { method = 'GET', body, headers = {} } = {}) {
    const send = () => fetch(path, {
      method,
      headers: {
        ...(body != null ? { 'Content-Type': 'application/json' } : {}),
        ...(session.accessToken ? { Authorization: `Bearer ${session.accessToken}` } : {}),
        ...headers,
      },
      ...(body != null ? { body: JSON.stringify(body) } : {}),
    })

    let res = await send()

    if (res.status === 401 && navigator.onLine && session.refreshToken) {
      try { await _refresh(); res = await send() }
      catch { /* preserve the original 401 */ }
    }

    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      const err = new Error(data.error?.message ?? `HTTP ${res.status}`)
      err.status = res.status
      err.code   = data.error?.code
      throw err
    }
    return data
  }

  return {
    get:    (path, opts)       => request(path, { ...opts, method: 'GET'    }),
    post:   (path, body, opts) => request(path, { ...opts, method: 'POST',   body }),
    put:    (path, body, opts) => request(path, { ...opts, method: 'PUT',    body }),
    patch:  (path, body, opts) => request(path, { ...opts, method: 'PATCH',  body }),
    delete: (path, opts)       => request(path, { ...opts, method: 'DELETE' }),
    request,
  }
}

