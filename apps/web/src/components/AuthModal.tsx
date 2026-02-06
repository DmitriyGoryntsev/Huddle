import React, { useState } from 'react';
import { X } from 'lucide-react';
import api from '../api/client';
import { useAuth } from '../context/AuthContext';

export default function AuthModal({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) {
  const [isLogin, setIsLogin] = useState(true);
  const { login } = useAuth();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');

  if (!isOpen) return null;

  const toggleMode = () => {
    setIsLogin(!isLogin);
    setEmail(''); setPassword(''); setFirstName(''); setLastName('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const endpoint = isLogin ? '/auth/login' : '/auth/register';
      const payload = isLogin ? { email, password } : { email, password, firstName, lastName };
      const { data } = await api.post(endpoint, payload);
      
      if (data.tokens?.access_token) {
        const userData = data.user || { id: 'temp', email, firstName: firstName || email.split('@')[0], lastName };
        login(data.tokens.access_token, userData);
        onClose();
      } else if (!isLogin) {
        alert("Регистрация успешна! Войдите в аккаунт.");
        setIsLogin(true);
      }
    } catch (err: any) {
      alert(err.response?.data?.error || 'Ошибка входа. Проверьте данные.');
    }
  };

  return (
    <div className="fixed inset-0 z-[2000] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/90 backdrop-blur-md" onClick={onClose} />
      <div className="relative w-full max-w-md bg-[#0a0a0a] border border-[#D8E983]/30 p-10 rounded-[2.5rem] shadow-2xl">
        <h2 className="text-4xl font-black text-[#D8E983] mb-8 italic uppercase tracking-tighter">
          {isLogin ? 'HUDDLE IN' : 'JOIN US'}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!isLogin && (
            <div className="grid grid-cols-2 gap-3">
              <input placeholder="ИМЯ" required className="bg-[#141414] p-4 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 transition-all text-xs font-bold" value={firstName} onChange={e => setFirstName(e.target.value)} />
              <input placeholder="ФАМИЛИЯ" required className="bg-[#141414] p-4 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 transition-all text-xs font-bold" value={lastName} onChange={e => setLastName(e.target.value)} />
            </div>
          )}
          <input type="email" placeholder="EMAIL" required className="w-full bg-[#141414] p-4 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 text-xs font-bold" value={email} onChange={e => setEmail(e.target.value)} />
          <input type="password" placeholder="ПАРОЛЬ" required className="w-full bg-[#141414] p-4 rounded-2xl text-white border border-white/5 outline-none focus:border-[#D8E983]/50 text-xs font-bold" value={password} onChange={e => setPassword(e.target.value)} />
          <button type="submit" className="w-full bg-[#D8E983] py-5 rounded-2xl font-black uppercase text-black hover:bg-[#FFFBB1] hover:scale-[1.02] active:scale-95 transition-all shadow-xl mt-4">
            {isLogin ? 'ВОЙТИ' : 'ЗАРЕГИСТРИРОВАТЬСЯ'}
          </button>
        </form>
        <button onClick={toggleMode} className="w-full text-center text-[9px] text-white/30 mt-8 uppercase tracking-[0.4em] hover:text-[#D8E983] transition-colors">
          {isLogin ? 'НЕТ АККАУНТА? РЕГИСТРАЦИЯ' : 'ЕСТЬ АККАУНТ? ВОЙТИ'}
        </button>
      </div>
    </div>
  );
}