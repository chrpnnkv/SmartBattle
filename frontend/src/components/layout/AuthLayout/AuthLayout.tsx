import { Link } from 'react-router-dom';
import Logo from '../../ui/Logo/Logo';
import styles from './AuthLayout.module.css';

interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
}

export default function AuthLayout({ children, title, subtitle }: AuthLayoutProps) {
  return (
    <div className={styles.page}>
      <div className={styles.blob1} aria-hidden="true" />
      <div className={styles.blob2} aria-hidden="true" />

      <div className={styles.card}>
        <div className={styles.cardHeader}>
          <Link to="/" className={styles.logoLink}>
            <Logo size="lg" />
          </Link>
          <h1 className={styles.title}>{title}</h1>
          <p className={styles.subtitle}>{subtitle}</p>
        </div>

        {children}
      </div>

      <p className={styles.footer}>
        Smart Battle © {new Date().getFullYear()}
      </p>
    </div>
  );
}
