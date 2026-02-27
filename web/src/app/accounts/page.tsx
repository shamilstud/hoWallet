'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '../../lib/api';
import type { Account, AccountType } from '../../types';

export default function AccountsPage() {
  const router = useRouter();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  // Form state
  const [formName, setFormName] = useState('');
  const [formType, setFormType] = useState<AccountType>('card');
  const [formBalance, setFormBalance] = useState('0');
  const [formCurrency, setFormCurrency] = useState('KZT');
  const [formError, setFormError] = useState('');

  useEffect(() => {
    if (!api.isAuthenticated()) {
      router.replace('/login');
      return;
    }
    loadAccounts();
  }, [router]);

  const loadAccounts = async () => {
    try {
      const accs = await api.listAccounts();
      setAccounts(accs);
    } catch {}
    setLoading(false);
  };

  const resetForm = () => {
    setFormName('');
    setFormType('card');
    setFormBalance('0');
    setFormCurrency('UAH');
    setFormError('');
    setEditingId(null);
    setShowForm(false);
  };

  const openEdit = (acc: Account) => {
    setFormName(acc.name);
    setFormType(acc.type);
    setFormBalance(acc.balance);
    setFormCurrency(acc.currency);
    setEditingId(acc.id);
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    try {
      if (editingId) {
        await api.updateAccount(editingId, {
          name: formName,
          type: formType,
          currency: formCurrency,
        });
      } else {
        await api.createAccount({
          name: formName,
          type: formType,
          balance: formBalance,
          currency: formCurrency,
        });
      }
      resetForm();
      loadAccounts();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Ошибка');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Удалить счёт?')) return;
    try {
      await api.deleteAccount(id);
      loadAccounts();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Не удалось удалить');
    }
  };

  const formatAmount = (amount: string, currency = 'KZT') => {
    const num = parseFloat(amount);
    try {
      return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency,
        minimumFractionDigits: 0,
      }).format(num);
    } catch {
      return `${num.toFixed(2)} ${currency}`;
    }
  };

  const typeLabels: Record<AccountType, string> = {
    card: 'Карта',
    deposit: 'Депозит',
    cash: 'Наличные',
  };

  if (loading) {
    return <div className="text-center py-20 text-gray-400">Загрузка...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Счета</h1>
        <button
          onClick={() => {
            resetForm();
            setShowForm(true);
          }}
          className="bg-indigo-600 text-white rounded-lg px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
        >
          + Новый счёт
        </button>
      </div>

      {/* Form */}
      {showForm && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold mb-4">
            {editingId ? 'Редактировать счёт' : 'Новый счёт'}
          </h2>
          {formError && (
            <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3 mb-4">
              {formError}
            </div>
          )}
          <form onSubmit={handleSubmit} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Название
              </label>
              <input
                type="text"
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                required
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                placeholder="Mono Black"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Тип
              </label>
              <select
                value={formType}
                onChange={(e) => setFormType(e.target.value as AccountType)}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
              >
                <option value="card">Карта</option>
                <option value="deposit">Депозит</option>
                <option value="cash">Наличные</option>
              </select>
            </div>
            {!editingId && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Начальный баланс
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={formBalance}
                  onChange={(e) => setFormBalance(e.target.value)}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none"
                />
              </div>
            )}
            {/* Валюта скрыта — по дефолту KZT, позже добавим выбор */}
            <div className="sm:col-span-2 flex gap-2">
              <button
                type="submit"
                className="bg-indigo-600 text-white rounded-lg px-4 py-2 text-sm font-medium hover:bg-indigo-700 transition-colors"
              >
                {editingId ? 'Сохранить' : 'Создать'}
              </button>
              <button
                type="button"
                onClick={resetForm}
                className="border border-gray-300 rounded-lg px-4 py-2 text-sm font-medium hover:bg-gray-50 transition-colors"
              >
                Отмена
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Accounts list */}
      {accounts.length === 0 ? (
        <div className="bg-white rounded-xl border border-gray-200 p-12 text-center text-gray-400">
          Нет счетов. Создайте свой первый счёт!
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {accounts.map((acc) => (
            <div
              key={acc.id}
              className="bg-white rounded-xl border border-gray-200 p-5 space-y-3"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold">{acc.name}</h3>
                  <span className="text-xs text-gray-400 bg-gray-100 rounded px-2 py-0.5">
                    {typeLabels[acc.type]}
                  </span>
                </div>
                <p className="text-lg font-bold">
                  {formatAmount(acc.balance, acc.currency)}
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => openEdit(acc)}
                  className="text-xs text-indigo-600 hover:underline"
                >
                  Изменить
                </button>
                <button
                  onClick={() => handleDelete(acc.id)}
                  className="text-xs text-red-500 hover:underline"
                >
                  Удалить
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
