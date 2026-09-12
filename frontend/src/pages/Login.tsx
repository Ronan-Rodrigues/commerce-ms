import React, { useState, useEffect } from 'react';
import { loginSchema, registerSchema } from '../lib/validators';
import { request, RateLimitError } from '../lib/api';
import { ShieldCheck, AlertCircle, Clock, ArrowRight } from 'lucide-react';

interface LoginProps {
  onSuccess: (user: any) => void;
}

export const LoginPage: React.FC<LoginProps> = ({ onSuccess }) => {
  const [isRegister, setIsRegister] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [rateLimitTimer, setRateLimitTimer] = useState<number | null>(null);

  // Efeito do timer de Rate Limit regressivo
  useEffect(() => {
    if (rateLimitTimer === null || rateLimitTimer <= 0) return;
    const interval = setInterval(() => {
      setRateLimitTimer((prev) => (prev && prev > 1 ? prev - 1 : null));
    }, 1000);
    return () => clearInterval(interval);
  }, [rateLimitTimer]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validação Zod no cliente
    if (isRegister) {
      const result = registerSchema.safeParse({ name, email, password });
      if (!result.success) {
        setError(result.error.errors[0].message);
        return;
      }
    } else {
      const result = loginSchema.safeParse({ email, password });
      if (!result.success) {
        setError(result.error.errors[0].message);
        return;
      }
    }

    setLoading(true);

    try {
      if (isRegister) {
        await request('/auth/register', {
          method: 'POST',
          body: JSON.stringify({ name, email, password }),
        });
        setIsRegister(false);
        setError('Conta criada com sucesso! Por favor, faça login.');
      } else {
        const data = await request<{ access_token: string; user: any }>('/auth/login', {
          method: 'POST',
          body: JSON.stringify({ email, password }),
        });
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('user', JSON.stringify(data.user));
        onSuccess(data.user);
      }
    } catch (err: any) {
      if (err instanceof RateLimitError) {
        setRateLimitTimer(err.retryAfter || 60);
        setError(`Muitas tentativas! Gateway bloqueou requisições temporariamente.`);
      } else {
        setError(err.message || 'Falha na comunicação com o servidor.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-[80vh] px-4">
      <div className="w-full max-w-md p-8 bg-slate-800/80 border border-slate-700 rounded-2xl shadow-2xl backdrop-blur-sm">
        <div className="flex items-center justify-center space-x-2 mb-6">
          <div className="p-2.5 bg-blue-600/20 text-blue-400 rounded-xl">
            <ShieldCheck className="w-7 h-7" />
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Commerce-MS</h1>
        </div>

        <h2 className="text-xl font-semibold text-slate-200 text-center mb-1">
          {isRegister ? 'Criar Nova Conta' : 'Acesse sua Conta'}
        </h2>
        <p className="text-sm text-slate-400 text-center mb-6">
          {isRegister ? 'Cadastre-se para iniciar suas compras' : 'Ambiente protegido com Argon2id e JWT RS256'}
        </p>

        {rateLimitTimer !== null && (
          <div className="mb-4 p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg flex items-center space-x-2 text-amber-400 text-sm">
            <Clock className="w-5 h-5 flex-shrink-0 animate-pulse" />
            <span>Rate Limit ativado pelo Gateway. Aguarde <strong>{rateLimitTimer}s</strong>.</span>
          </div>
        )}

        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg flex items-center space-x-2 text-red-400 text-sm">
            <AlertCircle className="w-5 h-5 flex-shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {isRegister && (
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">Nome Completo</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Ex: Ronan Rodrigues"
                disabled={loading || rateLimitTimer !== null}
                className="w-full px-4 py-2.5 bg-slate-900/60 border border-slate-700 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">E-mail</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="seu.email@exemplo.com"
              disabled={loading || rateLimitTimer !== null}
              className="w-full px-4 py-2.5 bg-slate-900/60 border border-slate-700 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1">Senha</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Mínimo 8 caracteres"
              disabled={loading || rateLimitTimer !== null}
              className="w-full px-4 py-2.5 bg-slate-900/60 border border-slate-700 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <button
            type="submit"
            disabled={loading || rateLimitTimer !== null}
            className="w-full mt-2 py-3 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-lg shadow-lg shadow-blue-600/30 transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span>{loading ? 'Processando...' : isRegister ? 'Cadastrar' : 'Entrar na Plataforma'}</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </form>

        <div className="mt-6 text-center">
          <button
            type="button"
            onClick={() => {
              setIsRegister(!isRegister);
              setError(null);
            }}
            className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
          >
            {isRegister ? 'Já possui uma conta? Faça login' : 'Não tem conta? Cadastre-se gratuitamente'}
          </button>
        </div>
      </div>
    </div>
  );
};

