const API_BASE = '/api/v1'

function getAuthHeaders(): HeadersInit {
  const token = localStorage.getItem('huddle_tokens')
  if (!token) return {}
  try {
    const parsed = JSON.parse(token)
    const access = parsed?.state?.accessToken
    if (access) return { Authorization: `Bearer ${access}` }
  } catch {
    // ignore
  }
  return {}
}

async function handleResponse<T>(res: Response): Promise<T> {
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    const err = (data as { error?: string })?.error || res.statusText
    throw new Error(err)
  }
  return data as T
}

// Auth
export const authApi = {
  login: async (email: string, password: string) => {
    const res = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })
    const data = await handleResponse<{ tokens: { access_token: string; refresh_token: string; expires_in: number; token_type: string } }>(res)
    return data.tokens
  },

  register: async (payload: { email: string; password: string; firstName: string; lastName: string }) => {
    const res = await fetch(`${API_BASE}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    return handleResponse(res)
  },

  refreshToken: async (refreshToken: string) => {
    const res = await fetch(`${API_BASE}/auth/refresh-token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    const data = await handleResponse<{ tokens: { access_token: string; refresh_token: string; expires_in: number; token_type: string } }>(res)
    return data.tokens
  },
}

// Categories
export interface Category {
  id: number
  parent_id?: number
  name: string
  slug: string
  icon_url: string
  color_code: string
}

export const categoriesApi = {
  list: async (): Promise<Category[]> => {
    const res = await fetch(`${API_BASE}/categories`, { headers: getAuthHeaders() })
    return handleResponse(res)
  },
}

// Events
export interface Event {
  id: string
  creator_id: string
  category_id: number
  title: string
  description: string
  lat: number
  lon: number
  start_time: string
  max_participants: number
  current_participants: number
  price: number
  requires_approval: boolean
  status: string
  created_at: string
}

export interface EventParticipant {
  event_id: string
  user_id: string
  status: 'pending' | 'accepted' | 'rejected'
  joined_at: string
}

export const eventsApi = {
  list: async (params: { lat?: number; lon?: number; radius?: number; category?: string }) => {
    const q = new URLSearchParams()
    if (params.lat != null) q.set('lat', String(params.lat))
    if (params.lon != null) q.set('lon', String(params.lon))
    if (params.radius != null) q.set('radius', String(params.radius))
    if (params.category) q.set('category', params.category)
    const url = `${API_BASE}/events${q.toString() ? '?' + q : ''}`
    const res = await fetch(url, { headers: getAuthHeaders() })
    return handleResponse<Event[]>(res)
  },

  get: async (id: string) => {
    const res = await fetch(`${API_BASE}/events/${id}`, { headers: getAuthHeaders() })
    return handleResponse<Event>(res)
  },

  getParticipants: async (id: string) => {
    const res = await fetch(`${API_BASE}/events/${id}/participants`, { headers: getAuthHeaders() })
    return handleResponse<EventParticipant[]>(res)
  },

  create: async (payload: {
    category_id: number
    title: string
    description: string
    lat: number
    lon: number
    start_time: string
    max_participants: number
    price?: number
    requires_approval: boolean
  }) => {
    const res = await fetch(`${API_BASE}/events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify(payload),
    })
    return handleResponse<Event>(res)
  },

  join: async (id: string) => {
    const res = await fetch(`${API_BASE}/events/${id}/participants`, {
      method: 'POST',
      headers: getAuthHeaders(),
    })
    return handleResponse(res)
  },

  leave: async (id: string) => {
    const res = await fetch(`${API_BASE}/events/${id}/participants`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    })
    if (res.status === 204) return
    return handleResponse(res)
  },

  approveParticipant: async (eventId: string, userId: string, status: 'accepted' | 'rejected') => {
    const res = await fetch(`${API_BASE}/events/${eventId}/participants/${userId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
      body: JSON.stringify({ status }),
    })
    return handleResponse(res)
  },

  delete: async (id: string) => {
    const res = await fetch(`${API_BASE}/events/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    })
    if (res.status === 204) return
    return handleResponse(res)
  },

  myEvents: async () => {
    const res = await fetch(`${API_BASE}/my-events`, { headers: getAuthHeaders() })
    return handleResponse<Event[]>(res)
  },
}
