import { useState } from 'react';
import { Link } from 'react-router-dom';
import AuthLayout from '../components/layout/AuthLayout/AuthLayout';
import Input from '../components/ui/Input/Input';
import Button from '../components/ui/Button/Button';
import { api } from '../api';
import styles from './AuthPage.module.css';

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [sent, setSent] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) { setError('Введите email'); return; }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { setError('Некорректный email'); return; }

    setIsLoading(true);
    try {
      await api.auth.forgotPassword({ email });
      setSent(true);
    } catch (err: unknown) {
      setError((err as Error).message ?? 'Произошла ошибка');
    } finally {
      setIsLoading(false);
    }
  };

  if (sent) {
    return (
      <AuthLayout title="Письмо отправлено" subtitle={`Инструкции по сбросу пароля отправлены на ${email}`}>
        <div className={styles.form}>
          <div className={styles.successBlock}>
            <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
              <polyline points="8 12 11 15 16 9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <p>Проверьте папку «Входящие» и перейдите по ссылке из письма.</p>
          </div>
          <Link to="/login" className={styles.switchLink} style={{ textAlign: 'center', display: 'block' }}>
             Вернуться ко входу
          </Link>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title="Сброс пароля"
      subtitle="Введите email — мы пришлем ссылку для восстановления"
    >
      <form onSubmit={handleSubmit} className={styles.form} noValidate>
        {error && <div className={styles.alertError}>{error}</div>}

        <Input
          label="Email"
          type="email"
          placeholder="teacher@example.com"
          value={email}
          onChange={(e) => { setEmail(e.target.value); setError(''); }}
          autoFocus
        />

        <Button type="submit" fullWidth size="lg" isLoading={isLoading}>
          Отправить письмо
        </Button>

        <p className={styles.switchText}>
          Вспомнили пароль?{' '}
          <Link to="/login" className={styles.switchLink}>Войти</Link>
        </p>
      </form>
    </AuthLayout>
  );
}
