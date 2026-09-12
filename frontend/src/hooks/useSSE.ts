import { useEffect, useState } from 'react';

export interface SSEMessage {
  type: string;
  data: any;
  timestamp: string;
}

export function useSSE() {
  const [status, setStatus] = useState<'connected' | 'connecting' | 'disconnected'>('disconnected');
  const [lastMessage, setLastMessage] = useState<SSEMessage | null>(null);

  useEffect(() => {
    setStatus('connecting');
    const eventSource = new EventSource('/api/events');

    eventSource.onopen = () => {
      setStatus('connected');
    };

    // Escuta evento padronizado de pedido criado
    eventSource.addEventListener('order.criado', (e) => {
      try {
        const parsed = JSON.parse(e.data);
        setLastMessage({ type: 'order.criado', data: parsed, timestamp: new Date().toISOString() });
      } catch (err) {
        console.error('Erro ao decodificar evento SSE order.criado', err);
      }
    });

    // Escuta evento padronizado de pagamento confirmado
    eventSource.addEventListener('payment.confirmado', (e) => {
      try {
        const parsed = JSON.parse(e.data);
        setLastMessage({ type: 'payment.confirmado', data: parsed, timestamp: new Date().toISOString() });
      } catch (err) {
        console.error('Erro ao decodificar evento SSE payment.confirmado', err);
      }
    });

    // Escuta evento de carrinho abandonado
    eventSource.addEventListener('cart.abandonado', (e) => {
      try {
        const parsed = JSON.parse(e.data);
        setLastMessage({ type: 'cart.abandonado', data: parsed, timestamp: new Date().toISOString() });
      } catch (err) {
        console.error('Erro ao decodificar evento SSE cart.abandonado', err);
      }
    });

    eventSource.onerror = () => {
      setStatus('disconnected');
    };

    return () => {
      eventSource.close();
    };
  }, []);

  return { status, lastMessage };
}

