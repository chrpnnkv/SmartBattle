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
  const [exportingId, setExportingId] = useState<string | null>(null);

  useEffect(() => {
    api.analytics.getReports()
      .then(setReports)
      .finally(() => setIsLoading(false));
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
            {reports.map((report, i) => (
              <div key={report.id} className={styles.reportCard} style={{ animationDelay: `${i * 60}ms` }}>
                <div className={styles.reportMain}>
                  <div className={styles.reportIcon}>
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                      <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                  </div>
                  <div className={styles.reportInfo}>
                    <span className={styles.reportTitle}>{report.quizTitle}</span>
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
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
}
