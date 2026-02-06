import React, { useState } from 'react';
import { X, MapPin } from 'lucide-react';
import api from '../api/client';

export default function CreateEventModal({ isOpen, onClose, coords, locationName, onSuccess }: any) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [maxParticipants, setMaxParticipants] = useState(10);
  const [startTime, setStartTime] = useState('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      title,
      description,
      lat: coords[0],
      lon: coords[1],
      start_time: new Date(startTime).toISOString(),
      max_participants: Number(maxParticipants),
      category_id: 1,
      price: 0
    };

    try {
      await api.post('/events', payload);
      onSuccess();
      onClose();
    } catch (err: any) {
      alert('Ошибка: ' + (err.response?.data?.error || 'проверьте поля'));
    }
  };

  return (
    <div className="fixed inset-0 z-[3000] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative w-full max-w-lg bg-[#0a0a0a] border border-[#D8E983]/30 p-8 rounded-[2.5rem] shadow-2xl">
        <h2 className="text-3xl font-black text-[#D8E983] italic uppercase mb-6">Новый движ</h2>
        {locationName && (
          <div className="flex items-center gap-2 mb-6 p-4 bg-[#D8E983]/5 rounded-2xl border border-[#D8E983]/20">
            <MapPin size={18} className="text-[#D8E983]" />
            <span className="text-xs text-[#D8E983] font-bold uppercase">{locationName}</span>
          </div>
        )}
        <form onSubmit={handleSubmit} className="space-y-4">
          <input placeholder="НАЗВАНИЕ" required className="w-full bg-[#141414] p-5 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 font-bold text-xs" onChange={e => setTitle(e.target.value)} />
          <textarea placeholder="ОПИСАНИЕ" required rows={3} className="w-full bg-[#141414] p-5 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 font-bold text-xs resize-none" onChange={e => setDescription(e.target.value)} />
          <div className="grid grid-cols-2 gap-4">
            <input type="datetime-local" required className="bg-[#141414] p-5 rounded-2xl text-white border border-white/5 text-xs font-bold" onChange={e => setStartTime(e.target.value)} />
            <input type="number" placeholder="МЕСТ" required className="bg-[#141414] p-5 rounded-2xl text-white border border-white/5 text-xs font-bold" value={maxParticipants} onChange={e => setMaxParticipants(Number(e.target.value))} />
          </div>
          <button type="submit" className="w-full bg-[#D8E983] py-6 rounded-2xl font-black uppercase text-black hover:bg-[#FFFBB1] transition-all shadow-xl mt-6">Опубликовать</button>
        </form>
      </div>
    </div>
  );
}