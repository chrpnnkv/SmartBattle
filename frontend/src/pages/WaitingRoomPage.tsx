import { useEffect, useRef, useState } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { participantJoined, participantLeft, questionStarted } from '../store/slices/sessionSlice';
import { wsService } from '../api/wsService';
import type { WsParticipantJoinedPayload, WsQuestionStartedPayload, WsJoinedPayload } from '../types';
import Logo from '../components/ui/Logo/Logo';
import styles from './WaitingRoomPage.module.css';

function avatarColor(str: string) {
  const colors = ['#7c3aed','#2563eb','#16a34a','#dc2626','#ea580c','#0891b2','#be185d','#d97706'];
  let hash = 0;
  for (const ch of str) hash = ch.charCodeAt(0) + ((hash << 5) - hash);
  return colors[Math.abs(hash) % colors.length];
}

export default function WaitingRoomPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { session } = useAppSelector((s) => s.session);
  const stateData = location.state as { participantId?: string; quizTitle?: string } | null;

  const [participants, setParticipants] = useState<{ id: string; nickname: string }[]>([]);
  // Авторитетный счётчик от сервера. Сервер шлёт totalCount в joined/participant_joined/participant_left,
  // и он не страдает от пропуска одного броадкаста, в отличие от participants.length.
  const [serverCount, setServerCount] = useState<number | null>(null);
  const [wsError, setWsError] = useState<string | null>(null);
  const wsConnected = useRef(false);

  
  useEffect(() => {
    if (!sessionId) return;
    const poll = () => {
      try {
        const raw = localStorage.getItem('sb_mock_sessions');
        const sessions = raw ? (JSON.parse(raw) as { id: string; participants: { id: string; nickname: string }[] }[]) : [];
        const found = sessions.find((s) => s.id === sessionId);
        if (found?.participants) {
          setParticipants(found.participants.map((p) => ({ id: p.id, nickname: p.nickname })));
        }
      } catch {  }
    };
    poll();
    const t = setInterval(poll, 1500);
    return () => clearInterval(t);
  }, [sessionId]);

  useEffect(() => {
    if (!sessionId || wsConnected.current) return;
    wsConnected.current = true;

    wsService.connect(sessionId, {
      roomCode: sessionStorage.getItem('sb_pin') ?? '',
      name: sessionStorage.getItem('sb_nickname') ?? '',
      participantId: stateData?.participantId,
    });

    wsService.on<{ code: string; message: string }>('error', (payload) => {
      setWsError(payload.message);
    });

    // Сидируем список участников из снимка, который реалтайм отдаёт в joined.
    // Это даёт корректный счётчик при подключении/переподключении сразу,
    // а не только после новых participant_joined.
    wsService.on<WsJoinedPayload>('joined', (payload) => {
      if (!payload) return;
      if (payload.participants && payload.participants.length > 0) {
        const onlyStudents = payload.participants
          .filter((p) => p.nickname && p.nickname !== '')
          .map((p) => ({ id: p.id, nickname: p.nickname }));
        setParticipants(onlyStudents);
      }
      if (typeof payload.totalCount === 'number') {
        setServerCount(payload.totalCount);
      }
    });

    wsService.on<WsParticipantJoinedPayload>('participant_joined', (payload) => {
      dispatch(participantJoined(payload.participant));
      setParticipants((prev) => {
        const exists = prev.find((p) => p.id === payload.participant.id);
        if (exists) return prev;
        return [...prev, { id: payload.participant.id, nickname: payload.participant.nickname }];
      });
      if (typeof payload.totalCount === 'number') {
        setServerCount(payload.totalCount);
      }
    });

    wsService.on<{ participant_id: string; totalCount?: number }>('participant_left', (payload) => {
      dispatch(participantLeft(payload.participant_id));
      setParticipants((prev) => prev.filter((p) => p.id !== payload.participant_id));
      if (typeof payload.totalCount === 'number') {
        setServerCount(payload.totalCount);
      }
    });

    wsService.on<WsQuestionStartedPayload>('question_started', (payload) => {
      dispatch(questionStarted({
        question: payload.question,
        timeLimit: payload.question.timeLimitSeconds,
      }));
      
      let themeColor = 'purple';
      try {
        const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; quizId: string }[];
        const quizzes = JSON.parse(localStorage.getItem('sb_mock_quizzes') || '[]') as { id: string; settings: { themeColor: string } }[];
        const sess = sessions.find((s) => s.id === sessionId);
        const quiz = quizzes.find((q) => q.id === sess?.quizId);
        themeColor = quiz?.settings?.themeColor ?? 'purple';
      } catch {  }

      
      let sessionMode = 'teacher_paced';
      try {
        const allSessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; mode: string }[];
        const curSess = allSessions.find((s) => s.id === sessionId);
        sessionMode = curSess?.mode ?? 'teacher_paced';
      } catch {  }

      navigate(`/session/${sessionId}/question`, {
        state: {
          participantId: stateData?.participantId,
          questionIndex: payload.questionIndex,
          totalQuestions: payload.totalQuestions,
          question: payload.question,
          themeColor,
          mode: sessionMode,
        },
      });
    });

    
    
    const joinedAt = Date.now();
    const pollStatus = () => {
      
      if (Date.now() - joinedAt < 2000) return;
      try {
        const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; status: string }[];
        const sess = sessions.find((s) => s.id === sessionId);
        if (sess?.status === 'finished') {
          navigate(`/session/${sessionId}/finished`, { state: { participantId: stateData?.participantId } });
        }
      } catch {  }
    };
    const statusTimer = setInterval(pollStatus, 1500);

    return () => {
      clearInterval(statusTimer);
      wsService.disconnect();
      wsConnected.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- WS поднимаем один раз
    // на жизнь сессии, не реагируем на пересоздание location.state и dispatch.
  }, [sessionId]);

  const quizTitle = stateData?.quizTitle ?? 'Квиз';
  // Берём максимум: серверный totalCount авторитетен, но если в участниках уже больше
  // (например, дошёл лишний broadcast) — показываем то, что реально знаем.
  const count = Math.max(serverCount ?? 0, participants.length);
  const shown = participants.slice(0, 7);
  const extra = Math.max(0, count - 7);

  if (wsError) {
    return (
      <div className={styles.page}>
        <div className={styles.bg} aria-hidden="true" />
        <div className={styles.card}>
          <p style={{ color: '#ef4444', fontWeight: 600 }}>Ошибка подключения</p>
          <p>{wsError}</p>
          <button onClick={() => navigate('/')}>На главную</button>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.bg} aria-hidden="true" />

      <div className={styles.card}>
        <div className={styles.logoRow}>
          <div className={styles.logoIcon}>
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
        </div>
        <h1 className={styles.title}>{quizTitle}</h1>
        <div className={styles.statusRow}>
          <span className={styles.statusLabel}>ОЖИДАНИЕ НАЧАЛА ПРЕПОДАВАТЕЛЕМ...</span>
        </div>

        <p className={styles.readyText}>
          <span className={styles.readyDot} />
          Приготовьтесь! Битва скоро начнётся
        </p>
        <div className={styles.counterBadge}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
            <circle cx="9" cy="7" r="4" stroke="currentColor" strokeWidth="2"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
          </svg>
          {count} студентов присоединились
        </div>
        <div className={styles.participantsBox}>
          <span className={styles.participantsLabel}>СЕЙЧАС В КОМНАТЕ</span>
          <div className={styles.avatarRow}>
            {shown.map((p) => (
              <div
                key={p.id}
                className={styles.avatar}
                style={{ background: avatarColor(p.nickname) }}
                title={p.nickname}
              >
                {p.nickname.split(' ').filter(Boolean).map((w: string) => w[0].toUpperCase()).slice(0, 2).join('')}
              </div>
            ))}
            {extra > 0 && (
              <div className={[styles.avatar, styles.avatarExtra].join(' ')}>
                +{extra}
              </div>
            )}
          </div>
          <div className={styles.joinedBadge}>
            {count} присоединились
          </div>
        </div>
      </div>

      <p className={styles.footerBrand}>SMART <strong>BATTLE</strong></p>
    </div>
  );
}
