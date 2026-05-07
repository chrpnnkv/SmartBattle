import { Link, useNavigate } from 'react-router-dom';
import { useAppSelector } from '../hooks/redux';
import Logo from '../components/ui/Logo/Logo';
import Button from '../components/ui/Button/Button';
import styles from './LandingPage.module.css';

export default function LandingPage() {
  const navigate = useNavigate();
  const { user } = useAppSelector((s) => s.auth);

  return (
    <div className={styles.page}>
      <header className={styles.navbar}>
        <Logo size="md" />
        <nav className={styles.navActions}>
          {user ? (
            <>
              <span className={styles.userGreet}>
                <div className={styles.userAvatar}>
                  {user.name.split(' ').filter(Boolean).map((w: string) => w[0].toUpperCase()).slice(0, 2).join('')}
                </div>
                {user.name}
              </span>
              <Button onClick={() => navigate('/dashboard')}>
                Перейти в кабинет
              </Button>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.loginLink}>Войти</Link>
              <Button onClick={() => navigate('/register')}>Зарегистрироваться</Button>
            </>
          )}
        </nav>
      </header>
      <main className={styles.hero}>
        <div className={styles.heroContent}>
          <div className={styles.heroBadge}>
            <span className={styles.heroBadgeDot} />
            Платформа для академических квизов
          </div>
          <h1 className={styles.heroTitle}>
            Учитесь лучше<br />
            с <span className={styles.heroAccent}>интересными<br />квизами</span>
          </h1>
          <p className={styles.heroSubtitle}>
            Осваивайте любой предмет через интерактивные квизы —
            обучение еще никогда не было таким интересным!
          </p>
          <div className={styles.heroActions}>
            {user ? (
              <>
                <Button size="lg" onClick={() => navigate('/dashboard')}>
                  Перейти в кабинет 
                </Button>
                <Button size="lg" variant="ghost" onClick={() => navigate('/join')}>
                  Войти по PIN
                </Button>
              </>
            ) : (
              <>
                <Button size="lg" onClick={() => navigate('/register')}>
                  Начать сейчас 
                </Button>
                <Button size="lg" variant="ghost" onClick={() => navigate('/join')}>
                  Войти по PIN
                </Button>
              </>
            )}
          </div>

        </div>
        <div className={styles.heroIllustration}>
          <div className={styles.illustrationCard}>
            <div className={styles.mockQuizHeader}>
              <span className={styles.mockQuizTitle}>История · Вопрос 5/12</span>
              <span className={styles.mockQuizTimer}>14</span>
            </div>
            <p className={styles.mockQuizQuestion}>
              В каком году человек впервые ступил на поверхность Луны?
            </p>
            <div className={styles.mockAnswers}>
              {[
                { color: '#ef4444', text: '1963', correct: false },
                { color: '#3b82f6', text: '1974', correct: false },
                { color: '#f59e0b', text: '1967', correct: false },
                { color: '#22c55e', text: '1969', correct: true },
              ].map((a) => (
                <div
                  key={a.text}
                  className={[styles.mockAnswer, a.correct ? styles.mockAnswerCorrect : ''].join(' ')}
                >
                  <span className={styles.mockAnswerDot} style={{ background: a.color }} />
                  <span>{a.text}</span>
                  {a.correct && (
                    <svg className={styles.mockAnswerCheck} width="14" height="14" viewBox="0 0 24 24" fill="none">
                      <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  )}
                </div>
              ))}
            </div>
          </div>

        </div>
      </main>
      <section className={styles.features}>
        {[
          {
            icon: (
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <path d="M12 20h9" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            ),
            title: 'Конструктор квизов',
            desc: '4 типа вопросов: выбор ответа, несколько ответов, правда/ложь, свободный ввод',
          },
          {
            icon: (
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            ),
            title: 'Аналитика в реальном времени',
            desc: 'Распределение ответов, среднее время, таблица лидеров — все обновляется мгновенно',
          },
          {
            icon: (
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                <rect x="2" y="3" width="20" height="14" rx="2" stroke="currentColor" strokeWidth="2"/>
                <path d="M8 21h8M12 17v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              </svg>
            ),
            title: 'Два режима проведения',
            desc: 'Управляемый преподавателем или свободный темп — выберите под свою аудиторию',
          },
        ].map((f) => (
          <div key={f.title} className={styles.featureCard}>
            <div className={styles.featureIcon}>{f.icon}</div>
            <h3 className={styles.featureTitle}>{f.title}</h3>
            <p className={styles.featureDesc}>{f.desc}</p>
          </div>
        ))}
      </section>
      <footer className={styles.footer}>
        <Logo size="sm" />
        <span className={styles.footerText}>© {new Date().getFullYear()} Smart Battle</span>
      </footer>
    </div>
  );
}
