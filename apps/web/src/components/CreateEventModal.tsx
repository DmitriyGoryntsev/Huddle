import { useState, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventsApi, type Category } from '@/lib/api'

interface CreateEventModalProps {
  initialCoords: { lat: number; lon: number } | null
  categories: Category[]
  onClose: () => void
  onCreated: () => void
}

export default function CreateEventModal({ initialCoords, categories, onClose, onCreated }: CreateEventModalProps) {
  const queryClient = useQueryClient()
  const [form, setForm] = useState({
    category_id: categories[0]?.id ?? 0,
    title: '',
    description: '',
    lat: 55.7558,
    lon: 37.6173,
    start_time: '',
    max_participants: 4,
    price: 0,
    requires_approval: false,
  })
  const [error, setError] = useState('')

  useEffect(() => {
    if (initialCoords) {
      setForm((f) => ({ ...f, lat: initialCoords.lat, lon: initialCoords.lon }))
    }
  }, [initialCoords])

  const createMu = useMutation({
    mutationFn: eventsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['my-events'] })
      onCreated()
    },
    onError: (err) => setError(err instanceof Error ? err.message : 'Ошибка'),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.start_time) {
      setError('Укажите время начала')
      return
    }
    createMu.mutate({
      ...form,
      start_time: new Date(form.start_time).toISOString(),
    })
  }

  const subcategories = categories.filter((c) => c.parent_id != null)

  return (
    <div className="fixed inset-0 z-[2000] flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white rounded-2xl shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Создать событие</h2>
            <button onClick={onClose} className="text-stone-400 hover:text-stone-600 text-2xl">
              ×
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm">{error}</div>
            )}

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Категория</label>
              <select
                value={form.category_id}
                onChange={(e) => setForm((f) => ({ ...f, category_id: Number(e.target.value) }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
                required
              >
                {subcategories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
                {subcategories.length === 0 &&
                  categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Название</label>
              <input
                type="text"
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
                placeholder="Например: Падл в парке"
                required
                minLength={3}
                maxLength={100}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Описание</label>
              <textarea
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
                rows={2}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Время начала</label>
              <input
                type="datetime-local"
                value={form.start_time}
                onChange={(e) => setForm((f) => ({ ...f, start_time: e.target.value }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Макс. участников</label>
              <input
                type="number"
                min={2}
                value={form.max_participants}
                onChange={(e) => setForm((f) => ({ ...f, max_participants: Number(e.target.value) }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-stone-700 mb-1">Стоимость (₽)</label>
              <input
                type="number"
                min={0}
                step={0.01}
                value={form.price}
                onChange={(e) => setForm((f) => ({ ...f, price: Number(e.target.value) }))}
                className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500"
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="req"
                checked={form.requires_approval}
                onChange={(e) => setForm((f) => ({ ...f, requires_approval: e.target.checked }))}
                className="rounded"
              />
              <label htmlFor="req" className="text-sm text-stone-700">
                Требуется подтверждение участников
              </label>
            </div>

            <p className="text-xs text-stone-500">
              Координаты: {form.lat.toFixed(5)}, {form.lon.toFixed(5)} (выберите место на карте)
            </p>

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 py-2 rounded-lg border border-stone-300 text-stone-600 hover:bg-stone-50"
              >
                Отмена
              </button>
              <button
                type="submit"
                disabled={createMu.isPending}
                className="flex-1 py-2 rounded-lg bg-huddle-600 text-white font-medium hover:bg-huddle-700 disabled:opacity-50"
              >
                {createMu.isPending ? 'Создание...' : 'Создать'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
