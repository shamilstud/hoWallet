'use client';

import { useEffect, useState, useCallback } from 'react';
import { api } from '../../lib/api';
import type { HouseholdMember, Invitation, Household } from '../../types';

export default function HouseholdPage() {
  const [household, setHousehold] = useState<Household | null>(null);
  const [members, setMembers] = useState<HouseholdMember[]>([]);
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [inviteEmail, setInviteEmail] = useState('');
  const [creating, setCreating] = useState(false);
  const [newHHName, setNewHHName] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(true);

  const hhId = api.getHouseholdId();

  const loadData = useCallback(async () => {
    if (!hhId) {
      setLoading(false);
      return;
    }
    try {
      const [hhs, mems, invs] = await Promise.all([
        api.listHouseholds(),
        api.listMembers(hhId),
        api.listPendingInvitations(hhId).catch(() => [] as Invitation[]),
      ]);
      const current = hhs.find((h) => h.id === hhId) || null;
      setHousehold(current);
      setMembers(mems);
      setInvitations(invs);
    } catch {
      setError('Не удалось загрузить данные');
    } finally {
      setLoading(false);
    }
  }, [hhId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!hhId || !inviteEmail.trim()) return;
    setError('');
    setSuccess('');
    try {
      await api.invite(hhId, inviteEmail.trim());
      setSuccess(`Приглашение отправлено на ${inviteEmail}`);
      setInviteEmail('');
      loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка отправки приглашения');
    }
  };

  const handleRemoveMember = async (userId: string, name: string) => {
    if (!hhId) return;
    if (!confirm(`Удалить участника ${name}?`)) return;
    setError('');
    try {
      await api.removeMember(hhId, userId);
      loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка удаления');
    }
  };

  const handleCreateHousehold = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newHHName.trim()) return;
    setError('');
    try {
      const hh = await api.createHousehold({ name: newHHName.trim() });
      api.setHouseholdId(hh.id);
      setNewHHName('');
      setCreating(false);
      window.location.reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка создания');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-400 text-lg">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">
          {household ? household.name : 'Домохозяйство'}
        </h1>
        <button
          onClick={() => setCreating(!creating)}
          className="text-sm bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition cursor-pointer"
        >
          + Новое домохозяйство
        </button>
      </div>

      {creating && (
        <form onSubmit={handleCreateHousehold} className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold mb-4">Создать домохозяйство</h2>
          <div className="flex gap-3">
            <input
              type="text"
              value={newHHName}
              onChange={(e) => setNewHHName(e.target.value)}
              placeholder="Название"
              required
              className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 outline-none"
            />
            <button
              type="submit"
              className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700 transition cursor-pointer"
            >
              Создать
            </button>
            <button
              type="button"
              onClick={() => setCreating(false)}
              className="text-gray-500 px-3 py-2 text-sm hover:text-gray-700 cursor-pointer"
            >
              Отмена
            </button>
          </div>
        </form>
      )}

      {error && (
        <div className="bg-red-50 text-red-600 text-sm rounded-lg p-3">{error}</div>
      )}
      {success && (
        <div className="bg-green-50 text-green-700 text-sm rounded-lg p-3">{success}</div>
      )}

      {/* Members */}
      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Участники</h2>
        {members.length === 0 ? (
          <p className="text-gray-400 text-sm">Нет участников</p>
        ) : (
          <div className="divide-y divide-gray-100">
            {members.map((m) => (
              <div key={m.user_id} className="flex items-center justify-between py-3">
                <div>
                  <p className="font-medium text-gray-900">{m.user_name || m.email}</p>
                  <p className="text-sm text-gray-500">{m.email}</p>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className={`text-xs px-2 py-1 rounded-full ${
                      m.role === 'owner'
                        ? 'bg-indigo-50 text-indigo-700'
                        : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {m.role === 'owner' ? 'Владелец' : 'Участник'}
                  </span>
                  {m.role !== 'owner' && household?.owner_id === members.find((x) => x.role === 'owner')?.user_id && (
                    <button
                      onClick={() => handleRemoveMember(m.user_id, m.user_name || m.email)}
                      className="text-xs text-red-500 hover:text-red-700 cursor-pointer"
                    >
                      Удалить
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Invite */}
      <div className="bg-white rounded-xl border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Пригласить участника</h2>
        <form onSubmit={handleInvite} className="flex gap-3">
          <input
            type="email"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            placeholder="email@example.com"
            required
            className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 outline-none"
          />
          <button
            type="submit"
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-indigo-700 transition cursor-pointer"
          >
            Пригласить
          </button>
        </form>
      </div>

      {/* Pending invitations */}
      {invitations.length > 0 && (
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Ожидающие приглашения</h2>
          <div className="divide-y divide-gray-100">
            {invitations.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between py-3">
                <div>
                  <p className="font-medium text-gray-900">{inv.email}</p>
                  <p className="text-sm text-gray-400">
                    до {new Date(inv.expires_at).toLocaleDateString('ru-RU')}
                  </p>
                </div>
                <span className="text-xs px-2 py-1 rounded-full bg-yellow-50 text-yellow-700">
                  Ожидает
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
