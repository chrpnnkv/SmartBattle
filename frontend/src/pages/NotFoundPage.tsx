import { useNavigate } from 'react-router-dom';
import Button from '../components/ui/Button/Button';
import styles from './NotFoundPage.module.css';

export default function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <div className={styles.page}>
      <h1 className={styles.code}>404</h1>
      <p className={styles.text}>Страница не найдена</p>
      <Button onClick={() => navigate('/')}>На главную</Button>
    </div>
  );
}
