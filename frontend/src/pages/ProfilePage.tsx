import { useState } from 'react';
import AppLayout from '../components/layout/AppLayout/AppLayout';
import Button from '../components/ui/Button/Button';
import Input from '../components/ui/Input/Input';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { changePassword } from '../store/slices/authSlice';
import styles from './ProfilePage.module.css';

type Tab = 'info' | 'password';

export default function ProfilePage() {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector((s) => s.auth);
  const [activeTab, setActiveTab] = useState<Tab>('info');

  
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPasswords, setShowPasswords] = useState(false);
  const [pwErrors, setPwErrors] = useState<Record<string, string>>({});
  const [pwLoading, setPwLoading] = useState(false);
  const [pwSuccess, setPwSuccess] = useState(false);
  const [pwServerError, setPwServerError] = useState('');

  const validatePassword = () => {
    const errs: Record<string, string> = {};
    if (!oldPassword) errs.old = 'Введите текущий пароль';
    if (!newPassword) errs.new = 'Введите новый пароль';
    else if (newPassword.length < 6) errs.new = 'Минимум 6 символов';
    if (!confirmPassword) errs.confirm = 'Подтвердите пароль';
    else if (confirmPassword !== newPassword) errs.confirm = 'Пароли не совпадают';
    if (oldPassword && newPassword && oldPassword === newPassword) {
      errs.new = 'Новый пароль должен отличаться от текущего';
    }
    setPwErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validatePassword()) return;
    setPwLoading(true);
    setPwServerError('');
    try {
      const result = await dispatch(changePassword({ oldPassword, newPassword }));
      if (changePassword.rejected.match(result)) {
        setPwServerError((result.payload as string) ?? 'Произошла ошибка');
        return;
      }
      setPwSuccess(true);
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: unknown) {
      setPwServerError((err as Error).message ?? 'Произошла ошибка');
    } finally {
      setPwLoading(false);
    }
  };

  const initials = user?.name
    .split(' ')
    .filter(Boolean)
    .map((w) => w[0].toUpperCase())
    .slice(0, 2)
    .join('') ?? '??';

  return (
    <AppLayout>
      <div className={styles.page}>
        <h1 className={styles.pageTitle}>Профиль</h1>
        <div className={styles.profileHeader}>
          <div className={styles.avatar}>{initials}</div>
          <div className={styles.profileMeta}>
            <span className={styles.profileName}>{user?.name}</span>
            <span className={styles.profileEmail}>{user?.email}</span>
            <span className={styles.profileRole}>Преподаватель</span>
          </div>
        </div>
        <div className={styles.tabs}>
          {(['info', 'password'] as Tab[]).map((tab) => (
            <button
              key={tab}
              className={[styles.tab, activeTab === tab ? styles.tabActive : ''].join(' ')}
              onClick={() => { setActiveTab(tab); setPwSuccess(false); setPwServerError(''); }}
            >
              {tab === 'info' ? 'Информация' : 'Смена пароля'}
            </button>
          ))}
        </div>

        <div className={styles.tabContent}>
          {activeTab === 'info' && (
            <div className={styles.infoGrid}>
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Имя</span>
                <span className={styles.infoValue}>{user?.name}</span>
              </div>
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Email</span>
                <span className={styles.infoValue}>{user?.email}</span>
              </div>
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Роль</span>
                <span className={styles.infoValue}>Преподаватель</span>
              </div>
              <div className={styles.infoRow}>
                <span className={styles.infoLabel}>Аккаунт создан</span>
                <span className={styles.infoValue}>
                  {user?.createdAt
                    ? new Date(user.createdAt).toLocaleDateString('ru-RU', {
                        day: 'numeric', month: 'long', year: 'numeric',
                      })
                    : '—'}
                </span>
              </div>
            </div>
          )}

          {activeTab === 'password' && (
            <form onSubmit={handleChangePassword} className={styles.pwForm} noValidate>
              {pwSuccess && (
                <div className={styles.successAlert}>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                    <polyline points="8 12 11 15 16 9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  Пароль успешно изменён
                </div>
              )}
              {pwServerError && (
                <div className={styles.errorAlert}>{pwServerError}</div>
              )}

              <div className={styles.pwFieldWrap}>
                <Input
                  label="Текущий пароль"
                  type={showPasswords ? 'text' : 'password'}
                  placeholder="Введите текущий пароль"
                  value={oldPassword}
                  onChange={(e) => { setOldPassword(e.target.value); setPwErrors((p) => ({ ...p, old: '' })); }}
                  error={pwErrors.old}
                />
              </div>

              <div className={styles.pwFieldWrap}>
                <Input
                  label="Новый пароль"
                  type={showPasswords ? 'text' : 'password'}
                  placeholder="Минимум 6 символов"
                  value={newPassword}
                  onChange={(e) => { setNewPassword(e.target.value); setPwErrors((p) => ({ ...p, new: '' })); }}
                  error={pwErrors.new}
                />
              </div>

              <div className={styles.pwFieldWrap}>
                <Input
                  label="Подтверждение пароля"
                  type={showPasswords ? 'text' : 'password'}
                  placeholder="Повторите новый пароль"
                  value={confirmPassword}
                  onChange={(e) => { setConfirmPassword(e.target.value); setPwErrors((p) => ({ ...p, confirm: '' })); }}
                  error={pwErrors.confirm}
                />
              </div>

              <label className={styles.showPwToggle}>
                <input
                  type="checkbox"
                  checked={showPasswords}
                  onChange={(e) => setShowPasswords(e.target.checked)}
                  className={styles.showPwCheckbox}
                />
                Показать пароли
              </label>

              <Button type="submit" size="lg" isLoading={pwLoading}>
                Сохранить изменения
              </Button>
            </form>
          )}
        </div>
      </div>
    </AppLayout>
  );
}
