import React, { useState, useEffect } from 'react';
import { YMaps, Map, Placemark, ZoomControl } from '@pbe/react-yandex-maps';
import { Plus, Navigation, LogOut } from 'lucide-react';
import api from './api/client';
import { useAuth } from './context/AuthContext';
import AuthModal from './components/AuthModal';
import CreateEventModal from './components/CreateEventModal';

function App() {
  const { user, logout } = useAuth();
  const [events, setEvents] = useState<any[]>([]);
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isCreatingMode, setIsCreatingMode] = useState(false);
  const [selectedCoords, setSelectedCoords] = useState<[number, number] | null>(null);
  const [locationName, setLocationName] = useState('');

  useEffect(() => { fetchEvents(); }, []);

  const fetchEvents = async () => {
    try {
      const { data } = await api.get('/events');
      setEvents(data || []);
    } catch (err) { console.error('Load error'); }
  };

  const handleMapClick = (e: any) => {
    if (isCreatingMode) {
      const coords = e.get('coords');
      const target = e.get('target');
      const name = target?.properties?.get('name') || "";
      setSelectedCoords(coords);
      setLocationName(name);
      setIsCreateModalOpen(true);
      setIsCreatingMode(false);
    }
  };

  return (
    <div className="relative w-full h-screen bg-[#111] overflow-hidden">
      <YMaps query={{ apikey: 'ff8d2252-800f-4772-92e7-ee3857489035', lang: 'ru_RU' }}>
        <Map defaultState={{ center: [55.75, 37.61], zoom: 14 }} width="100%" height="100%" onClick={handleMapClick}
          instanceRef={(ref: any) => {
            if (ref) ref.container.getElement().style.filter = 'invert(100%) hue-rotate(180deg) brightness(1.3) contrast(1.1)';
          }}
          options={{ yandexMapDisablePoiInteractivity: false }}
        >
          {events.map(ev => (
            <Placemark key={ev.id} geometry={[ev.lat, ev.lon]} properties={{ balloonContent: ev.title }} options={{ iconColor: '#D8E983' }} />
          ))}
          <ZoomControl options={{ position: { right: 20, bottom: 100 } }} />
        </Map>
      </YMaps>

      <header className="absolute top-6 left-6 right-6 z-10 flex justify-between items-start pointer-events-none">
        <h1 className="text-5xl font-black text-[#D8E983] italic uppercase pointer-events-auto">HUDDLE</h1>
        <div className="pointer-events-auto">
          {user ? (
            <div className="bg-black/80 border border-[#D8E983]/30 p-2 pr-5 rounded-2xl flex items-center gap-4">
              <div className="w-12 h-12 bg-[#D8E983] rounded-xl flex items-center justify-center font-black text-black">{user.firstName[0]}</div>
              <div className="flex flex-col text-white font-bold uppercase text-[10px]">
                {user.firstName} {user.lastName}
                <button onClick={logout} className="text-red-400 text-left hover:text-red-300">Выйти</button>
              </div>
            </div>
          ) : (
            <button onClick={() => setIsAuthOpen(true)} className="bg-[#D8E983] text-black px-8 py-4 rounded-2xl font-black uppercase text-xs">Войти</button>
          )}
        </div>
      </header>

      <footer className="absolute bottom-12 left-0 right-0 z-10 flex justify-center pointer-events-none">
        <button onClick={() => user ? setIsCreatingMode(!isCreatingMode) : setIsAuthOpen(true)}
          className={`pointer-events-auto flex items-center gap-4 px-10 py-5 rounded-[2rem] font-black transition-all border-2 ${isCreatingMode ? 'bg-white text-black border-white animate-pulse scale-110' : 'bg-[#D8E983] text-black border-[#D8E983] hover:bg-[#FFFBB1]'}`}>
          {isCreatingMode ? <Navigation className="animate-bounce" /> : <Plus />}
          {isCreatingMode ? 'Выбери место' : 'Создать движ'}
        </button>
      </footer>

      <AuthModal isOpen={isAuthOpen} onClose={() => setIsAuthOpen(false)} />
      {selectedCoords && <CreateEventModal isOpen={isCreateModalOpen} onClose={() => { setIsCreateModalOpen(false); setSelectedCoords(null); }} coords={selectedCoords} locationName={locationName} onSuccess={fetchEvents} />}
    </div>
  );
}

export default App;