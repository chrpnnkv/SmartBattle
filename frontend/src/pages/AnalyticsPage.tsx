import { useEffect, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout/AppLayout';
import Button from '../components/ui/Button/Button';
import Modal from '../components/ui/Modal/Modal';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { startSession, endSession, questionEnded, participantJoined, sessionFinished } from '../store/slices/sessionSlice';
import { fetchSession } from '../store/slices/sessionSlice';
import { fetchQuizById } from '../store/slices/quizSlice';
import { wsService } from '../api/wsService';
import type { WsQuestionEndedPayload, WsParticipantJoinedPayload, QuestionReport, Question, QuestionType } from '../types';
import styles from './AnalyticsPage.module.css';

const IconClock = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2"/>
    <polyline points="12 7 12 12 15 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
);

const IconWarning = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    <line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
    <line x1="12" y1="17" x2="12.01" y2="17" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
  </svg>
);

interface DistributionBarProps {
  report: QuestionReport;
  questionType: QuestionType;
}

function DistributionBar({ report, questionType }: DistributionBarProps) {
  const max = Math.max(...report.distribution.map((d) => d.count), 1);

  
  if (questionType === 'open_answer') {
    const correctCount = report.distribution.filter((d) => d.isCorrect).reduce((s, d) => s + d.count, 0);
    const wrongCount = report.distribution.filter((d) => !d.isCorrect).reduce((s, d) => s + d.count, 0);
    const total = correctCount + wrongCount || 1;
    const bars = [
      { label: 'Верно', count: correctCount, isCorrect: true, color: '#22c55e' },
      { label: 'Неверно', count: wrongCount, isCorrect: false, color: '#f87171' },
    ];
    return (
      <div className={styles.distChart}>
        {bars.map((b) => (
          <div key={b.label} className={styles.distBar}>
            <div className={styles.distBarTrack}>
              <div
                className={styles.distBarFill}
                style={{ height: `${(b.count / total) * 100}%`, background: b.color }}
              />
            </div>
            <span className={styles.distBarLabel}>{b.label}</span>
            <span className={styles.distBarCount}>{b.count}</span>
          </div>
        ))}
      </div>
    );
  }

  
  
  const bars = questionType === 'true_false'
    ? report.distribution.slice(0, 2)
    : report.distribution;

  return (
    <div className={styles.distChart}>
      {bars.map((d) => (
        <div key={d.optionId} className={styles.distBar}>
          <div className={styles.distBarTrack}>
            <div
              className={styles.distBarFill}
              style={{
                height: `${(d.count / max) * 100}%`,
                background: d.isCorrect ? '#22c55e' : '#f87171',
              }}
            />
          </div>
          <span className={styles.distBarLabel}>
            {questionType === 'true_false' ? d.optionText : d.optionText.charAt(0).toUpperCase()}
          </span>
          <span className={styles.distBarCount}>{d.count}</span>
        </div>
      ))}
    </div>
  );
}

function getMostCommonError(report: QuestionReport, questionType: QuestionType): string {
  if (questionType === 'open_answer') {
    
    const wrong = report.distribution
      .filter((d) => !d.isCorrect)
      .sort((a, b) => b.count - a.count)[0];
    return wrong ? `"${wrong.optionText}" (${wrong.count} чел.)` : '—';
  }
  return report.mostCommonWrongOptionText ?? '—';
}

function LobbyView({ pin, participantCount, quizTitle, onStart, isLoading }: {
  pin: string; participantCount: number; quizTitle: string;
  onStart: () => void; isLoading: boolean;
}) {
  const pinFormatted = pin.slice(0, 3) + ' ' + pin.slice(3);
  return (
    <div className={styles.lobby}>
      <div className={styles.lobbyCard}>
        <h1 className={styles.lobbyTitle}>{quizTitle}</h1>
        <p className={styles.lobbySubtitle}>Ожидание подключения студентов</p>
        <div className={styles.pinBlock}>
          <span className={styles.pinLabel}>PIN-код для входа</span>
          <span className={styles.pinValue}>{pinFormatted}</span>
          <span className={styles.pinHint}>Студенты вводят этот код на главной странице</span>
        </div>
        <div className={styles.participantCount}>
          <div className={styles.participantDot} />
          <span>{participantCount} студентов подключилось</span>
        </div>
        <Button size="lg" onClick={onStart} isLoading={isLoading}>
          Начать квиз →
        </Button>
      </div>
    </div>
  );
}

function TimerCircle({ seconds, total }: { seconds: number; total: number }) {
  const r = 22;
  const circ = 2 * Math.PI * r;
  const pct = seconds / total;
  const dash = pct * circ;
  const color = seconds <= 5 ? '#ef4444' : seconds <= 10 ? '#f59e0b' : '#7c3aed';
  return (
    <div className={styles.timerCircle}>
      <svg width="60" height="60" viewBox="0 0 60 60">
        <circle cx="30" cy="30" r={r} fill="none" stroke="#e4e4f0" strokeWidth="4"/>
        <circle
          cx="30" cy="30" r={r} fill="none"
          stroke={color} strokeWidth="4"
          strokeDasharray={`${dash} ${circ}`}
          strokeLinecap="round"
          transform="rotate(-90 30 30)"
          style={{ transition: 'stroke-dasharray 1s linear, stroke 0.3s' }}
        />
      </svg>
      <span className={styles.timerValue} style={{ color }}>{seconds}</span>
    </div>
  );
}

export default function AnalyticsPage() {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { session, isLoading } = useAppSelector((s) => s.session);
  const { currentQuiz } = useAppSelector((s) => s.quiz);

  
  const [report, setReport] = useState<QuestionReport | null>(null);
  const [questionIdx, setQuestionIdx] = useState(0);
  const [phase, setPhase] = useState<'lobby' | 'question_active' | 'question_results' | 'finished'>('lobby');
  const [participantCount, setParticipantCount] = useState(0);
  const [answeredCount, setAnsweredCount] = useState(0);
  const [timeLeft, setTimeLeft] = useState(0);
  const [totalTime, setTotalTime] = useState(30);
  const [resultsCountdown, setResultsCountdown] = useState(0);
  const [showQuestionModal, setShowQuestionModal] = useState(false);

  
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const resultsTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const answeredPollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const questionIdxRef = useRef(0);
  const wsConnected = useRef(false);

  
  const currentQuestion: Question | undefined = currentQuiz?.questions[questionIdx];
  const totalQuestions = currentQuiz?.questions.length ?? 0;
  
  const isStudentPaced = session?.mode === 'student_paced';

  
  useEffect(() => {
    if (!sessionId) return;
    
    if (!session || session.id !== sessionId) {
      dispatch(fetchSession(sessionId));
      return;
    }
    
    if (session.quizId && (!currentQuiz || currentQuiz.id !== session.quizId)) {
      dispatch(fetchQuizById(session.quizId));
    }
  }, [sessionId, session, currentQuiz, dispatch]);

  
  useEffect(() => { questionIdxRef.current = questionIdx; }, [questionIdx]);

  
  const startTimer = (seconds: number) => {
    setTotalTime(seconds);
    setTimeLeft(seconds);
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => {
      setTimeLeft((t) => {
        if (t <= 1) {
          clearInterval(timerRef.current!);
          if (answeredPollRef.current) clearInterval(answeredPollRef.current);
          wsService.send('end_question', { questionIndex: questionIdxRef.current });
          return 0;
        }
        return t - 1;
      });
    }, 1000);
  };

  const startResultsCountdown = (nextAction: () => void) => {
    setResultsCountdown(15);
    if (resultsTimerRef.current) clearInterval(resultsTimerRef.current);
    resultsTimerRef.current = setInterval(() => {
      setResultsCountdown((t) => {
        if (t <= 1) {
          clearInterval(resultsTimerRef.current!);
          nextAction();
          return 0;
        }
        return t - 1;
      });
    }, 1000);
  };

  const startAnsweredPoll = (sId: string, totalParticipants: number) => {
    if (answeredPollRef.current) clearInterval(answeredPollRef.current);
    answeredPollRef.current = setInterval(() => {
      try {
        const raw = localStorage.getItem(`sb_answered_${sId}_${questionIdxRef.current}`);
        const count = raw ? Number(raw) : 0;
        setAnsweredCount(count);
        if (count >= totalParticipants && totalParticipants > 0 && isStudentPaced) {
          clearInterval(answeredPollRef.current!);
          if (timerRef.current) clearInterval(timerRef.current);
          wsService.send('end_question', { questionIndex: questionIdxRef.current });
        }
      } catch {  }
    }, 1000);
  };

  
  const handleNextQuestion = () => {
    if (!session || !currentQuiz) return;
    const nextIdx = questionIdx + 1;
    if (nextIdx >= currentQuiz.questions.length) {
      dispatch(endSession(session.id));
      setPhase('finished');
      return;
    }
    const nextQ = currentQuiz.questions[nextIdx];
    setQuestionIdx(nextIdx);
    questionIdxRef.current = nextIdx;
    setReport(null);
    setAnsweredCount(0);
    setPhase('question_active');
    wsService.send('start_question', {
      question: nextQ,
      questionIndex: nextIdx,
      totalQuestions: currentQuiz.questions.length,
    });
    startTimer(nextQ.timeLimitSeconds ?? 30);
    startAnsweredPoll(session.id, session.participants.length);
  };

  const handleStart = async () => {
    if (!session || !currentQuiz) return;
    await dispatch(startSession(session.id));
    const firstQ = currentQuiz.questions[0];
    setPhase('question_active');
    setQuestionIdx(0);
    questionIdxRef.current = 0;
    setAnsweredCount(0);
    wsService.send('start_question', {
      question: firstQ,
      questionIndex: 0,
      totalQuestions: currentQuiz.questions.length,
    });
    startTimer(firstQ?.timeLimitSeconds ?? 30);
    startAnsweredPoll(session.id, session.participants.length);
  };

  const handleEndQuestion = () => {
    if (timerRef.current) clearInterval(timerRef.current);
    if (answeredPollRef.current) clearInterval(answeredPollRef.current);
    wsService.send('end_question', { questionIndex: questionIdxRef.current });
  };

  
  useEffect(() => {
    if (!session || wsConnected.current) return;
    wsConnected.current = true;
    wsService.connect(session.id);

    wsService.on<WsParticipantJoinedPayload>('participant_joined', (payload) => {
      dispatch(participantJoined(payload.participant));
      setParticipantCount(payload.totalCount);
    });

    wsService.on<WsQuestionEndedPayload>('question_ended', (payload) => {
      dispatch(questionEnded(payload.questionReport));
      setReport(payload.questionReport);
      setPhase('question_results');
      if (timerRef.current) clearInterval(timerRef.current);
      if (answeredPollRef.current) clearInterval(answeredPollRef.current);
      setAnsweredCount(0);
    });

    wsService.on('session_finished', () => {
      dispatch(sessionFinished());
      setPhase('finished');
    });

    return () => {
      wsService.disconnect();
      wsConnected.current = false;
      if (timerRef.current) clearInterval(timerRef.current);
      if (answeredPollRef.current) clearInterval(answeredPollRef.current);
      if (resultsTimerRef.current) clearInterval(resultsTimerRef.current);
    };
  }, [session, dispatch]);

  
  useEffect(() => {
    if (!isStudentPaced || phase !== 'question_results') return;
    startResultsCountdown(handleNextQuestion);
    return () => {
      if (resultsTimerRef.current) clearInterval(resultsTimerRef.current);
    };
  }, [phase, isStudentPaced]); 

  

  if (phase === 'lobby' && session) {
    return (
      <AppLayout>
        <LobbyView
          pin={session.pin}
          participantCount={participantCount}
          quizTitle={currentQuiz?.title ?? 'Квиз'}
          onStart={handleStart}
          isLoading={isLoading}
        />
      </AppLayout>
    );
  }

  if (phase === 'finished') {
    return (
      <AppLayout>
        <div className={styles.finishedScreen}>
          <div className={styles.finishedCard}>
            <div className={styles.finishedIcon}>
              <svg width="40" height="40" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <polyline points="8 12 11 15 16 9" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <h2 className={styles.finishedTitle}>Квиз завершён!</h2>
            <p className={styles.finishedSub}>Результаты сохранены в отчётах</p>
            <div className={styles.finishedActions}>
              <Button variant="ghost" onClick={() => navigate('/dashboard')}>
                ← В личный кабинет
              </Button>
              <Button size="lg" onClick={() => navigate('/reports')}>
                Смотреть отчёт →
              </Button>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className={styles.page}>
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <span className={styles.logoText}>Smart Battle</span>
            <span className={styles.analyticsTag}>Аналитика</span>
            {isStudentPaced && (
              <span className={styles.studentPacedTag}>Свободный темп</span>
            )}
          </div>
          <div className={styles.headerRight}>
            <span className={styles.questionCounter}>Вопрос {questionIdx + 1}/{totalQuestions}</span>
            <span className={styles.quizName}>{currentQuiz?.title}</span>
          </div>
        </div>

        {phase === 'question_active' && (
          <div className={styles.activeQuestion}>
            <div className={styles.waitingCard}>
              <TimerCircle seconds={timeLeft} total={totalTime} />
              <p className={styles.waitingText}>
                {isStudentPaced ? 'Студенты отвечают в своём темпе' : 'Ожидание ответов студентов...'}
              </p>
              <div className={styles.questionPreviewBox}>
                <p className={styles.questionPreviewText}>{currentQuestion?.text}</p>
              </div>
              <p className={styles.waitingCount}>
                <strong>{answeredCount}</strong> из <strong>{participantCount}</strong> ответили
              </p>
              {!isStudentPaced && (
                <Button onClick={handleEndQuestion} variant="secondary">
                  Завершить вопрос досрочно
                </Button>
              )}
            </div>
          </div>
        )}

        {phase === 'question_results' && report && (
          <div className={styles.results}>
            <div className={styles.statsRow}>
              <div className={styles.statCard}>
                <span className={styles.statLabel}>Процент верного ответа</span>
                <span className={styles.statValue}>{report.correctPercent}%</span>
              </div>
              <div className={[styles.statCard, styles.statCardBlue].join(' ')}>
                <div className={styles.statIcon}><IconClock /></div>
                <span className={styles.statLabel}>Среднее время ответа</span>
                <span className={styles.statValue}>{(report.avgResponseTimeMs / 1000).toFixed(1)} сек</span>
              </div>
              <div className={[styles.statCard, styles.statCardRed].join(' ')}>
                <div className={styles.statIcon}><IconWarning /></div>
                <span className={styles.statLabel}>
                  {currentQuestion?.type === 'open_answer' ? 'Частый неверный ответ' : 'Частая ошибка'}
                </span>
                <span className={[styles.statValue, styles.statValueError].join(' ')}>
                  {getMostCommonError(report, currentQuestion?.type ?? 'multiple_choice')}
                </span>
              </div>
            </div>

            <div className={styles.mainGrid}>
              <div className={styles.chartCard}>
                <h3 className={styles.chartTitle}>Распределение ответов</h3>
                <p className={styles.chartSub}>Для всех {participantCount} участников в реальном времени</p>
                {currentQuestion?.type !== 'open_answer' && (
                  <div className={styles.chartLegend}>
                    <span><span className={styles.legendDot} style={{ background: '#22c55e' }} />Верно</span>
                    <span><span className={styles.legendDot} style={{ background: '#f87171' }} />Неверно</span>
                  </div>
                )}
                <DistributionBar
                  report={report}
                  questionType={currentQuestion?.type ?? 'multiple_choice'}
                />
              </div>

              <div className={styles.leaderCard}>
                <h3 className={styles.leaderTitle}>Самые быстрые верные ответы:</h3>
                <div className={styles.leaderList}>
                  {report.fastestCorrectParticipants.slice(0, 5).map((p, i) => (
                    <div key={p.id} className={styles.leaderRow}>
                      <span className={styles.leaderRank}>{i + 1}</span>
                      <span className={styles.leaderName}>{p.nickname}</span>
                    </div>
                  ))}
                </div>
                <p className={styles.leaderHint}>Таблица лидеров обновляется после каждого вопроса</p>
              </div>
            </div>

            <div className={styles.actions}>
              <Button variant="ghost" onClick={() => setShowQuestionModal(true)}>
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" strokeWidth="2"/>
                  <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
                </svg>
                Просмотреть вопрос
              </Button>
              {isStudentPaced ? (
                <div className={styles.autoAdvance}>
                  <div className={styles.autoAdvanceTimer}>{resultsCountdown}</div>
                  <span className={styles.autoAdvanceText}>
                    {questionIdx + 1 >= totalQuestions
                      ? 'Квиз завершится автоматически'
                      : 'Следующий вопрос автоматически'}
                  </span>
                </div>
              ) : (
                <Button size="lg" onClick={handleNextQuestion}>
                  {questionIdx + 1 >= totalQuestions ? 'Завершить квиз' : 'Следующий вопрос →'}
                </Button>
              )}
            </div>
          </div>
        )}
      </div>

      <Modal
        isOpen={showQuestionModal}
        onClose={() => setShowQuestionModal(false)}
        title={`Вопрос ${questionIdx + 1} из ${totalQuestions}`}
      >
        <div className={styles.questionModalContent}>
          {currentQuestion?.imageUrl && (
            <img src={currentQuestion.imageUrl} alt="" className={styles.questionModalImg} />
          )}
          <p className={styles.questionModalText}>{currentQuestion?.text}</p>
          <div className={styles.questionModalOptions}>
            {currentQuestion?.options.map((opt) => (
              <div
                key={opt.id}
                className={[styles.questionModalOption, opt.isCorrect ? styles.questionModalOptionCorrect : ''].join(' ')}
              >
                <span className={styles.questionModalDot} style={{
                  background: { red: '#ef4444', blue: '#3b82f6', yellow: '#f59e0b', green: '#22c55e' }[opt.color]
                }} />
                <span>{opt.text}</span>
                {opt.isCorrect && (
                  <svg className={styles.questionModalCheck} width="16" height="16" viewBox="0 0 24 24" fill="none">
                    <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                )}
              </div>
            ))}
          </div>
        </div>
      </Modal>
    </AppLayout>
  );
}
