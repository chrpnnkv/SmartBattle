import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout/AppLayout';
import Button from '../components/ui/Button/Button';
import Badge from '../components/ui/Badge/Badge';
import ActivityChart from '../components/ui/ActivityChart/ActivityChart';
import Modal from '../components/ui/Modal/Modal';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { fetchMyQuizzes, updateQuiz, deleteQuiz } from '../store/slices/quizSlice';
import { createSession } from '../store/slices/sessionSlice';
import { api } from '../api';
import type { Quiz, QuizMode, GameReport } from '../types';
import styles from './DashboardPage.module.css';

// Полная неделя с понедельника по воскресенье — выходные тоже видны.
const WEEK_DAYS = ['ПН', 'ВТ', 'СР', 'ЧТ', 'ПТ', 'СБ', 'ВС'];

function buildActivityData(reports: GameReport[]) {
  const now = new Date();

  // dayOfWeek: 0 = Mon, 6 = Sun (Date#getDay() считает с воскресенья).
  const dayOfWeek = now.getDay() === 0 ? 6 : now.getDay() - 1;
  const monday = new Date(now);
  monday.setDate(now.getDate() - dayOfWeek);
  monday.setHours(0, 0, 0, 0);

  // 7 ячеек на 7 дней.
  const counts = [0, 0, 0, 0, 0, 0, 0];
  reports.forEach((r) => {
    const d = new Date(r.playedAt);
    if (Number.isNaN(d.getTime())) return;
    if (d < monday) return;
    const idx = d.getDay() === 0 ? 6 : d.getDay() - 1;
    if (idx >= 0 && idx < 7) counts[idx] += 1;
  });

  const max = Math.max(...counts, 1);
  const formatter = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long' });

  return WEEK_DAYS.map((label, i) => {
    const dayDate = new Date(monday);
    dayDate.setDate(monday.getDate() + i);

    let state: 'past' | 'today' | 'future' = 'past';
    if (i === dayOfWeek) state = 'today';
    else if (i > dayOfWeek) state = 'future';

    return {
      label,
      value: counts[i],
      max,
      state,
      title: `${label} · ${formatter.format(dayDate)} · ${counts[i]} игр`,
    };
  });
}

interface LaunchModalState {
  quiz: Quiz | null;
  mode: QuizMode;
  isOpen: boolean;
}

export default function DashboardPage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { user } = useAppSelector((s) => s.auth);
  const { quizzes, isLoading, error: quizError } = useAppSelector((s) => s.quiz);
  const { session, isLoading: sessionLoading } = useAppSelector((s) => s.session);

  const [reports, setReports] = useState<GameReport[]>([]);
  const [reportsError, setReportsError] = useState<string | null>(null);
  const [launchModal, setLaunchModal] = useState<LaunchModalState>({
    quiz: null, mode: 'teacher_paced', isOpen: false,
  });

  // Загружаем квизы и отчёты. Делаем функцию отдельно, чтобы можно было дёрнуть
  // её при возврате фокуса на вкладку и после того, как пользователь сыграет квиз.
  const loadReports = () => {
    setReportsError(null);
    api.analytics
      .getReports()
      .then((data) => setReports(data ?? []))
      .catch((err: unknown) => {
        // 401 уже обработан в realApiService (редирект на /login).
        // Здесь показываем только реальные ошибки фронту.
        const msg = (err as Error)?.message ?? 'Не удалось загрузить отчёты';
        setReportsError(msg);
      });
  };

  useEffect(() => {
    dispatch(fetchMyQuizzes());
    loadReports();
  }, [dispatch]);

  // Перезагружаем отчёты при возврате фокуса на вкладку, чтобы
  // после сыгранного квиза счётчик «Игр проведено» обновился сам.
  useEffect(() => {
    const onFocus = () => {
      dispatch(fetchMyQuizzes());
      loadReports();
    };
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [dispatch]);

  const [isLaunching, setIsLaunching] = useState(false);

  const openLaunchModal = (quiz: Quiz) => {
    setIsLaunching(false);
    setLaunchModal({ quiz, mode: 'teacher_paced', isOpen: true });
  };

  const handleLaunch = async () => {
    if (!launchModal.quiz || isLaunching) return;
    setIsLaunching(true);
    
    try {
      const keys = Object.keys(localStorage).filter((k) =>
        k.startsWith('sb_ws_signal_') ||
        k.startsWith('sb_current_question_') ||
        k.startsWith('sb_answered_')
      );
      keys.forEach((k) => localStorage.removeItem(k));
    } catch {  }
    
    sessionStorage.removeItem('sb_score');
    sessionStorage.removeItem('sb_correct');
    sessionStorage.removeItem('sb_total');

    const result = await dispatch(createSession({ quizId: launchModal.quiz.id, mode: launchModal.mode }));
    if (createSession.fulfilled.match(result)) {
      setLaunchModal((p) => ({ ...p, isOpen: false }));
      navigate(`/session/${result.payload.id}/analytics`);
    }
    setIsLaunching(false);
  };

  const handleTogglePublish = (quiz: Quiz) => {
    const newStatus = quiz.status === 'published' ? 'draft' : 'published';
    dispatch(updateQuiz({ id: quiz.id, data: { status: newStatus } }));
  };

  const handleDelete = (id: string) => {
    if (window.confirm('Удалить квиз? Это действие нельзя отменить.')) {
      dispatch(deleteQuiz(id));
    }
  };

  const recentQuizzes = quizzes.slice(0, 5);
  const activityData = buildActivityData(reports);
  const avgScore = reports.length
    ? (reports.reduce((s, r) => s + r.avgScore, 0) / reports.length).toFixed(1)
    : '—';

  return (
    <AppLayout>
      <div className={styles.page}>
        <div className={styles.pageHeader}>
          <div>
            <h1 className={styles.pageTitle}>Панель преподавателя</h1>
            <p className={styles.pageSubtitle}>
              Добро пожаловать, {user?.name}. Управляйте квизами и следите за прогрессом.
            </p>
          </div>
          <Button size="lg" onClick={() => navigate('/quiz/new')}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
              <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
            </svg>
            Создать новый квиз
          </Button>
        </div>
        {(reportsError || quizError) && (
          <div
            style={{
              padding: '12px 16px',
              marginBottom: 16,
              borderRadius: 8,
              background: '#fef2f2',
              color: '#b91c1c',
              fontSize: 14,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              gap: 12,
            }}
            role="alert"
          >
            <span>
              {quizError && reportsError
                ? `Ошибка загрузки данных: ${quizError}; ${reportsError}`
                : quizError
                  ? `Не удалось загрузить квизы: ${quizError}`
                  : `Не удалось загрузить отчёты: ${reportsError}`}
            </span>
            <button
              onClick={() => { dispatch(fetchMyQuizzes()); loadReports(); }}
              style={{
                background: 'transparent',
                border: '1px solid #b91c1c',
                color: '#b91c1c',
                padding: '4px 12px',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 13,
              }}
            >
              Повторить
            </button>
          </div>
        )}
        <div className={styles.statsGrid}>
          <div className={styles.statCard}>
            <span className={styles.statLabel}>Всего квизов</span>
            <span className={styles.statValue}>{quizzes.length}</span>
            <span className={styles.statHint}>
              {quizzes.filter(q => q.status === 'published').length} опубликовано
            </span>
          </div>
          <div className={styles.statCard}>
            <span className={styles.statLabel}>Средний балл</span>
            <span className={styles.statValue}>{avgScore}</span>
            <span className={styles.statHint}>
              {reports.length} игр проведено
            </span>
          </div>
        </div>
        <div className={styles.mainGrid}>
          <div className={styles.quizzesCard}>
            <div className={styles.cardHeader}>
              <h2 className={styles.cardTitle}>Ваши квизы</h2>
              <button className={styles.seeAllBtn} onClick={() => navigate('/reports')}>
                Отчёты →
              </button>
            </div>

            {isLoading ? (
              <div className={styles.loadingRows}>
                {[1, 2, 3].map((i) => <div key={i} className={styles.skeletonRow} />)}
              </div>
            ) : recentQuizzes.length === 0 ? (
              <div className={styles.emptyState}>
                <p>У вас пока нет квизов</p>
                <Button size="sm" onClick={() => navigate('/quiz/new')}>
                  Создать первый квиз
                </Button>
              </div>
            ) : (
              <div className={styles.quizList}>
                {recentQuizzes.map((quiz, i) => (
                  <div key={quiz.id} className={styles.quizRow} style={{ animationDelay: `${i * 60}ms` }}>
                    <div className={styles.quizIcon}>
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
                        <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    </div>
                    <div className={styles.quizInfo}>
                      <span className={styles.quizTitle}>{quiz.title}</span>
                      <span className={styles.quizMeta}>
                        {(() => {
                          const qCount = quiz.questions?.length ?? quiz.questionCount ?? 0;
                          
                          const totalSec = quiz.questions?.length
                            ? quiz.questions.reduce((s, q) => s + (q.timeLimitSeconds ?? 30), 0) + (quiz.questions.length * 15)
                            : (quiz.estimatedMinutes ?? 0) * 60;
                          const mins = Math.ceil(totalSec / 60);
                          return `${qCount} вопр. · ${mins} мин`;
                        })()}{' '}
                        · {new Date(quiz.createdAt).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })}
                      </span>
                      <Badge status={quiz.status} />
                    </div>
                    <div className={styles.quizActions}>
                      <Button variant="ghost" size="sm" onClick={() => navigate(`/quiz/${quiz.id}/edit`)}>
                        Ред.
                      </Button>
                      {quiz.status === 'published' ? (
                        <>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleTogglePublish(quiz)}
                            title="Вернуть в черновик"
                          >
                            В черновик
                          </Button>
                          <Button size="sm" onClick={() => openLaunchModal(quiz)}>
                            Начать
                          </Button>
                        </>
                      ) : (
                        <>
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => handleTogglePublish(quiz)}
                          >
                            Опубликовать
                          </Button>
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
          <div className={styles.activityCard}>
            <h2 className={styles.cardTitle}>Ваша активность</h2>
            <p className={styles.activitySub}>Квизов проведено за неделю</p>
            <ActivityChart data={activityData} />
          </div>
        </div>
      </div>
      <Modal
        isOpen={launchModal.isOpen}
        onClose={() => setLaunchModal((p) => ({ ...p, isOpen: false }))}
        title={`Запустить: ${launchModal.quiz?.title}`}
        footer={
          <>
            <Button variant="ghost" onClick={() => setLaunchModal((p) => ({ ...p, isOpen: false }))}>
              Отмена
            </Button>
            <Button onClick={handleLaunch} isLoading={sessionLoading}>
              Запустить →
            </Button>
          </>
        }
      >
        <div className={styles.modalContent}>
          <p className={styles.modalDesc}>Выберите режим проведения квиза:</p>
          <div className={styles.modeCards}>
            {([
              ['teacher_paced', '🎓', 'Управляет преподаватель', 'Вы контролируете переход между вопросами'],
              ['student_paced', '🚀', 'Свободный темп', 'Студенты проходят квиз в своём темпе'],
            ] as [QuizMode, string, string, string][]).map(([val, icon, name, desc]) => (
              <button
                key={val}
                className={[styles.modeCard, launchModal.mode === val ? styles.modeCardActive : ''].join(' ')}
                onClick={() => setLaunchModal((p) => ({ ...p, mode: val }))}
              >
                <span className={styles.modeIcon}>{icon}</span>
                <span className={styles.modeName}>{name}</span>
                <span className={styles.modeDesc}>{desc}</span>
              </button>
            ))}
          </div>
        </div>
      </Modal>
    </AppLayout>
  );
}
