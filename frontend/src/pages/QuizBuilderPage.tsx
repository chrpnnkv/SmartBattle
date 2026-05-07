import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import AppLayout from '../components/layout/AppLayout/AppLayout';
import Button from '../components/ui/Button/Button';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { fetchQuizById, createQuiz, updateQuiz } from '../store/slices/quizSlice';
import type { Question, QuestionType, AnswerOption, QuizSettings } from '../types';
import { api } from '../api';
import styles from './QuizBuilderPage.module.css';

const uid = () => Math.random().toString(36).slice(2, 8);

const ANSWER_COLORS: AnswerOption['color'][] = ['red', 'blue', 'yellow', 'green'];
const COLOR_MAP: Record<AnswerOption['color'], string> = {
  red: '#ef4444', blue: '#3b82f6', yellow: '#f59e0b', green: '#22c55e',
};

const TYPE_LABELS: Record<QuestionType, string> = {
  multiple_choice: 'Один ответ',
  multiple_select: 'Несколько ответов',
  true_false: 'Правда / Ложь',
  open_answer: 'Свободный ввод',
};

function makeEmptyQuestion(order: number): Question {
  return {
    id: uid(), quizId: '', type: 'multiple_choice', text: '',
    timeLimitSeconds: 30, order,
    options: ANSWER_COLORS.map((color, i) => ({ id: uid(), text: '', isCorrect: i === 0, color })),
  };
}

function QuestionCard({ question, index, isActive, onClick }: {
  question: Question; index: number; isActive: boolean; onClick: () => void;
}) {
  return (
    <button className={[styles.qCard, isActive ? styles.qCardActive : ''].join(' ')} onClick={onClick}>
      <div className={styles.qCardHeader}>
        <span className={styles.qCardNum}>Q{index + 1}</span>
        <span className={styles.qCardType}>{TYPE_LABELS[question.type]}</span>
        <span className={styles.qCardTime}>{question.timeLimitSeconds}с</span>
      </div>
      <p className={styles.qCardText}>{question.text || 'Введите вопрос...'}</p>
    </button>
  );
}

function AnswerRow({ option, isMulti, onChange, onSetCorrect }: {
  option: AnswerOption; isMulti: boolean;
  onChange: (u: AnswerOption) => void; onSetCorrect: () => void;
}) {
  return (
    <div className={[styles.answerRow, option.isCorrect ? styles.answerRowCorrect : ''].join(' ')}>
      <span className={styles.answerDot} style={{ background: COLOR_MAP[option.color] }} />
      <input
        className={styles.answerInput}
        type="text"
        placeholder="Добавить вариант ответа"
        value={option.text}
        onChange={(e) => onChange({ ...option, text: e.target.value })}
      />
      <button
        className={[styles.correctBtn, option.isCorrect ? styles.correctBtnActive : ''].join(' ')}
        onClick={onSetCorrect}
        title={isMulti ? 'Отметить как верный (можно несколько)' : 'Отметить как верный'}
        type="button"
      >
        {option.isCorrect ? (
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none">
            <polyline points="20 6 9 17 4 12" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        ) : (
          <span className={styles.correctBtnEmpty} />
        )}
      </button>
    </div>
  );
}

export default function QuizBuilderPage() {
  const { id } = useParams<{ id: string }>();
  const isEditing = Boolean(id);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { currentQuiz, isLoading } = useAppSelector((s) => s.quiz);

  const [title, setTitle] = useState('');
  const [mode, setMode] = useState<'teacher_paced' | 'student_paced'>('teacher_paced');
  const [settings, setSettings] = useState<QuizSettings>({
    shuffleQuestions: false, shuffleAnswers: false, showLeaderboard: true, themeColor: 'purple',
  });
  const [questions, setQuestions] = useState<Question[]>([makeEmptyQuestion(1)]);
  const [activeIdx, setActiveIdx] = useState(0);
  const [isSaving, setIsSaving] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [titleError, setTitleError] = useState('');

  useEffect(() => {
    if (isEditing && id) dispatch(fetchQuizById(id));
  }, [isEditing, id, dispatch]);

  useEffect(() => {
    if (isEditing && currentQuiz) {
      setTitle(currentQuiz.title);
      setMode(currentQuiz.mode);
      setSettings(currentQuiz.settings);
      setQuestions(currentQuiz.questions.length > 0 ? currentQuiz.questions : [makeEmptyQuestion(1)]);
    }
  }, [isEditing, currentQuiz]);

  const activeQuestion = questions[activeIdx];

  const addQuestion = () => {
    const newQ = makeEmptyQuestion(questions.length + 1);
    setQuestions((p) => [...p, newQ]);
    setActiveIdx(questions.length);
  };

  const updateActive = useCallback((updated: Partial<Question>) => {
    setQuestions((p) => p.map((q, i) => i === activeIdx ? { ...q, ...updated } : q));
  }, [activeIdx]);

  const deleteQuestion = () => {
    if (questions.length === 1) return;
    setQuestions((p) => p.filter((_, i) => i !== activeIdx));
    setActiveIdx((p) => Math.max(0, p - 1));
  };

  const duplicateQuestion = () => {
    const copy = { ...activeQuestion, id: uid(), order: questions.length + 1 };
    setQuestions((p) => [...p, copy]);
    setActiveIdx(questions.length);
  };

  const updateOption = (optionId: string, updated: AnswerOption) => {
    updateActive({ options: activeQuestion.options.map((o) => o.id === optionId ? updated : o) });
  };

  const toggleCorrectOption = (optionId: string) => {
    const isMulti = activeQuestion.type === 'multiple_select';
    updateActive({
      options: activeQuestion.options.map((o) =>
        isMulti
          ? o.id === optionId ? { ...o, isCorrect: !o.isCorrect } : o
          : { ...o, isCorrect: o.id === optionId }
      ),
    });
  };

  const changeType = (type: QuestionType) => {
    if (type === 'true_false') {
      updateActive({ type, options: [
        { id: uid(), text: 'Верно', isCorrect: true, color: 'green' },
        { id: uid(), text: 'Неверно', isCorrect: false, color: 'red' },
      ]});
    } else if (type === 'open_answer') {
      updateActive({ type, options: [{ id: uid(), text: '', isCorrect: true, color: 'green' }] });
    } else {
      updateActive({ type, options: ANSWER_COLORS.map((color, i) => ({
        id: uid(), text: '', isCorrect: i === 0, color,
      }))});
    }
  };

  const handleSave = async (publish = false) => {
    if (!title.trim()) { setTitleError('Введите название квиза'); return; }
    setIsSaving(true);
    const payload = {
      title, settings, mode,
      status: publish ? 'published' as const : 'draft' as const,
      questions: questions.map(({ id: _id, quizId: _qid, ...rest }) => rest),
    };
    try {
      if (isEditing && id) await dispatch(updateQuiz({ id, data: payload })).unwrap();
      else await dispatch(createQuiz(payload)).unwrap();
      navigate('/dashboard');
    } finally {
      setIsSaving(false);
    }
  };

  if (isEditing && isLoading && !currentQuiz) {
    return <AppLayout><div className={styles.loadingCenter}>Загрузка...</div></AppLayout>;
  }

  const isMulti = activeQuestion?.type === 'multiple_select';
  const isOpenAnswer = activeQuestion?.type === 'open_answer';

  return (
    <AppLayout>
      <div className={styles.builder}>
        <div className={styles.topBar}>
          <div className={styles.topBarLeft}>
            <input
              className={[styles.titleInput, titleError ? styles.titleInputError : ''].join(' ')}
              value={title}
              onChange={(e) => { setTitle(e.target.value); setTitleError(''); }}
              placeholder="Название квиза..."
            />
            {titleError && <span className={styles.titleError}>{titleError}</span>}
            <span className={styles.topBarMeta}>
              {isEditing ? 'Редактирование' : 'Новый квиз'} · {questions.length} вопр.
            </span>
          </div>
          <div className={styles.topBarActions}>
            <Button variant="ghost" onClick={() => navigate('/dashboard')}>Отмена</Button>
            <Button variant="secondary" onClick={() => handleSave(false)} isLoading={isSaving}>Сохранить черновик</Button>
            <Button onClick={() => handleSave(true)} isLoading={isSaving}>Опубликовать</Button>
          </div>
        </div>

        <div className={styles.content}>
          <aside className={styles.sidebar}>
            <Button variant="secondary" fullWidth onClick={addQuestion}>
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none">
                <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
              </svg>
              Добавить вопрос
            </Button>
            <div className={styles.questionList}>
              {questions.map((q, i) => (
                <QuestionCard key={q.id} question={q} index={i} isActive={i === activeIdx} onClick={() => setActiveIdx(i)} />
              ))}
            </div>
          </aside>
          <main className={styles.editor}>
            <div className={styles.mediaArea}>
              <span className={styles.mediaLabel}>Медиафайл к вопросу (опционально)</span>
              {activeQuestion.imageUrl ? (
                <div className={styles.mediaPreview}>
                  <img src={activeQuestion.imageUrl} alt="preview" />
                  <button className={styles.mediaRemove} onClick={() => updateActive({ imageUrl: undefined })}>
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none">
                      <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                      <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                    </svg>
                  </button>
                </div>
              ) : (
                <label className={styles.mediaUpload}>
                  <svg width="26" height="26" viewBox="0 0 24 24" fill="none">
                    <rect x="3" y="3" width="18" height="18" rx="3" stroke="currentColor" strokeWidth="1.5"/>
                    <circle cx="8.5" cy="8.5" r="1.5" fill="currentColor"/>
                    <polyline points="21 15 16 10 5 21" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span>Перетащите изображение или видео сюда</span>
                  <input
                  type="file"
                  accept="image/*"
                  className={styles.mediaInput}
                  disabled={isUploading}
                  onChange={async (e) => {
                    const file = e.target.files?.[0];
                    if (!file) return;
                    setIsUploading(true);
                    try {
                      const url = await api.quizzes.uploadImage(file);
                      updateActive({ imageUrl: url });
                    } finally {
                      setIsUploading(false);
                      e.target.value = '';
                    }
                  }}
                />
                </label>
              )}
            </div>
            <div className={styles.typeTabs}>
              {(Object.keys(TYPE_LABELS) as QuestionType[]).map((t) => (
                <button
                  key={t}
                  className={[styles.typeTab, activeQuestion.type === t ? styles.typeTabActive : ''].join(' ')}
                  onClick={() => changeType(t)}
                >
                  {TYPE_LABELS[t]}
                </button>
              ))}
            </div>
            <textarea
              className={styles.questionTextarea}
              value={activeQuestion.text}
              onChange={(e) => updateActive({ text: e.target.value })}
              placeholder="Введите ваш вопрос здесь..."
              rows={3}
            />
            <div className={styles.timerRow}>
              <label className={styles.timerLabel}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2"/>
                  <polyline points="12 7 12 12 15 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
                Таймер (сек)
              </label>
              <input
                className={styles.timerInput}
                type="number" min={5} max={300} step={5}
                value={activeQuestion.timeLimitSeconds}
                onChange={(e) => updateActive({ timeLimitSeconds: Number(e.target.value) })}
              />
              {isMulti && (
                <span className={styles.multiHint}>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                    <line x1="12" y1="8" x2="12" y2="12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                    <line x1="12" y1="16" x2="12.01" y2="16" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                  Отметьте все верные варианты
                </span>
              )}
            </div>
            {isOpenAnswer ? (
              <div className={styles.openAnswerSection}>
                <div className={styles.openAnswerHint}>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span>Студент вводит ответ текстом. Засчитывается если совпадает хотя бы с одним вариантом.</span>
                </div>
                <div className={styles.openAnswerCorrect}>
                  <label className={styles.openAnswerLabel}>Варианты правильного ответа:</label>
                  {activeQuestion.options.map((opt, optIdx) => (
                    <div key={opt.id} className={styles.openAnswerRow}>
                      <input
                        className={styles.openAnswerInput}
                        type="text"
                        placeholder={`Вариант ${optIdx + 1}...`}
                        value={opt.text}
                        onChange={(e) => updateActive({
                          options: activeQuestion.options.map((o) =>
                            o.id === opt.id ? { ...o, text: e.target.value } : o
                          )
                        })}
                      />
                      {activeQuestion.options.length > 1 && (
                        <button
                          className={styles.openAnswerRemove}
                          onClick={() => updateActive({
                            options: activeQuestion.options.filter((o) => o.id !== opt.id)
                          })}
                          type="button"
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                            <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                            <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                          </svg>
                        </button>
                      )}
                    </div>
                  ))}
                  <button
                    className={styles.openAnswerAdd}
                    onClick={() => updateActive({
                      options: [...activeQuestion.options, { id: uid(), text: '', isCorrect: true, color: 'green' }]
                    })}
                    type="button"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                      <line x1="12" y1="5" x2="12" y2="19" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                      <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"/>
                    </svg>
                    Добавить вариант
                  </button>
                  <span className={styles.openAnswerNote}>
                    Сравнение без учёта регистра и пробелов по краям
                  </span>
                </div>
              </div>
            ) : (
              <div className={styles.answersGrid}>
                {activeQuestion.options.map((option) => (
                  <AnswerRow
                    key={option.id}
                    option={option}
                    isMulti={isMulti}
                    onChange={(updated) => updateOption(option.id, updated)}
                    onSetCorrect={() => toggleCorrectOption(option.id)}
                  />
                ))}
              </div>
            )}
            <div className={styles.editorFooter}>
              <button className={styles.footerBtn} onClick={deleteQuestion} disabled={questions.length === 1}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                  <polyline points="3 6 5 6 21 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  <path d="M19 6l-1 14H6L5 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M10 11v6M14 11v6" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                </svg>
                Удалить вопрос
              </button>
              <button className={[styles.footerBtn, styles.footerBtnPrimary].join(' ')} onClick={duplicateQuestion}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                  <rect x="9" y="9" width="13" height="13" rx="2" stroke="currentColor" strokeWidth="2"/>
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" stroke="currentColor" strokeWidth="2"/>
                </svg>
                Дублировать
              </button>
            </div>
          </main>
          <aside className={styles.settings}>
            <h3 className={styles.settingsTitle}>Настройки квиза</h3>

            <div className={styles.settingsList}>
              {[
                { key: 'shuffleQuestions', label: 'Перемешать вопросы' },
                { key: 'shuffleAnswers', label: 'Перемешать ответы' },
                { key: 'showLeaderboard', label: 'Доска лидеров' },
              ].map(({ key, label }) => (
                <label key={key} className={styles.settingRow}>
                  <span>{label}</span>
                  <input
                    type="checkbox"
                    className={styles.checkbox}
                    checked={settings[key as keyof QuizSettings] as boolean}
                    onChange={(e) => setSettings((p) => ({ ...p, [key]: e.target.checked }))}
                  />
                </label>
              ))}
            </div>

            <div className={styles.settingSection}>
              <span className={styles.settingLabel}>Цвет темы</span>
              <p className={styles.settingHint}>Применяется к экрану студента во время игры</p>
              <div className={styles.colorPicker}>
                {([
                  ['purple', '#7c3aed'],
                  ['red', '#ef4444'],
                  ['orange', '#f59e0b'],
                  ['blue', '#3b82f6'],
                ] as [QuizSettings['themeColor'], string][]).map(([color, hex]) => (
                  <button
                    key={color}
                    className={[styles.colorBtn, settings.themeColor === color ? styles.colorBtnActive : ''].join(' ')}
                    style={{ background: hex }}
                    onClick={() => setSettings((p) => ({ ...p, themeColor: color }))}
                    title={color}
                  />
                ))}
              </div>
              <div className={styles.colorPreview} style={{
                background: {
                  purple: 'linear-gradient(135deg, #7c3aed22, #7c3aed11)',
                  red: 'linear-gradient(135deg, #ef444422, #ef444411)',
                  orange: 'linear-gradient(135deg, #f59e0b22, #f59e0b11)',
                  blue: 'linear-gradient(135deg, #3b82f622, #3b82f611)',
                }[settings.themeColor],
                borderColor: {
                  purple: '#7c3aed33', red: '#ef444433', orange: '#f59e0b33', blue: '#3b82f633',
                }[settings.themeColor],
              }}>
                <span className={styles.colorPreviewText} style={{
                  color: {
                    purple: '#7c3aed', red: '#ef4444', orange: '#f59e0b', blue: '#3b82f6',
                  }[settings.themeColor]
                }}>Предпросмотр темы</span>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </AppLayout>
  );
}
