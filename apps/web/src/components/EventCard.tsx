import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventsApi, type Event, type Category } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'
import { useState } from 'react'

interface EventCardProps {
  event: Event
  categories: Category[]
  compact?: boolean
  onJoin?: () => void
  onLeave?: () => void
  onClose?: () => void
}

function getCategoryName(categories: Category[], id: number): string {
  const c = categories.find((x) => x.id === id)
  return c?.name ?? ''
}

function formatDate(s: string): string {
  const d = new Date(s)
  return d.toLocaleString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function EventCard({ event, categories, compact, onJoin, onLeave, onClose }: EventCardProps) {
  const queryClient = useQueryClient()
  const [error, setError] = useState('')
  const token = useAuthStore((s) => s.accessToken)
  const userId = token ? parseJwtUserId(token) : null
  const isCreator = userId === event.creator_id

  const joinMu = useMutation({
    mutationFn: () => eventsApi.join(event.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['my-events'] })
      onJoin?.()
    },
    onError: (err) => setError(err instanceof Error ? err.message : 'Ошибка'),
  })

  const leaveMu = useMutation({
    mutationFn: () => eventsApi.leave(event.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['my-events'] })
      onLeave?.()
    },
    onError: (err) => setError(err instanceof Error ? err.message : 'Ошибка'),
  })

  if (compact) {
    return (
      <div className="min-w-[200px]">
        <p className="font-medium text-stone-800">{event.title}</p>
        <p className="text-sm text-stone-500">{getCategoryName(categories, event.category_id)}</p>
        <p className="text-xs text-stone-400">{formatDate(event.start_time)}</p>
        <p className="text-sm text-stone-600 mt-1">
          {event.current_participants}/{event.max_participants} участников
        </p>
      </div>
    )
  }

  const canJoin = !isCreator && event.status === 'open' && event.current_participants < event.max_participants
  const showLeave = false

  return (
    <div className="bg-white rounded-xl shadow-lg border border-stone-200 p-4">
      <div className="flex justify-between items-start">
        <div>
          <h3 className="font-semibold text-stone-900">{event.title}</h3>
          <p className="text-sm text-stone-500">{getCategoryName(categories, event.category_id)}</p>
        </div>
        {onClose && (
          <button onClick={onClose} className="text-stone-400 hover:text-stone-600">
            ×
          </button>
        )}
      </div>

      {event.description && (
        <p className="text-sm text-stone-600 mt-2">{event.description}</p>
      )}

      <div className="mt-3 flex flex-wrap gap-2 text-sm text-stone-500">
        <span>{formatDate(event.start_time)}</span>
        <span>•</span>
        <span>
          {event.current_participants}/{event.max_participants} участников
        </span>
        {event.price > 0 && (
          <>
            <span>•</span>
            <span>{event.price} ₽</span>
          </>
        )}
        {event.requires_approval && (
          <>
            <span>•</span>
            <span className="text-amber-600">Нужно подтверждение</span>
          </>
        )}
      </div>

      {error && <p className="text-red-600 text-sm mt-2">{error}</p>}

      <div className="mt-4 flex gap-2">
        {isCreator && (
          <span className="text-sm text-huddle-600 font-medium">Вы создатель</span>
        )}
        {canJoin && (
          <button
            onClick={() => joinMu.mutate()}
            disabled={joinMu.isPending}
            className="px-4 py-1.5 rounded-lg bg-huddle-600 text-white text-sm font-medium hover:bg-huddle-700 disabled:opacity-50"
          >
            {joinMu.isPending ? '...' : 'Присоединиться'}
          </button>
        )}
        {!isCreator && !canJoin && showLeave && (
          <button
            onClick={() => leaveMu.mutate()}
            disabled={leaveMu.isPending}
            className="px-4 py-1.5 rounded-lg border border-stone-300 text-stone-600 text-sm hover:bg-stone-50 disabled:opacity-50"
          >
            {leaveMu.isPending ? '...' : 'Покинуть'}
          </button>
        )}
      </div>
    </div>
  )
}

function parseJwtUserId(token: string): string | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.sub ?? null
  } catch {
    return null
  }
}
