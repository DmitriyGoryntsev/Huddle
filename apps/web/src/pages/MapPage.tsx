import { useState, useEffect, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { MapContainer, TileLayer, Marker, Popup, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { eventsApi, categoriesApi, type Event, type Category } from '@/lib/api'
import CreateEventModal from '@/components/CreateEventModal'
import EventCard from '@/components/EventCard'

// Fix default marker icon in Vite
delete (L.Icon.Default.prototype as unknown as { _getIconUrl?: unknown })._getIconUrl
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png',
})

const DEFAULT_CENTER: [number, number] = [55.7558, 37.6173]
const DEFAULT_ZOOM = 12
const DEFAULT_RADIUS = 10000

function MapClickHandler({ onMapClick }: { onMapClick: (lat: number, lon: number) => void }) {
  useMapEvents({
    click: (e) => onMapClick(e.latlng.lat, e.latlng.lng),
  })
  return null
}

export default function MapPage() {
  const queryClient = useQueryClient()
  const [center, setCenter] = useState<[number, number]>(DEFAULT_CENTER)
  const [clickedCoords, setClickedCoords] = useState<{ lat: number; lon: number } | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [userLocation, setUserLocation] = useState<[number, number] | null>(null)

  useEffect(() => {
    navigator.geolocation.getCurrentPosition(
      (pos) => setUserLocation([pos.coords.latitude, pos.coords.longitude]),
      () => {},
      { enableHighAccuracy: true }
    )
  }, [])

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: categoriesApi.list,
  })

  const { data: events = [], isLoading } = useQuery({
    queryKey: ['events', center[0], center[1], DEFAULT_RADIUS],
    queryFn: () => eventsApi.list({ lat: center[0], lon: center[1], radius: DEFAULT_RADIUS }),
  })

  const onMapClick = useCallback((lat: number, lon: number) => {
    setClickedCoords({ lat, lon })
    setShowCreateModal(true)
    setSelectedEvent(null)
  }, [])

  const onMarkerClick = useCallback((e: Event) => {
    setSelectedEvent(e)
    setClickedCoords(null)
    setShowCreateModal(false)
  }, [])

  const onEventCreated = useCallback(() => {
    setShowCreateModal(false)
    setClickedCoords(null)
    queryClient.invalidateQueries({ queryKey: ['events'] })
  }, [queryClient])

  return (
    <div className="relative h-[calc(100vh-3.5rem)]">
      <MapContainer
        center={center}
        zoom={DEFAULT_ZOOM}
        className="h-full w-full z-0"
        scrollWheelZoom
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <MapClickHandler onMapClick={onMapClick} />

        {userLocation && (
          <Marker position={userLocation}>
            <Popup>Вы здесь</Popup>
          </Marker>
        )}

        {events.map((ev) => (
          <Marker
            key={ev.id}
            position={[ev.lat, ev.lon]}
            eventHandlers={{ click: () => onMarkerClick(ev) }}
          >
            <Popup>
              <div className="min-w-[180px]">
                <p className="font-medium text-stone-800">{ev.title}</p>
                <p className="text-sm text-stone-500">
                  {categories.find((c) => c.id === ev.category_id)?.name}
                </p>
                <p className="text-xs text-stone-400">
                  {new Date(ev.start_time).toLocaleString('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}
                </p>
                <p className="text-sm text-stone-600 mt-1">
                  {ev.current_participants}/{ev.max_participants} участников
                </p>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>

      <div className="absolute top-4 left-4 z-[1000] flex gap-2">
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-huddle-600 text-white rounded-lg shadow-lg font-medium hover:bg-huddle-700"
        >
          + Создать событие
        </button>
      </div>

      {selectedEvent && (
        <div className="absolute bottom-4 left-4 right-4 z-[1000] max-w-md mx-auto">
          <EventCard
            event={selectedEvent}
            categories={categories}
            onJoin={() => queryClient.invalidateQueries({ queryKey: ['events'] })}
            onLeave={() => queryClient.invalidateQueries({ queryKey: ['events'] })}
            onClose={() => setSelectedEvent(null)}
          />
        </div>
      )}

      {showCreateModal && (
        <CreateEventModal
          initialCoords={clickedCoords}
          categories={categories}
          onClose={() => {
            setShowCreateModal(false)
            setClickedCoords(null)
          }}
          onCreated={onEventCreated}
        />
      )}

      {isLoading && (
        <div className="absolute top-4 right-4 z-[1000] px-3 py-1.5 bg-white/90 rounded-lg text-sm text-stone-600">
          Загрузка...
        </div>
      )}
    </div>
  )
}
