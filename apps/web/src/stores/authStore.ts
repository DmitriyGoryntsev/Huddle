import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const TOKEN_KEY = 'huddle_tokens'

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

interface AuthState {
  accessToken: string | null
  refreshToken: string | null
  expiresAt: number | null
  setTokens: (tokens: TokenPair) => void
  clearAuth: () => void
  getAccessToken: () => string | null
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      expiresAt: null,

      setTokens: (tokens) => {
        const expiresAt = Date.now() + tokens.expires_in * 1000
        set({
          accessToken: tokens.access_token,
          refreshToken: tokens.refresh_token,
          expiresAt,
        })
      },

      clearAuth: () => set({ accessToken: null, refreshToken: null, expiresAt: null }),

      getAccessToken: () => get().accessToken,
    }),
    {
      name: TOKEN_KEY,
      partialize: (s) => ({
        accessToken: s.accessToken,
        refreshToken: s.refreshToken,
        expiresAt: s.expiresAt,
      }),
    },
  ),
)
