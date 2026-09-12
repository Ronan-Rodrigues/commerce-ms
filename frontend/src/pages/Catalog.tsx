import React, { useEffect, useState } from 'react';
import { request } from '../lib/api';
import { ShoppingCart, Package, Plus, Trash2, CheckCircle2, ArrowRight } from 'lucide-react';

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  sku: string;
}

interface CartItem {
  product_id: string;
  name: string;
  price: number;
  quantity: number;
  subtotal: number;
}

interface Cart {
  items: CartItem[];
  total_amount: number;
}

export const CatalogPage: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [cart, setCart] = useState<Cart>({ items: [], total_amount: 0 });
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [checkoutSuccess, setCheckoutSuccess] = useState<string | null>(null);

  // Carrega produtos do catalog-service
  useEffect(() => {
    loadProducts();
    loadCart();
  }, []);

  const loadProducts = async () => {
    try {
      const data = await request<{ products: Product[] }>('/products');
      setProducts(data.products || []);
    } catch (err) {
      console.error('Falha ao carregar catálogo', err);
    } finally {
      setLoading(false);
    }
  };

  const loadCart = async () => {
    try {
      const data = await request<Cart>('/cart');
      setCart(data || { items: [], total_amount: 0 });
    } catch (err) {
      console.error('Falha ao buscar carrinho', err);
    }
  };

  const addToCart = async (product: Product) => {
    try {
      const updated = await request<Cart>('/cart/items', {
        method: 'POST',
        body: JSON.stringify({
          product_id: product.id,
          name: product.name,
          price: product.price,
          quantity: 1,
        }),
      });
      setCart(updated);
      setIsCartOpen(true);
    } catch (err) {
      alert('Erro ao adicionar produto ao carrinho');
    }
  };

  const removeFromCart = async (productId: string) => {
    try {
      const updated = await request<Cart>(`/cart/items/${productId}`, {
        method: 'DELETE',
      });
      setCart(updated);
    } catch (err) {
      console.error(err);
    }
  };

  const handleCheckout = async () => {
    if (cart.items.length === 0) return;
    setCheckoutLoading(true);
    setCheckoutSuccess(null);

    // Gera chave de idempotência única para evitar cobrança duplicada
    const idempotencyKey = `frontend-checkout-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;

    try {
      const order = await request<any>('/orders/checkout', {
        method: 'POST',
        headers: {
          'Idempotency-Key': idempotencyKey,
        },
      });

      setCheckoutSuccess(`Pedido #${order.id.substring(0, 8)} gerado com sucesso!`);
      setCart({ items: [], total_amount: 0 });
      setIsCartOpen(false);
    } catch (err: any) {
      alert(err.message || 'Falha ao processar checkout');
    } finally {
      setCheckoutLoading(false);
    }
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Barra de Ações Superior */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Catálogo de Produtos</h1>
          <p className="text-slate-400 text-sm mt-1">Conectado ao Catalog Service com estoque atômico</p>
        </div>

        <button
          onClick={() => setIsCartOpen(true)}
          className="relative px-4 py-2.5 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-100 rounded-xl flex items-center space-x-2 transition-all shadow-lg"
        >
          <ShoppingCart className="w-5 h-5 text-blue-400" />
          <span className="font-medium">Carrinho</span>
          {cart.items.length > 0 && (
            <span className="bg-blue-600 text-white text-xs px-2 py-0.5 rounded-full font-bold">
              {cart.items.reduce((acc, item) => acc + item.quantity, 0)}
            </span>
          )}
        </button>
      </div>

      {checkoutSuccess && (
        <div className="mb-6 p-4 bg-emerald-500/10 border border-emerald-500/30 rounded-xl flex items-center space-x-3 text-emerald-400">
          <CheckCircle2 className="w-6 h-6 flex-shrink-0" />
          <span className="font-medium">{checkoutSuccess} Veja o status em tempo real na aba "Meus Pedidos".</span>
        </div>
      )}

      {/* Grid de Produtos */}
      {loading ? (
        <div className="text-center py-20 text-slate-400">Carregando catálogo...</div>
      ) : products.length === 0 ? (
        <div className="text-center py-20 bg-slate-800/40 rounded-2xl border border-slate-800">
          <Package className="w-12 h-12 text-slate-600 mx-auto mb-3" />
          <p className="text-slate-300 font-medium">Nenhum produto cadastrado no momento</p>
          <p className="text-slate-500 text-sm">Os produtos criados via API aparecerão aqui automaticamente.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {products.map((p) => (
            <div
              key={p.id}
              className="bg-slate-800/80 border border-slate-700/80 rounded-2xl p-6 shadow-xl flex flex-col justify-between hover:border-slate-600 transition-all"
            >
              <div>
                <div className="flex items-start justify-between gap-2 mb-2">
                  <h3 className="font-bold text-lg text-white">{p.name}</h3>
                  <span className="text-xs px-2.5 py-1 bg-slate-700/60 text-slate-300 rounded-lg font-mono">
                    {p.sku}
                  </span>
                </div>
                <p className="text-sm text-slate-400 line-clamp-2 mb-4">{p.description}</p>
              </div>

              <div>
                <div className="flex items-baseline justify-between mb-4">
                  <span className="text-2xl font-black text-blue-400">
                    R$ {(p.price / 100).toFixed(2).replace('.', ',')}
                  </span>
                  <span className="text-xs text-slate-400">
                    Estoque: <strong className="text-slate-200">{p.stock} un.</strong>
                  </span>
                </div>

                <button
                  onClick={() => addToCart(p)}
                  disabled={p.stock <= 0}
                  className="w-full py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-xl flex items-center justify-center space-x-2 transition-all shadow-md shadow-blue-600/20 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  <Plus className="w-4 h-4" />
                  <span>{p.stock > 0 ? 'Adicionar ao Carrinho' : 'Esgotado'}</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Drawer Lateral do Carrinho */}
      {isCartOpen && (
        <div className="fixed inset-0 z-50 overflow-hidden">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setIsCartOpen(false)} />

          <div className="absolute inset-y-0 right-0 max-w-full flex">
            <div className="w-screen max-w-md bg-slate-900 border-l border-slate-800 p-6 flex flex-col justify-between shadow-2xl">
              <div>
                <div className="flex items-center justify-between pb-4 border-b border-slate-800">
                  <div className="flex items-center space-x-2">
                    <ShoppingCart className="w-5 h-5 text-blue-400" />
                    <h2 className="text-lg font-bold text-white">Meu Carrinho</h2>
                  </div>
                  <button onClick={() => setIsCartOpen(false)} className="text-slate-400 hover:text-white">
                    ✕
                  </button>
                </div>

                {/* Lista de Itens do Carrinho */}
                <div className="divide-y divide-slate-800 mt-4 max-h-[50vh] overflow-y-auto">
                  {cart.items.length === 0 ? (
                    <p className="text-center text-slate-500 py-10">Seu carrinho está vazio.</p>
                  ) : (
                    cart.items.map((item) => (
                      <div key={item.product_id} className="py-3 flex items-center justify-between">
                        <div>
                          <p className="text-sm font-semibold text-white">{item.name}</p>
                          <p className="text-xs text-slate-400">
                            {item.quantity}x R$ {(item.price / 100).toFixed(2).replace('.', ',')}
                          </p>
                        </div>
                        <div className="flex items-center space-x-3">
                          <span className="font-bold text-sm text-slate-200">
                            R$ {(item.subtotal / 100).toFixed(2).replace('.', ',')}
                          </span>
                          <button
                            onClick={() => removeFromCart(item.product_id)}
                            className="p-1 text-slate-400 hover:text-red-400 transition-colors"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>

              {/* Checkout e Total */}
              <div className="pt-4 border-t border-slate-800">
                <div className="flex items-center justify-between mb-4">
                  <span className="text-slate-400 text-sm">Total do Pedido</span>
                  <span className="text-2xl font-black text-white">
                    R$ {(cart.total_amount / 100).toFixed(2).replace('.', ',')}
                  </span>
                </div>

                <button
                  onClick={handleCheckout}
                  disabled={cart.items.length === 0 || checkoutLoading}
                  className="w-full py-3.5 bg-emerald-600 hover:bg-emerald-500 text-white font-bold rounded-xl flex items-center justify-center space-x-2 shadow-lg shadow-emerald-600/20 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  <span>{checkoutLoading ? 'Processando...' : 'Finalizar Compra'}</span>
                  <ArrowRight className="w-5 h-5" />
                </button>
                <p className="text-center text-xs text-slate-500 mt-2">
                  Garantido com Idempotency-Key contra cobrança dupla
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

