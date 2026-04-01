import { useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import AuthLayout from '../components/layout/AuthLayout/AuthLayout';
import Input from '../components/ui/Input/Input';
import Button from '../components/ui/Button/Button';
import { api } from '../api';
import styles from './AuthPage.module.css';

export default function ResetPasswordPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token') ?? '';

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<{ password?: string; confirm?: string }>({});
  const [serverError, setServerError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  if (!token) {
    return (
      <AuthLayout title="Недействительная ссылка" subtitle="Ссылка для сброса пароля устарела или неверна">
        <div className={styles.form}>
          <Link to="/forgot-password" className={styles.switchLink} style={{ textAlign: 'center', display: 'block' }}>
            Запросить новую ссылку
          </Link>
        </div>
      </AuthLayout>
    );
  }

  const validate = () => {
    const errs: typeof fieldErrors = {};
    if (!password) errs.password = 'Введите пароль';
    else if (password.length < 6) errs.password = 'Минимум 6 символов';
    if (!confirm) errs.confirm = 'Подтвердите пароль';
    else if (confirm !== password) errs.confirm = 'Пароли не совпадают';
    setFieldErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    setIsLoading(true);
    try {
      await api.auth.resetPassword({ token, newPassword: password });
      navigate('/login', { state: { message: 'Пароль успешно изменён. Войдите с новым паролем.' } });
    } catch (err: unknown) {
      setServerError((err as Error).message ?? 'Произошла ошибка');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthLayout title="Новый пароль" subtitle="Придумайте надёжный пароль для вашего аккаунта">
      <form onSubmit={handleSubmit} className={styles.form} noValidate>
        {serverError && <div className={styles.alertError}>{serverError}</div>}

        <div className={styles.passwordWrapper}>
          <Input
            label="Новый пароль"
            type={showPassword ? 'text' : 'password'}
            placeholder="Минимум 6 символов"
            value={password}
            onChange={(e) => { setPassword(e.target.value); setFieldErrors((p) => ({ ...p, password: undefined })); }}
            error={fieldErrors.password}
            autoFocus
          />
          <button
            type="button"
            className={styles.showPasswordBtn}
            onClick={() => setShowPassword((v) => !v)}
            aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
          >
            {showPassword ? (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                <line x1="1" y1="1" x2="23" y2="23" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              </svg>
            ) : (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
              </svg>
            )}
          </button>
        </div>

        <Input
          label="Подтверждение пароля"
          type={showPassword ? 'text' : 'password'}
          placeholder="Повторите пароль"
          value={confirm}
          onChange={(e) => { setConfirm(e.target.value); setFieldErrors((p) => ({ ...p, confirm: undefined })); }}
          error={fieldErrors.confirm}
        />

        <Button type="submit" fullWidth size="lg" isLoading={isLoading}>
          Сохранить новый пароль
        </Button>
      </form>
    </AuthLayout>
  );
}
