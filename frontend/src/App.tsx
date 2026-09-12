import React, { useState, useEffect } from 'react';
import { LoginPage } from './pages/Login';
import { CatalogPage } from './pages/Catalog';
import { OrdersPage } from './pages/Orders';
import { ShieldCheck, LogOut, ShoppingBag, Clock } from 'lucide-react';

export const App: React.FC = () => {
  const [user, setUser] = useState<any>(null);
  const [currentTab, setCurrentTab] = useState<'catalog' | 'orders'>('catalog');

  useEffect(() => {
    const savedUser = localStorage.getItem('user');
    const token = localStorage.getItem('access_token');
    if (savedUser && token) {
      setUser(JSON.parse(savedUser));
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    setUser(null);
  };

  if (!user) {
    return <LoginPage onSuccess={setUser} />;
  }

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 flex flex-col">
      {/* Header Principal */}
      <header className="bg-slate-950/80 border-b border-slate-800 sticky top-0 z-40 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-blue-600/20 text-blue-400 rounded-xl">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <span className="font-bold text-lg text-white">Commerce-MS</span>
              <span className="hidden sm:inline-block ml-2 px-2 py-0.5 bg-blue-900/50 text-blue-300 text-xs rounded border border-blue-700/50 font-mono">
                Clean Architecture + SSE
              </span>
            </div>
          </div>

          {/* Abas Centrais */}
          <nav className="flex items-center space-x-1 bg-slate-900 p-1 rounded-xl border border-slate-800">
            <button
              onClick={() => setCurrentTab('catalog')}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all flex items-center space-x-2 ${
                currentTab === 'catalog'
                  ? 'bg-blue-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <ShoppingBag className="w-4 h-4" />
              <span>Catálogo</span>
            </button>

            <button
              onClick={() => setCurrentTab('orders')}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all flex items-center space-x-2 ${
                currentTab === 'orders'
                  ? 'bg-blue-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Clock className="w-4 h-4" />
              <span>Pedidos (SSE)</span>
            </button>
          </nav>

          {/* Usuário e Logout */}
          <div className="flex items-center space-x-4">
            <div className="hidden md:block text-right">
              <p className="text-xs font-semibold text-white">{user.name}</p>
              <p className="text-[10px] text-slate-400">{user.email}</p>
            </div>
            <button
              onClick={handleLogout}
              className="p-2 text-slate-400 hover:text-red-400 transition-colors rounded-lg hover:bg-slate-800"
              title="Sair da conta"
            >
              <LogOut className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Conteúdo da Aba Ativa */}
      <main className="flex-1">
        {currentTab === 'catalog' ? <CatalogPage /> : <OrdersPage />}
      </main>

      {/* Rodapé Informativo */}
      <footer className="border-t border-slate-800/80 py-6 text-center text-xs text-slate-500 bg-slate-950/40">
        <p>Commerce-MS — Plataforma de Microsserviços em Go (Clean Architecture, Redis, PostgreSQL, SSE, n8n)</p>
        <p className="mt-1">Desenvolvido por Ronan Rodrigues</p>
      </footer>
    </div>
  );
};

