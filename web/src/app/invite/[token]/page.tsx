'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { api } from '../../../lib/api';

export default function AcceptInvitationPage() {
  const params = useParams();
  const router = useRouter();
  const token = params.token as string;

  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'needLogin'>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) return;

    if (!api.isAuthenticated()) {
      setStatus('needLogin');
      setMessage('Для принятия приглашения необходимо войти в аккаунт.');
      return;
    }

    api
      .acceptInvitation(token)
      .then(() => {
        setStatus('success');
        setMessage('Приглашение принято! Вы добавлены в домохозяйство.');
      })
      .catch((err: Error) => {
        setStatus('error');
        setMessage(err.message || 'Не удалось принять приглашение.');
      });
  }, [token]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg)] px-4">
      <div className="w-full max-w-md bg-[var(--color-surface)] rounded-2xl p-8 shadow-lg text-center">
        <h1 className="text-2xl font-semibold text-[var(--color-text)] mb-6">
          Приглашение в hoWallet
        </h1>

        {status === 'loading' && (
          <p className="text-[var(--color-muted)]">Принимаем приглашение...</p>
        )}

        {status === 'success' && (
          <>
            <div className="text-5xl mb-4">✅</div>
            <p className="text-[var(--color-text)] mb-6">{message}</p>
            <button
              onClick={() => router.push('/dashboard')}
              className="w-full py-3 bg-[var(--color-primary)] text-white rounded-xl font-medium hover:opacity-90 transition cursor-pointer"
            >
              Перейти к дашборду
            </button>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="text-5xl mb-4">❌</div>
            <p className="text-red-400 mb-6">{message}</p>
            <button
              onClick={() => router.push('/dashboard')}
              className="w-full py-3 bg-[var(--color-border)] text-[var(--color-text)] rounded-xl font-medium hover:opacity-90 transition cursor-pointer"
            >
              На главную
            </button>
          </>
        )}

        {status === 'needLogin' && (
          <>
            <div className="text-5xl mb-4">🔑</div>
            <p className="text-[var(--color-muted)] mb-6">{message}</p>
            <button
              onClick={() => {
                // Save invitation token to redirect back after login
                localStorage.setItem('pending_invite_token', token);
                router.push('/login');
              }}
              className="w-full py-3 bg-[var(--color-primary)] text-white rounded-xl font-medium hover:opacity-90 transition cursor-pointer"
            >
              Войти
            </button>
            <button
              onClick={() => {
                localStorage.setItem('pending_invite_token', token);
                router.push('/register');
              }}
              className="w-full py-3 mt-3 bg-[var(--color-border)] text-[var(--color-text)] rounded-xl font-medium hover:opacity-90 transition cursor-pointer"
            >
              Зарегистрироваться
            </button>
          </>
        )}
      </div>
    </div>
  );
}
