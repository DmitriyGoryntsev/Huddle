import { useQuery } from '@tanstack/react-query'
import { eventsApi, categoriesApi, type Event, type Category } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'
import { useMutation, useQueryClient } from '@tanstack/react-query'

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

export default function MyEventsPage() {
  const queryClient = useQueryClient()
  const userId = useAuthStore((s) => {
    const t = s.accessToken
    if (!t) return null
    try {
      const p = JSON.parse(atob(t.split('.')[1]))
      return p.sub ?? null
    } catch {
      return null
    }
  })

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: categoriesApi.list,
  })

  const { data: events = [], isLoading } = useQuery({
    queryKey: ['my-events'],
    queryFn: eventsApi.myEvents,
  })

  const leaveMu = useMutation({
    mutationFn: (id: string) => eventsApi.leave(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-events'] })
      queryClient.invalidateQueries({ queryKey: ['events'] })
    },
  })

  const deleteMu = useMutation({
    mutationFn: (id: string) => eventsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-events'] })
      queryClient.invalidateQueries({ queryKey: ['events'] })
    },
  })

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto p-4">
        <p className="text-stone-500">Загрузка...</p>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto p-4">
      <h1 className="text-2xl font-bold text-stone-900 mb-6">Мои события</h1>

      {events.length === 0 ? (
        <p className="text-stone-500">Вы пока не участвуете ни в одном событии. Создайте своё или присоединяйтесь к другим на карте!</p>
      ) : (
        <div className="space-y-4">
          {events.map((ev) => {
            const isCreator = userId === ev.creator_id
            return (
              <div
                key={ev.id}
                className="bg-white rounded-xl border border-stone-200 p-4 shadow-sm"
              >
                <div className="flex justify-between items-start">
                  <div>
                    <h2 className="font-semibold text-stone-900">{ev.title}</h2>
                    <p className="text-sm text-stone-500">{getCategoryName(categories, ev.category_id)}</p>
                    <p className="text-sm text-stone-600 mt-1">
                      {formatDate(ev.start_time)} • {ev.current_participants}/{ev.max_participants} участников
                    </p>
                  </div>
                  <div className="flex gap-2">
                    {!isCreator && (
                      <button
                        onClick={() => leaveMu.mutate(ev.id)}
                        disabled={leaveMu.isPending}
                        className="px-3 py-1 rounded-lg border border-stone-300 text-stone-600 text-sm hover:bg-stone-50"
                      >
                        Покинуть
                      </button>
                    )}
                    {isCreator && (
                      <button
                        onClick={() => deleteMu.mutate(ev.id)}
                        disabled={deleteMu.isPending}
                        className="px-3 py-1 rounded-lg border border-red-200 text-red-600 text-sm hover:bg-red-50"
                      >
                        Удалить
                      </button>
                    )}
                  </div>
                </div>
                {ev.description && (
                  <p className="text-sm text-stone-600 mt-2">{ev.description}</p>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
