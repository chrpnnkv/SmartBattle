import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import AuthLayout from '../components/layout/AuthLayout/AuthLayout';
import Input from '../components/ui/Input/Input';
import Button from '../components/ui/Button/Button';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { register, clearError } from '../store/slices/authSlice';
import styles from './AuthPage.module.css';

export default function RegisterPage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { isLoading, error, user } = useAppSelector((s) => s.auth);

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<{
    name?: string; email?: string; password?: string; confirm?: string;
  }>({});

  useEffect(() => {
    if (user) navigate('/dashboard', { replace: true });
  }, [user, navigate]);

  useEffect(() => {
    return () => { dispatch(clearError()); };
  }, [dispatch]);

  const validate = () => {
    const errs: typeof fieldErrors = {};
    if (!name.trim()) errs.name = 'Введите имя';
    if (!email.trim()) errs.email = 'Введите email';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) errs.email = 'Некорректный email';
    if (!password) errs.password = 'Введите пароль';
    else if (password.length < 6) errs.password = 'Минимум 6 символов';
    if (!confirm) errs.confirm = 'Подтвердите пароль';
    else if (confirm !== password) errs.confirm = 'Пароли не совпадают';
    setFieldErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;
    dispatch(register({ name, email, password }));
  };

  return (
    <AuthLayout title="Создать аккаунт" subtitle="Зарегистрируйтесь как преподаватель">
      <form onSubmit={handleSubmit} className={styles.form} noValidate>
        {error && (
          <div className={styles.alertError} role="alert">{error}</div>
        )}

        <Input
          label="Имя"
          type="text"
          placeholder="Сергей Михайлов"
          value={name}
          onChange={(e) => { setName(e.target.value); setFieldErrors((p) => ({ ...p, name: undefined })); }}
          error={fieldErrors.name}
          autoComplete="name"
          autoFocus
        />

        <Input
          label="Email"
          type="email"
          placeholder="teacher@example.com"
          value={email}
          onChange={(e) => { setEmail(e.target.value); setFieldErrors((p) => ({ ...p, email: undefined })); }}
          error={fieldErrors.email}
          autoComplete="email"
        />

        <div className={styles.passwordWrapper}>
          <Input
            label="Пароль"
            type={showPassword ? 'text' : 'password'}
            placeholder="Минимум 6 символов"
            value={password}
            onChange={(e) => { setPassword(e.target.value); setFieldErrors((p) => ({ ...p, password: undefined })); }}
            error={fieldErrors.password}
            autoComplete="new-password"
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
          autoComplete="new-password"
        />

        <Button type="submit" fullWidth size="lg" isLoading={isLoading}>
          Зарегистрироваться 
        </Button>

        <div className={styles.divider}><span>или</span></div>

        <p className={styles.switchText}>
          Уже есть аккаунт?{' '}
          <Link to="/login" className={styles.switchLink}>
            Войти
          </Link>
        </p>
      </form>
    </AuthLayout>
  );
}
