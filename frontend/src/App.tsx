import { useEffect } from 'react';
import { useAppDispatch, useAppSelector } from './hooks/redux';
import { initAuth, logout } from './store/slices/authSlice';
import { AppRouter } from './router/AppRouter';

export default function App() {
  const dispatch = useAppDispatch();
  const { isInitialized } = useAppSelector((s) => s.auth);

  useEffect(() => {

    dispatch(initAuth());
  }, [dispatch]);

  // Глобальный листенер для 401 от realApiService.
  // Когда токен протух/потерян, мы хотим, чтобы Redux тоже знал об этом
  // (иначе ProtectedRoute будет держать пользователя на странице).
  useEffect(() => {
    const onUnauthorized = () => {
      dispatch(logout());
    };
    window.addEventListener('sb:unauthorized', onUnauthorized);
    return () => window.removeEventListener('sb:unauthorized', onUnauthorized);
  }, [dispatch]);

  
  
  if (!isInitialized) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          background: 'var(--color-bg)',
        }}
      >
        <div
          style={{
            width: 40,
            height: 40,
            border: '3px solid var(--color-primary-light)',
            borderTopColor: 'var(--color-primary)',
            borderRadius: '50%',
            animation: 'spin 0.7s linear infinite',
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return <AppRouter />;
}
