import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout/AppLayout';
import Button from '../components/ui/Button/Button';
import { api } from '../api';
import type { GameReport } from '../types';
import styles from './ReportsPage.module.css';

export default function ReportsPage() {
  const navigate = useNavigate();
  const [reports, setReports] = useState<GameReport[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [exportingId, setExportingId] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const loadReports = () => {
    setIsLoading(true);
    setError(null);
    api.analytics.getReports()
      .then((data) => setReports(data ?? []))
      .catch((err: unknown) => {
        setError((err as Error)?.message ?? 'Не удалось загрузить отчёты');
      })
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    loadReports();
    const onFocus = () => loadReports();
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, []);

  const handleExport = async (report: GameReport) => {
    setExportingId(report.id);
    try {
      const blob = await api.analytics.exportReportCsv(report.id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `report_${report.quizTitle}_${new Date(report.playedAt).toLocaleDateString('ru')}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setExportingId(null);
    }
  };

  return (
    <AppLayout>
      <div className={styles.page}>
        <div className={styles.pageHeader}>
          <div>
            <h1 className={styles.pageTitle}>Отчёты</h1>
            <p className={styles.pageSubtitle}>История проведённых квизов и результаты студентов</p>
          </div>
        </div>

        {error && (
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
            <span>{error}</span>
            <button
              onClick={loadReports}
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
        {isLoading ? (
          <div className={styles.loadingRows}>
            {[1, 2, 3].map((i) => <div key={i} className={styles.skeletonRow} />)}
          </div>
        ) : reports.length === 0 ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                <polyline points="14 2 14 8 20 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <p className={styles.emptyText}>Отчётов пока нет</p>
            <p className={styles.emptyHint}>Проведите квиз — результаты появятся здесь</p>
            <Button onClick={() => navigate('/dashboard')}>Перейти к квизам</Button>
          </div>
        ) : (
          <div className={styles.reportsList}>
            {reports.map((report, i) => {
              const isOpen = expandedId === report.id;
              return (
                <div key={report.id} className={styles.reportCard} style={{ animationDelay: `${i * 60}ms` }}>
                  <div className={styles.reportMain}>
                    <div className={styles.reportIcon}>
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    </div>
                    <div className={styles.reportInfo}>
                      <span className={styles.reportTitle}>{report.quizTitle || 'Без названия'}</span>
                      <span className={styles.reportMeta}>
                        {new Date(report.playedAt).toLocaleDateString('ru-RU', {
                          day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit',
                        })}
                      </span>
                    </div>
                  </div>

                  <div className={styles.reportStats}>
                    <div className={styles.reportStat}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                        <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                        <path d="M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                      </svg>
                      <span>{report.participantCount} участников</span>
                    </div>
                    <div className={styles.reportStat}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                        <polyline points="8 12 11 15 16 9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                      <span>Средний балл: <strong>{report.avgScore.toFixed(1)}</strong></span>
                    </div>
                  </div>

                  <div style={{ display: 'flex', gap: 8 }}>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setExpandedId(isOpen ? null : report.id)}
                      title={isOpen ? 'Свернуть результаты' : 'Показать результаты'}
                    >
                      {isOpen ? 'Скрыть' : 'Подробнее'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleExport(report)}
                      isLoading={exportingId === report.id}
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        <polyline points="7 10 12 15 17 10" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        <line x1="12" y1="15" x2="12" y2="3" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                      </svg>
                      CSV
                    </Button>
                  </div>

                  {isOpen && (
                    <div
                      style={{
                        gridColumn: '1 / -1',
                        marginTop: 16,
                        paddingTop: 16,
                        borderTop: '1px solid #e5e7eb',
                      }}
                    >
                      {report.leaderboard && report.leaderboard.length > 0 ? (
                        <div>
                          <div
                            style={{
                              display: 'grid',
                              gridTemplateColumns: '40px 1fr 90px 110px 70px',
                              gap: 12,
                              alignItems: 'center',
                              padding: '8px 0',
                              fontSize: 12,
                              color: '#6b7280',
                              textTransform: 'uppercase',
                              letterSpacing: '0.04em',
                            }}
                          >
                            <span>Место</span>
                            <span>Участник</span>
                            <span style={{ textAlign: 'right' }}>Очки</span>
                            <span style={{ textAlign: 'right' }}>Верных</span>
                            <span style={{ textAlign: 'right' }}>Точность</span>
                          </div>
                          {[...report.leaderboard]
                            .sort((a, b) => a.rank - b.rank)
                            .map((p) => {
                              const total = p.totalQuestions || 0;
                              const correct = p.correctAnswers ?? 0;
                              const percent = total > 0 ? Math.round((correct / total) * 100) : 0;
                              const medal = p.rank === 1 ? '🥇' : p.rank === 2 ? '🥈' : p.rank === 3 ? '🥉' : '';
                              return (
                                <div
                                  key={p.id}
                                  style={{
                                    display: 'grid',
                                    gridTemplateColumns: '40px 1fr 90px 110px 70px',
                                    gap: 12,
                                    alignItems: 'center',
                                    padding: '8px 0',
                                    borderTop: '1px solid #f3f4f6',
                                  }}
                                >
                                  <span style={{ fontWeight: 600, fontSize: 14 }}>
                                    {medal || `#${p.rank}`}
                                  </span>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                                    <div
                                      style={{
                                        width: 28,
                                        height: 28,
                                        borderRadius: '50%',
                                        background: p.avatarColor || '#7c3aed',
                                        color: '#fff',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        fontSize: 12,
                                        fontWeight: 600,
                                      }}
                                    >
                                      {p.avatarInitials || p.nickname.slice(0, 2).toUpperCase()}
                                    </div>
                                    <span style={{ fontSize: 14 }}>{p.nickname}</span>
                                  </div>
                                  <span style={{ textAlign: 'right', fontWeight: 600 }}>
                                    {p.score.toLocaleString('ru-RU')}
                                  </span>
                                  <span style={{ textAlign: 'right', color: '#6b7280' }}>
                                    {correct}{total > 0 ? `/${total}` : ''}
                                  </span>
                                  <span
                                    style={{
                                      textAlign: 'right',
                                      fontWeight: 600,
                                      color: percent >= 80 ? '#16a34a' : percent >= 50 ? '#d97706' : '#dc2626',
                                    }}
                                  >
                                    {percent}%
                                  </span>
                                </div>
                              );
                            })}
                        </div>
                      ) : (
                        <p style={{ color: '#6b7280', fontSize: 14, padding: '8px 0' }}>
                          В этой игре участников не было.
                        </p>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </AppLayout>
  );
}
