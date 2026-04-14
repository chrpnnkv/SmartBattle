import { useNavigate, useLocation } from 'react-router-dom';
import { useAppDispatch } from '../hooks/redux';
import { resetSession } from '../store/slices/sessionSlice';
import Logo from '../components/ui/Logo/Logo';
import Button from '../components/ui/Button/Button';
import styles from './FinishedPage.module.css';
import type { SessionParticipant } from '../types';

export default function FinishedPage() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const nickname = sessionStorage.getItem('sb_nickname') ?? '';

  const location = useLocation();
  const stateData = location.state as {
    participantId?: string;
    score?: number;
    correctAnswers?: number;
    totalAnswered?: number;
  } | null;

  const myScore = stateData?.score ?? 0;
  const correctAnswers = stateData?.correctAnswers ?? 0;
  const totalAnswered = stateData?.totalAnswered ?? 0;

  
  const leaderboard = (() => {
    try {
      const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as {
        id: string; participants: SessionParticipant[];
      }[];
      
      const sess = sessions.filter((s) => s.participants?.length > 0).pop();
      if (!sess?.participants) return [];
      return [...sess.participants].sort((a, b) => b.score - a.score);
    } catch { return []; }
  })();

  const myRank = leaderboard.findIndex((p) => p.nickname === nickname) + 1;
  const totalParticipants = leaderboard.length;

  const getInitials = (name: string) =>
    name.split(' ').filter(Boolean).map((w) => w[0].toUpperCase()).slice(0, 2).join('');

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <Logo size="md" />

        <div className={styles.resultBlock}>
          <div className={styles.trophyIcon}>
            <svg width="36" height="36" viewBox="0 0 24 24" fill="none">
              <path d="M12 15c-4 0-7-3-7-7V4h14v4c0 4-3 7-7 7z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M5 4H2v2a4 4 0 0 0 3 3.87M19 4h3v2a4 4 0 0 1-3 3.87" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              <path d="M12 15v4M8 21h8" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
            </svg>
          </div>
          <h1 className={styles.title}>Квиз завершён!</h1>
          {nickname && <p className={styles.playerName}>{nickname}</p>}

          <div className={styles.myResult}>
            <div className={styles.myRank}>
              <span className={styles.myRankNum}>
                {myRank > 0 ? `${myRank}/${totalParticipants}` : '—'}
              </span>
              <span className={styles.myRankLabel}>место</span>
            </div>
            <div className={styles.myScore}>
              <span className={styles.myScoreNum}>{myScore.toLocaleString('ru-RU')}</span>
              <span className={styles.myScoreLabel}>очков</span>
            </div>
            <div className={styles.myScore}>
              <span className={styles.myScoreNum}>
                {correctAnswers}/{totalAnswered > 0 ? totalAnswered : '?'}
              </span>
              <span className={styles.myScoreLabel}>верных</span>
            </div>
          </div>
        </div>

        {leaderboard.length > 0 && (
          <div className={styles.leaderboard}>
            <h2 className={styles.leaderTitle}>Таблица лидеров</h2>
            <div className={styles.leaderList}>
              {leaderboard.slice(0, 10).map((p, i) => {
                const isMe = p.nickname === nickname;
                return (
                  <div key={p.id} className={[styles.leaderRow, isMe ? styles.leaderRowMe : ''].join(' ')}>
                    <span className={[
                      styles.leaderRank,
                      i === 0 ? styles.rankGold : i === 1 ? styles.rankSilver : i === 2 ? styles.rankBronze : '',
                    ].join(' ')}>
                      {i + 1}
                    </span>
                    <div className={styles.leaderAvatar} style={{ background: p.avatarColor }}>
                      {getInitials(p.nickname)}
                    </div>
                    <span className={styles.leaderName}>
                      {p.nickname}
                      {isMe && <span className={styles.meTag}>вы</span>}
                    </span>
                    <span className={styles.leaderScore}>
                      {p.score.toLocaleString('ru-RU')}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        <Button size="lg" fullWidth onClick={() => { dispatch(resetSession()); navigate('/'); }}>
          На главную
        </Button>
      </div>
    </div>
  );
}
