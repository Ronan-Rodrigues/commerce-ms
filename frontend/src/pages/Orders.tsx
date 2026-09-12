import React, { useEffect, useState } from 'react';
import { request } from '../lib/api';
import { useSSE } from '../hooks/useSSE';
import { Clock, CheckCircle2, Truck, RefreshCw, ShoppingBag, Radio } from 'lucide-react';

interface OrderItem {
  name: string;
  price: number;
  quantity: number;
  subtotal: number;
}

interface Order {
  id: string;
  total_amount: number;
  status: string;
  created_at: string;
  items: OrderItem[];
}

export const OrdersPage: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const { status: sseStatus, lastMessage } = useSSE();

  useEffect(() => {
    loadOrders();
  }, []);

  // Efeito reativo: Atualiza lista automaticamente quando chega evento SSE do backend!
  useEffect(() => {
    if (!lastMessage) return;

    if (lastMessage.type === 'order.criado') {
      loadOrders();
    } else if (lastMessage.type === 'payment.confirmado') {
      const { order_id } = lastMessage.data;
      setOrders((prev) =>
        prev.map((o) => (o.id === order_id ? { ...o, status: 'pago' } : o))
      );
    }
  }, [lastMessage]);

  const loadOrders = async () => {
    try {
      const data = await request<Order[]>('/orders');
      setOrders(data || []);
    } catch (err) {
      console.error('Falha ao carregar pedidos', err);
    } finally {
      setLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pago':
        return (
          <span className="inline-flex items-center space-x-1.5 px-3 py-1 bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 rounded-full text-xs font-semibold">
            <CheckCircle2 className="w-3.5 h-3.5" />
            <span>Pago</span>
          </span>
        );
      case 'enviado':
        return (
          <span className="inline-flex items-center space-x-1.5 px-3 py-1 bg-blue-500/10 text-blue-400 border border-blue-500/30 rounded-full text-xs font-semibold">
            <Truck className="w-3.5 h-3.5" />
            <span>Enviado</span>
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center space-x-1.5 px-3 py-1 bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded-full text-xs font-semibold">
            <Clock className="w-3.5 h-3.5" />
            <span>Aguardando Pagamento</span>
          </span>
        );
    }
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Meus Pedidos</h1>
          <p className="text-slate-400 text-sm mt-1">
            Status atualizado em tempo real via Server-Sent Events (SSE)
          </p>
        </div>

        {/* Indicador de Conexão SSE */}
        <div className="flex items-center space-x-3">
          <div className="flex items-center space-x-2 px-3 py-1.5 bg-slate-800 border border-slate-700 rounded-lg text-xs">
            <Radio
              className={`w-4 h-4 ${
                sseStatus === 'connected' ? 'text-emerald-400 animate-pulse' : 'text-slate-500'
              }`}
            />
            <span className="text-slate-300">
              SSE: {sseStatus === 'connected' ? 'Conectado (Tempo Real)' : 'Conectando...'}
            </span>
          </div>

          <button
            onClick={loadOrders}
            className="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg transition-colors border border-slate-700"
            title="Recarregar manualmente"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-20 text-slate-400">Carregando pedidos...</div>
      ) : orders.length === 0 ? (
        <div className="text-center py-20 bg-slate-800/40 rounded-2xl border border-slate-800">
          <ShoppingBag className="w-12 h-12 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-300 font-medium">Você ainda não realizou nenhum pedido</p>
          <p className="text-slate-500 text-sm">Visite o catálogo para fazer sua primeira compra.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <div
              key={order.id}
              className="bg-slate-800/80 border border-slate-700/80 rounded-2xl p-6 shadow-xl"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-slate-700/60 gap-3">
                <div>
                  <div className="flex items-center space-x-3">
                    <span className="text-base font-bold text-white">
                      Pedido #{order.id.substring(0, 8)}
                    </span>
                    {getStatusBadge(order.status)}
                  </div>
                  <p className="text-xs text-slate-400 mt-1">
                    Realizado em: {new Date(order.created_at).toLocaleString('pt-BR')}
                  </p>
                </div>

                <div className="text-right">
                  <span className="text-xs text-slate-400 block">Total do Pedido</span>
                  <span className="text-xl font-black text-blue-400">
                    R$ {(order.total_amount / 100).toFixed(2).replace('.', ',')}
                  </span>
                </div>
              </div>

              {/* Lista de Itens do Pedido */}
              <div className="mt-4 divide-y divide-slate-800/60">
                {order.items?.map((item, idx) => (
                  <div key={idx} className="py-2 flex items-center justify-between text-sm">
                    <span className="text-slate-200">
                      {item.quantity}x {item.name}
                    </span>
                    <span className="text-slate-400 font-mono">
                      R$ {(item.subtotal / 100).toFixed(2).replace('.', ',')}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

