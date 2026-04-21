import { useEffect, useRef, useState, useCallback } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { wsService } from '../api/wsService';
import { api } from '../api';
import type { Question, WsQuestionStartedPayload, WsQuestionEndedPayload } from '../types';
import styles from './QuestionPage.module.css';

function TimerRing({ seconds, total }: { seconds: number; total: number }) {
  const r = 38;
  const circ = 2 * Math.PI * r;
  const pct = total > 0 ? seconds / total : 0;
  const dash = pct * circ;
  const color = seconds <= 5 ? '#ef4444' : seconds <= 10 ? '#f59e0b' : 'var(--color-primary)';
  const prevTotal = useRef(total);
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (total !== prevTotal.current) {
      prevTotal.current = total;
      isFirstRender.current = true;
      const t = requestAnimationFrame(() => {
        isFirstRender.current = false;
      });
      return () => cancelAnimationFrame(t);
    }
  }, [total]);

  useEffect(() => {
    isFirstRender.current = false;
  }, []);

  return (
    <div className={styles.timerRing}>
      <svg width="96" height="96" viewBox="0 0 96 96">
        <circle cx="48" cy="48" r={r} fill="none" stroke="#e4e4f0" strokeWidth="5"/>
        <circle
          cx="48" cy="48" r={r} fill="none"
          stroke={color} strokeWidth="5"
          strokeDasharray={`${dash} ${circ}`}
          strokeLinecap="round"
          transform="rotate(-90 48 48)"
          style={{
            transition: isFirstRender.current ? 'none' : 'stroke-dasharray 1s linear, stroke 0.3s',
          }}
        />
      </svg>
      <span className={styles.timerValue} style={{ color }}>{seconds}</span>
    </div>
  );
}

const COLOR_HEX: Record<string, string> = {
  red: '#ef4444', blue: '#3b82f6', yellow: '#f59e0b', green: '#22c55e',
};

function AnswerOption({ text, color, selected, correct, submitted, showCorrect, onClick }: {
  text: string; color: string; selected: boolean;
  correct?: boolean; submitted: boolean; showCorrect: boolean; onClick: () => void;
}) {
  const hex = COLOR_HEX[color] ?? 'var(--color-primary)';
  let state = '';
  if (submitted && showCorrect) {
    
    if (correct) state = styles.optionCorrect;
    else if (selected) state = styles.optionWrong;
    else state = styles.optionDimmed;
  } else if (submitted && !showCorrect) {
    
    if (selected) state = styles.optionSelected;
    else state = styles.optionDimmed;
  } else if (selected) {
    state = styles.optionSelected;
  }

  return (
    <button
      className={[styles.option, state].join(' ')}
      onClick={onClick}
      disabled={submitted}
      style={{ '--option-color': hex } as React.CSSProperties}
    >
      <span className={styles.optionDot} style={{ background: hex }} />
      <span className={styles.optionText}>{text}</span>
      {submitted && showCorrect && correct && (
        <svg className={styles.optionCheck} width="18" height="18" viewBox="0 0 24 24" fill="none">
          <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      )}
      {submitted && showCorrect && selected && !correct && (
        <svg className={styles.optionX} width="18" height="18" viewBox="0 0 24 24" fill="none">
          <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
          <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
        </svg>
      )}
    </button>
  );
}

function OpenAnswerInput({ onSubmit, submitted, isCorrect, showResults }: {
  onSubmit: (text: string) => void;
  submitted: boolean;
  isCorrect?: boolean;
  showResults?: boolean;
}) {
  const [value, setValue] = useState('');
  return (
    <div className={styles.openAnswerWrap}>
      <input
        className={[
          styles.openAnswerField,
          
          submitted && showResults
            ? (isCorrect ? styles.openCorrect : styles.openWrong)
            : submitted
              ? styles.openNeutral
              : '',
        ].join(' ')}
        type="text"
        placeholder="Введите ваш ответ..."
        value={value}
        onChange={(e) => setValue(e.target.value)}
        disabled={submitted}
        onKeyDown={(e) => { if (e.key === 'Enter' && value.trim()) onSubmit(value.trim()); }}
        autoFocus
      />
      {!submitted && (
        <button
          className={styles.openSubmitBtn}
          onClick={() => { if (value.trim()) onSubmit(value.trim()); }}
          disabled={!value.trim()}
        >
          Ответить →
        </button>
      )}

    </div>
  );
}

const THEME_COLORS: Record<string, { primary: string; light: string; soft: string }> = {
  purple: { primary: '#7c3aed', light: '#ede9fe', soft: '#f5f3ff' },
  red:    { primary: '#ef4444', light: '#fee2e2', soft: '#fef2f2' },
  orange: { primary: '#f59e0b', light: '#fef3c7', soft: '#fffbeb' },
  blue:   { primary: '#3b82f6', light: '#dbeafe', soft: '#eff6ff' },
};

const LS_QUESTION_KEY = (sessionId: string) => `sb_current_question_${sessionId}`;

interface StoredQuestionState {
  question: Question;
  questionIndex: number;
  totalQuestions: number;
  themeColor: string;
  participantId: string;
  startedAt: number; 
}

function saveQuestionState(sessionId: string, state: StoredQuestionState) {
  try { localStorage.setItem(LS_QUESTION_KEY(sessionId), JSON.stringify(state)); } catch {  }
}

function loadQuestionState(sessionId: string): StoredQuestionState | null {
  try {
    const raw = localStorage.getItem(LS_QUESTION_KEY(sessionId));
    return raw ? JSON.parse(raw) as StoredQuestionState : null;
  } catch { return null; }
}

export default function QuestionPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  
  const stateData = location.state as {
    participantId?: string;
    questionIndex?: number;
    totalQuestions?: number;
    question?: Question;
    themeColor?: string;
    mode?: string;
  } | null;

  
  const [pageState, setPageState] = useState<StoredQuestionState | null>(() => {
    if (stateData?.question) {
      return {
        question: stateData.question,
        questionIndex: stateData.questionIndex ?? 0,
        totalQuestions: stateData.totalQuestions ?? 0,
        themeColor: stateData.themeColor ?? 'purple',
        participantId: stateData.participantId ?? '',
        startedAt: Date.now(),
      };
    }
    
    return sessionId ? loadQuestionState(sessionId) : null;
  });

  const participantId = pageState?.participantId ?? '';
  const questionIndex = pageState?.questionIndex ?? 0;
  const totalQuestions = pageState?.totalQuestions ?? 0;
  const themeColor = pageState?.themeColor ?? 'purple';
  const currentQuestion = pageState?.question ?? null;

  
  useEffect(() => {
    if (pageState && sessionId) {
      saveQuestionState(sessionId, pageState);
    }
  }, [pageState, sessionId]);

  const theme = THEME_COLORS[themeColor] ?? THEME_COLORS.purple;

  
  const [timeLeft, setTimeLeft] = useState(() => {
    if (!pageState) return 0;
    const elapsed = Math.floor((Date.now() - pageState.startedAt) / 1000);
    return Math.max(0, (pageState.question.timeLimitSeconds ?? 30) - elapsed);
  });

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const startTimer = useCallback((seconds: number) => {
    if (timerRef.current) clearInterval(timerRef.current);
    
    setTimeLeft(seconds);
    timerRef.current = setInterval(() => {
      setTimeLeft((t) => {
        if (t <= 1) { clearInterval(timerRef.current!); return 0; }
        return t - 1;
      });
    }, 1000);
  }, []);

  useEffect(() => {
    if (currentQuestion) {
      
      startTimer(currentQuestion.timeLimitSeconds);
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [currentQuestion?.id]); 

  
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [submitted, setSubmitted] = useState(false);
  const [openIsCorrect, setOpenIsCorrect] = useState(false);
  const [score, setScore] = useState(0);
  const [correctAnswers, setCorrectAnswers] = useState(0);
  const [totalAnswered, setTotalAnswered] = useState(0);
  const [showResults, setShowResults] = useState(false);
  const [lastAnswerCorrect, setLastAnswerCorrect] = useState<boolean | null>(null);
  const [nextQuestionCountdown, setNextQuestionCountdown] = useState(0);
  const nextTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const answerStartTime = useRef(Date.now());
  const [pin, setPin] = useState('');
  const [wsErrorMsg, setWsErrorMsg] = useState<string | null>(null);

  
  const scoreRef = useRef(0);
  const correctRef = useRef(0);
  const totalRef = useRef(0);
  const pendingScoreRef = useRef(0);
  const pendingCorrectRef = useRef(false);
  
  const isStudentPacedRef = useRef(false);
  const answeredThisQuestionRef = useRef(false);

  
  useEffect(() => {
    scoreRef.current = score;
    sessionStorage.setItem('sb_score', String(score));
  }, [score]);
  useEffect(() => {
    correctRef.current = correctAnswers;
    sessionStorage.setItem('sb_correct', String(correctAnswers));
  }, [correctAnswers]);
  useEffect(() => {
    totalRef.current = totalAnswered;
    sessionStorage.setItem('sb_total', String(totalAnswered));
  }, [totalAnswered]);

  const nickname = sessionStorage.getItem('sb_nickname') ?? '';

  
  const [isStudentPaced, setIsStudentPaced] = useState(() => {
    if (stateData?.mode) return stateData.mode === 'student_paced';
    try {
      const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; mode: string }[];
      const sess = sessions.find((s) => s.id === sessionId);
      return sess?.mode === 'student_paced';
    } catch { return false; }
  });
  
  isStudentPacedRef.current = isStudentPaced;

  
  useEffect(() => {
    if (!sessionId) return;
    try {
      const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; pin: string; mode: string }[];
      const sess = sessions.find((s) => s.id === sessionId);
      if (sess) {
        setPin(sess.pin);
        setIsStudentPaced(sess.mode === 'student_paced');
      }
    } catch {  }
  }, [sessionId]);

  
  useEffect(() => {
    if (!sessionId) return;
    const mountedAt = Date.now();
    const check = () => {
      
      if (Date.now() - mountedAt < 3000) return;
      try {
        const sessions = JSON.parse(localStorage.getItem('sb_mock_sessions') || '[]') as { id: string; status: string }[];
        const sess = sessions.find((s) => s.id === sessionId);
        if (sess?.status === 'finished') {
          navigate(`/session/${sessionId}/finished`, {
            state: {
              participantId,
              score: scoreRef.current || Number(sessionStorage.getItem('sb_score') ?? 0),
              correctAnswers: correctRef.current || Number(sessionStorage.getItem('sb_correct') ?? 0),
              totalAnswered: totalRef.current || Number(sessionStorage.getItem('sb_total') ?? 0),
            },
          });
        }
      } catch {  }
    };
    const t = setInterval(check, 1500);
    return () => clearInterval(t);
  }, [sessionId, navigate, participantId]); 

  
  const wsConnected = useRef(false);

  useEffect(() => {
    if (!sessionId) return;

    
    if (!wsConnected.current) {
      wsConnected.current = true;
      wsService.connect(sessionId, {
        roomCode: sessionStorage.getItem('sb_pin') ?? '',
        name: sessionStorage.getItem('sb_nickname') ?? '',
        participantId: participantId || undefined,
      });
    }

    wsService.on<WsQuestionStartedPayload>('question_started', (payload) => {
      const newState: StoredQuestionState = {
        question: payload.question,
        questionIndex: payload.questionIndex,
        totalQuestions: payload.totalQuestions,
        themeColor,
        participantId,
        startedAt: payload.startedAt ?? Date.now(),
      };
      setPageState(newState);
      setSelectedIds([]);
      setSubmitted(false);
      setOpenIsCorrect(false);
      setShowResults(false);
      setLastAnswerCorrect(null);
      setNextQuestionCountdown(0);
      if (nextTimerRef.current) clearInterval(nextTimerRef.current);
      pendingScoreRef.current = 0;
      pendingCorrectRef.current = false;
      answeredThisQuestionRef.current = false;
      answerStartTime.current = Date.now();
      
      const elapsed = payload.startedAt != null
        ? Math.round((Date.now() - payload.startedAt) / 1000)
        : 0;
      const syncedTime = Math.max(1, payload.question.timeLimitSeconds - elapsed);
      startTimer(syncedTime);
      
      navigate(`/session/${sessionId}/question`, {
        replace: true,
        state: {
          participantId,
          questionIndex: payload.questionIndex,
          totalQuestions: payload.totalQuestions,
          question: payload.question,
          themeColor,
          mode: isStudentPacedRef.current ? 'student_paced' : 'teacher_paced',
        },
      });
    });

    wsService.on<WsQuestionEndedPayload & { endedAt?: number }>('question_ended', (endPayload) => {
      if (timerRef.current) clearInterval(timerRef.current);
      if (pendingScoreRef.current > 0) {
        const pending = pendingScoreRef.current;
        setScore((s) => {
          const ns = s + pending;
          scoreRef.current = ns;
          sessionStorage.setItem('sb_score', String(ns));
          return ns;
        });
      }
      if (pendingCorrectRef.current) {
        setCorrectAnswers((c) => {
          const nc = c + 1;
          correctRef.current = nc;
          sessionStorage.setItem('sb_correct', String(nc));
          return nc;
        });
      }
      pendingScoreRef.current = 0;
      pendingCorrectRef.current = false;
      
      if (!answeredThisQuestionRef.current) {
        setTotalAnswered((t) => {
          const newT = t + 1;
          totalRef.current = newT;
          sessionStorage.setItem('sb_total', String(newT));
          return newT;
        });
      }
      setLastAnswerCorrect((prev) => prev); 
      setSubmitted(true);
      setShowResults(true);
      
      if (isStudentPacedRef.current) {
        const elapsedSinceEnd = endPayload?.endedAt
          ? Math.floor((Date.now() - endPayload.endedAt) / 1000)
          : 0;
        const syncedCountdown = Math.max(1, 15 - elapsedSinceEnd);
        if (nextTimerRef.current) clearInterval(nextTimerRef.current);
        setNextQuestionCountdown(syncedCountdown);
        nextTimerRef.current = setInterval(() => {
          setNextQuestionCountdown((t) => {
            if (t <= 1) { clearInterval(nextTimerRef.current!); return 0; }
            return t - 1;
          });
        }, 1000);
      }
    });

    wsService.on<{ code: string; message: string }>('error', (payload) => {
      setWsErrorMsg(payload.message);
    });

    wsService.on<{ quiz_title: string; total_questions: number }>('joined', (_payload) => {});

    wsService.on<{ correct: boolean; score: number; total_score: number }>('answer_result', (result) => {
      pendingScoreRef.current = result.score;
      pendingCorrectRef.current = result.correct;
      setLastAnswerCorrect(result.correct);
    });

    wsService.on('session_finished', () => {
      navigate(`/session/${sessionId}/finished`, {
        state: {
          participantId,
          score: scoreRef.current > 0 ? scoreRef.current : Number(sessionStorage.getItem('sb_score') ?? 0),
          correctAnswers: correctRef.current > 0 ? correctRef.current : Number(sessionStorage.getItem('sb_correct') ?? 0),
          totalAnswered: totalRef.current || Number(sessionStorage.getItem('sb_total') ?? 0),
        },
      });
    });

    return () => {
      wsService.disconnect();
      wsConnected.current = false;
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [sessionId]); 

  
  const calcScore = (timeSpentMs: number, timeLimitMs: number) =>
    500 + Math.max(0, Math.round(500 * (1 - timeSpentMs / timeLimitMs)));

  
  const applyStudentPacedResult = () => {
    if (timerRef.current) clearInterval(timerRef.current);
    if (pendingScoreRef.current > 0) {
      setScore((s) => { const ns = s + pendingScoreRef.current; scoreRef.current = ns; return ns; });
    }
    if (pendingCorrectRef.current) {
      setCorrectAnswers((c) => { const nc = c + 1; correctRef.current = nc; return nc; });
    }
    pendingScoreRef.current = 0;
    pendingCorrectRef.current = false;
    setShowResults(true);
    
    
  };

  
  const notifyAnswered = () => {
    if (!sessionId) return;
    try {
      const key = `sb_answered_${sessionId}_${questionIndex}`;
      const prev = Number(localStorage.getItem(key) ?? 0);
      localStorage.setItem(key, String(prev + 1));
    } catch {  }
  };

  const handleSelectOption = (optionId: string) => {
    if (submitted || !currentQuestion) return;
    const isMulti = currentQuestion.type === 'multiple_select';

    if (isMulti) {
      setSelectedIds((prev) =>
        prev.includes(optionId) ? prev.filter((id) => id !== optionId) : [...prev, optionId]
      );
      return;
    }

    
    setSelectedIds([optionId]);
    const isCorrect = currentQuestion.options.find((o) => o.id === optionId)?.isCorrect ?? false;
    const timeSpent = Date.now() - answerStartTime.current;
    
    pendingScoreRef.current = isCorrect ? calcScore(timeSpent, currentQuestion.timeLimitSeconds * 1000) : 0;
    pendingCorrectRef.current = isCorrect;
    setTotalAnswered((t) => { const n = t+1; totalRef.current = n; return n; });
    answeredThisQuestionRef.current = true;
    setLastAnswerCorrect(isCorrect);
    notifyAnswered();
    wsService.send('answer', { question_id: currentQuestion.id, answer_id: optionId });
    api.sessions.submitAnswer({
      sessionId: sessionId!,
      participantId,
      questionId: currentQuestion.id,
      answerId: optionId,
      timeSpentMs: timeSpent,
    }).catch(() => {  });

    setSubmitted(true);

    if (isStudentPaced) {
      setTimeout(applyStudentPacedResult, 800);
    }
  };

  const handleSubmitMulti = () => {
    if (!currentQuestion || selectedIds.length === 0) return;
    const timeSpent = Date.now() - answerStartTime.current;
    const allCorrect = selectedIds.every((id) =>
      currentQuestion.options.find((o) => o.id === id)?.isCorrect
    );
    pendingScoreRef.current = allCorrect ? calcScore(timeSpent, currentQuestion.timeLimitSeconds * 1000) : 0;
    pendingCorrectRef.current = allCorrect;
    setTotalAnswered((t) => { const n = t+1; totalRef.current = n; return n; });
    answeredThisQuestionRef.current = true;
    setLastAnswerCorrect(allCorrect);
    notifyAnswered();
    wsService.send('answer', { question_id: currentQuestion.id, answer_id: selectedIds.join(',') });
    api.sessions.submitAnswer({
      sessionId: sessionId!,
      participantId,
      questionId: currentQuestion.id,
      answerId: selectedIds.join(','),
      timeSpentMs: timeSpent,
    }).catch(() => {  });
    setSubmitted(true);

    if (isStudentPaced) {
      setTimeout(applyStudentPacedResult, 800);
    }
  };

  const handleOpenSubmit = (text: string) => {
    if (!currentQuestion) return;
    const correct = currentQuestion.options.some(
      (o) => o.isCorrect && o.text.trim().toLowerCase() === text.trim().toLowerCase()
    );
    setOpenIsCorrect(correct);
    setLastAnswerCorrect(correct);
    const timeSpent = Date.now() - answerStartTime.current;
    pendingScoreRef.current = correct ? calcScore(timeSpent, currentQuestion.timeLimitSeconds * 1000) : 0;
    pendingCorrectRef.current = correct;
    setTotalAnswered((t) => { const n = t+1; totalRef.current = n; return n; });
    answeredThisQuestionRef.current = true;
    notifyAnswered();
    wsService.send('answer', { question_id: currentQuestion.id, answer_id: text });
    api.sessions.submitAnswer({
      sessionId: sessionId!,
      participantId,
      questionId: currentQuestion.id,
      answerId: text,
      timeSpentMs: timeSpent,
    }).catch(() => {  });
    setSubmitted(true);

    if (isStudentPaced) {
      setTimeout(applyStudentPacedResult, 800);
    }
  };

  
  if (wsErrorMsg) {
    return (
      <div className={styles.loading}>
        <p style={{ color: '#ef4444', fontWeight: 600 }}>Ошибка подключения</p>
        <p>{wsErrorMsg}</p>
        <button onClick={() => navigate('/')}>На главную</button>
      </div>
    );
  }

  if (!currentQuestion) {
    return (
      <div className={styles.loading}>
        <div className={styles.loadingSpinner} />
        <p>Ожидание вопроса...</p>
      </div>
    );
  }

  const isMulti = currentQuestion.type === 'multiple_select';
  const isOpen = currentQuestion.type === 'open_answer';

  return (
    <div
      className={styles.page}
      style={{
        '--color-primary': theme.primary,
        '--color-primary-light': theme.light,
        '--color-primary-soft': theme.soft,
      } as React.CSSProperties}
    >
      <div className={styles.topBar}>
        <span className={styles.questionNum}>
          Вопрос {questionIndex + 1} из {totalQuestions}
        </span>
        <span className={styles.brandName}>Smart Battle</span>
        <div className={styles.scoreBox}>
          <span className={styles.scoreLabel}>СЧЁТ</span>
          <span className={styles.scoreValue}>{score.toLocaleString('ru-RU')}</span>
        </div>
      </div>
      <div className={styles.timerWrap}>
        <TimerRing seconds={timeLeft} total={currentQuestion.timeLimitSeconds} />
      </div>
      <div className={styles.questionWrap}>
        <h2 className={styles.questionText}>{currentQuestion.text}</h2>
        {currentQuestion.imageUrl && (
          <img src={currentQuestion.imageUrl} alt="question media" className={styles.questionImage} />
        )}
      </div>
      {isOpen ? (
        <div className={styles.answersWrap}>
          <OpenAnswerInput
            onSubmit={handleOpenSubmit}
            submitted={submitted}
            isCorrect={showResults ? openIsCorrect : undefined}
            showResults={showResults}
          />
        </div>
      ) : (
        <>
          <div className={[styles.answersWrap, styles.answersGrid].join(' ')}>
            {currentQuestion.options.map((opt) => (
              <AnswerOption
                key={opt.id}
                text={opt.text}
                color={opt.color}
                selected={selectedIds.includes(opt.id)}
                correct={opt.isCorrect}
                submitted={submitted}
                showCorrect={showResults}
                onClick={() => handleSelectOption(opt.id)}
              />
            ))}
          </div>

          {isMulti && !submitted && selectedIds.length > 0 && (
            <div className={styles.multiSubmit}>
              <button className={styles.multiSubmitBtn} onClick={handleSubmitMulti}>
                Подтвердить ответ ({selectedIds.length}) →
              </button>
            </div>
          )}
        </>
      )}
      {submitted && !showResults && (
        <div className={styles.answeredBadge}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
            <polyline points="8 12 11 15 16 9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          Ожидание завершения вопроса...
        </div>
      )}
      {showResults && (
        isStudentPaced && nextQuestionCountdown > 0 ? (
          <div className={styles.autoAdvanceStudent}>
            <div className={styles.autoAdvanceRing}>
              <svg width="44" height="44" viewBox="0 0 44 44">
                <circle cx="22" cy="22" r="18" fill="none" stroke="var(--color-primary-light)" strokeWidth="3"/>
                <circle cx="22" cy="22" r="18" fill="none" stroke="var(--color-primary)" strokeWidth="3"
                  strokeDasharray={`${(nextQuestionCountdown / 15) * 113} 113`}
                  strokeLinecap="round" transform="rotate(-90 22 22)"
                  style={{ transition: 'stroke-dasharray 1s linear' }}
                />
              </svg>
              <span className={styles.autoAdvanceNum}>{nextQuestionCountdown}</span>
            </div>
            <span className={styles.autoAdvanceLabel}>
              {questionIndex + 1 >= totalQuestions
                ? 'Квиз завершается...'
                : 'Ожидание следующего вопроса'}
            </span>
          </div>
        ) : (
          <div className={styles.answeredBadge}>
            <div className={styles.waitingDot} />
            {questionIndex + 1 >= totalQuestions
              ? 'Ожидание завершения квиза...'
              : 'Ожидание следующего вопроса'}
          </div>
        )
      )}
      <div className={styles.footer}>
        <span>{pin ? `PIN: ${pin.slice(0, 3)} ${pin.slice(3)}` : ''}</span>
        <span>{nickname}</span>
      </div>
    </div>
  );
}
