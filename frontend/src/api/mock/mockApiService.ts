import type {
  IApiService,
  IAuthApi,
  IQuizApi,
  ISessionApi,
  IAnalyticsApi,
} from '../IApiService';
import { writeSignal } from '../wsService';
import type {
  AuthResponse,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  GameReport,
  GameSession,
  JoinSessionRequest,
  JoinSessionResponse,
  LoginRequest,
  Quiz,
  CreateQuizRequest,
  UpdateQuizRequest,
  RegisterRequest,
  ResetPasswordRequest,
  User,
} from '../../types';

const delay = (ms = 400) => new Promise((r) => setTimeout(r, ms));

const uid = () => Math.random().toString(36).slice(2, 10);

const MOCK_USER: User = {
  id: 'u1',
  email: 'teacher@example.com',
  name: 'Сергей М.',
  role: 'teacher',
  createdAt: '2024-09-01T10:00:00Z',
};

const MOCK_QUIZZES: Quiz[] = [
  {
    id: 'q1',
    title: 'Интегралы 2',
    description: 'Определённые и неопределённые интегралы',
    status: 'published',
    mode: 'teacher_paced',
    settings: {
      shuffleQuestions: false,
      shuffleAnswers: true,
      showLeaderboard: true,
      themeColor: 'purple',
    },
    questionCount: 15,
    estimatedMinutes: 45,
    createdAt: '2024-10-12T08:00:00Z',
    updatedAt: '2024-10-12T08:00:00Z',
    authorId: 'u1',
    questions: [
      {
        id: 'qn1',
        quizId: 'q1',
        type: 'multiple_choice',
        text: 'В каком году человек впервые ступил на поверхность Луны?',
        imageUrl: 'https://upload.wikimedia.org/wikipedia/commons/thumb/9/9c/Aldrin_Apollo_11.jpg/640px-Aldrin_Apollo_11.jpg',
        timeLimitSeconds: 30,
        order: 1,
        options: [
          { id: 'a1', text: '1963', isCorrect: false, color: 'red' },
          { id: 'a2', text: '1974', isCorrect: false, color: 'blue' },
          { id: 'a3', text: '1967', isCorrect: false, color: 'yellow' },
          { id: 'a4', text: '1969', isCorrect: true, color: 'green' },
        ],
      },
      {
        id: 'qn2',
        quizId: 'q1',
        type: 'true_false',
        text: 'Интеграл от константы равен нулю',
        timeLimitSeconds: 15,
        order: 2,
        options: [
          { id: 'b1', text: 'Верно', isCorrect: false, color: 'green' },
          { id: 'b2', text: 'Неверно', isCorrect: true, color: 'red' },
        ],
      },
    ],
  },
  {
    id: 'q2',
    title: 'Ряды 3',
    description: 'Числовые и функциональные ряды',
    status: 'published',
    mode: 'student_paced',
    settings: {
      shuffleQuestions: false,
      shuffleAnswers: false,
      showLeaderboard: false,
      themeColor: 'blue',
    },
    questionCount: 10,
    estimatedMinutes: 15,
    createdAt: '2024-10-03T09:00:00Z',
    updatedAt: '2024-10-03T09:00:00Z',
    authorId: 'u1',
    questions: [],
  },
  {
    id: 'q3',
    title: 'Пределы 1',
    description: 'Основы теории пределов',
    status: 'draft',
    mode: 'teacher_paced',
    settings: {
      shuffleQuestions: true,
      shuffleAnswers: true,
      showLeaderboard: true,
      themeColor: 'orange',
    },
    questionCount: 7,
    estimatedMinutes: 10,
    createdAt: '2024-09-10T07:00:00Z',
    updatedAt: '2024-09-10T07:00:00Z',
    authorId: 'u1',
    questions: [],
  },
  {
    id: 'q4',
    title: 'Общая эрудиция: демо-квиз',
    description: 'Тестовый квиз для проверки работы платформы',
    status: 'published',
    mode: 'teacher_paced',
    settings: {
      shuffleQuestions: false,
      shuffleAnswers: false,
      showLeaderboard: true,
      themeColor: 'purple',
    },
    questionCount: 4,
    estimatedMinutes: 5,
    createdAt: '2024-10-01T10:00:00Z',
    updatedAt: '2024-10-01T10:00:00Z',
    authorId: 'u1',
    questions: [
      {
        id: 'dq1',
        quizId: 'q4',
        type: 'multiple_choice',
        text: 'В каком году человек впервые ступил на поверхность Луны?',
        timeLimitSeconds: 20,
        order: 1,
        options: [
          { id: 'da1', text: '1963', isCorrect: false, color: 'red' },
          { id: 'da2', text: '1974', isCorrect: false, color: 'blue' },
          { id: 'da3', text: '1967', isCorrect: false, color: 'yellow' },
          { id: 'da4', text: '1969', isCorrect: true, color: 'green' },
        ],
      },
      {
        id: 'dq2',
        quizId: 'q4',
        type: 'true_false',
        text: 'Земля является третьей планетой от Солнца',
        timeLimitSeconds: 15,
        order: 2,
        options: [
          { id: 'db1', text: 'Верно', isCorrect: true, color: 'green' },
          { id: 'db2', text: 'Неверно', isCorrect: false, color: 'red' },
        ],
      },
      {
        id: 'dq3',
        quizId: 'q4',
        type: 'multiple_select',
        text: 'Какие из перечисленных языков программирования являются компилируемыми?',
        timeLimitSeconds: 30,
        order: 3,
        options: [
          { id: 'dc1', text: 'Python', isCorrect: false, color: 'red' },
          { id: 'dc2', text: 'C++', isCorrect: true, color: 'blue' },
          { id: 'dc3', text: 'JavaScript', isCorrect: false, color: 'yellow' },
          { id: 'dc4', text: 'Go', isCorrect: true, color: 'green' },
        ],
      },
      {
        id: 'dq4',
        quizId: 'q4',
        type: 'open_answer',
        text: 'Как называется столица Франции?',
        timeLimitSeconds: 15,
        order: 4,
        options: [
          { id: 'dd1', text: 'Париж', isCorrect: true, color: 'green' },
          { id: 'dd2', text: 'Paris', isCorrect: true, color: 'green' },
          { id: 'dd3', text: 'париж', isCorrect: true, color: 'green' },
        ],
      },
    ],
  },
];

const MOCK_REPORTS: GameReport[] = [
  {
    id: 'r1',
    sessionId: 's1',
    quizId: 'q1',
    quizTitle: 'Интегралы 2',
    playedAt: '2024-10-14T10:00:00Z',
    participantCount: 42,
    avgScore: 7.8,
    questionReports: [
      {
        questionId: 'qn1',
        questionText: 'В каком году человек впервые ступил на поверхность Луны?',
        correctPercent: 78,
        avgResponseTimeMs: 2300,
        mostCommonWrongOptionId: 'a2',
        mostCommonWrongOptionText: 'Вариант Б',
        distribution: [
          { optionId: 'a1', optionText: '1963', count: 28, isCorrect: false, color: 'red' },
          { optionId: 'a2', optionText: '1974', count: 6, isCorrect: false, color: 'blue' },
          { optionId: 'a3', optionText: '1967', count: 3, isCorrect: false, color: 'yellow' },
          { optionId: 'a4', optionText: '1969', count: 5, isCorrect: true, color: 'green' },
        ],
        fastestCorrectParticipants: [
          { id: 'p1', nickname: 'Иван Р' },
          { id: 'p2', nickname: 'София К' },
          { id: 'p3', nickname: 'Мария Л' },
          { id: 'p4', nickname: 'Даниил С' },
          { id: 'p5', nickname: 'Виктор М' },
        ],
      },
    ],
    leaderboard: [
      { id: 'p1', nickname: 'Иван Р', avatarInitials: 'ИР', avatarColor: '#7c3aed', score: 9200, answeredCount: 15, correctAnswers: 15, totalQuestions: 15, rank: 1 },
      { id: 'p2', nickname: 'София К', avatarInitials: 'СК', avatarColor: '#2563eb', score: 8900, answeredCount: 15, correctAnswers: 15, totalQuestions: 15, rank: 2 },
      { id: 'p3', nickname: 'Мария Л', avatarInitials: 'МЛ', avatarColor: '#16a34a', score: 8400, answeredCount: 14, correctAnswers: 14, totalQuestions: 15, rank: 3 },
    ],
  },
];

const LS_SESSIONS_KEY = 'sb_mock_sessions';

function getSessions(): GameSession[] {
  try {
    const raw = localStorage.getItem(LS_SESSIONS_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveSessions(sessions: GameSession[]): void {
  try {
    localStorage.setItem(LS_SESSIONS_KEY, JSON.stringify(sessions));
  } catch {
    
  }
}

function findSession(predicate: (s: GameSession) => boolean): GameSession | undefined {
  return getSessions().find(predicate);
}

function updateSession(id: string, updater: (s: GameSession) => GameSession): void {
  const sessions = getSessions();
  const idx = sessions.findIndex((s) => s.id === id);
  if (idx !== -1) {
    sessions[idx] = updater(sessions[idx]);
    saveSessions(sessions);
  }
}

const authApi: IAuthApi = {
  async getMe() {
    await delay();
    const token = localStorage.getItem('accessToken');
    if (!token) throw new Error('Unauthorized');
    return MOCK_USER;
  },
  async login({ email, password }: LoginRequest): Promise<AuthResponse> {
    await delay(600);
    if (email === 'teacher@example.com' && password === 'password') {
      return { user: MOCK_USER, tokens: { accessToken: 'mock-token-123' } };
    }
    throw Object.assign(new Error('Неверный email или пароль'), { statusCode: 401 });
  },
  async register(data: RegisterRequest): Promise<AuthResponse> {
    await delay(700);
    const user: User = { ...MOCK_USER, id: uid(), email: data.email, name: data.name };
    return { user, tokens: { accessToken: 'mock-token-' + uid() } };
  },
  async changePassword(_data: ChangePasswordRequest): Promise<AuthResponse & { message?: string }> {
    await delay();
    // В моке всё равно отдаём свежий "токен", чтобы FE-флоу не ветвился по режиму.
    return {
      user: { ...MOCK_USER },
      tokens: { accessToken: 'mock-token-' + uid() },
      message: 'password changed successfully',
    };
  },
  async forgotPassword(_data: ForgotPasswordRequest) {
    await delay();
  },
  async resetPassword(_data: ResetPasswordRequest) {
    await delay();
  },
};

const quizzesApi: IQuizApi = {
  async getMyQuizzes() {
    await delay();
    
    try { localStorage.setItem('sb_mock_quizzes', JSON.stringify(MOCK_QUIZZES)); } catch {  }
    return [...MOCK_QUIZZES];
  },
  async getPublicQuizzes() {
    await delay();
    return MOCK_QUIZZES.filter((q) => q.status === 'published');
  },
  async getQuizById(id: string) {
    await delay();
    const quiz = MOCK_QUIZZES.find((q) => q.id === id);
    if (!quiz) throw new Error('Quiz not found');
    return quiz;
  },
  async createQuiz(data: CreateQuizRequest): Promise<Quiz> {
    await delay(500);
    const quiz: Quiz = {
      ...data,
      id: uid(),
      status: data.status ?? ('draft' as const),
      questionCount: data.questions.length,
      estimatedMinutes: Math.ceil(
        data.questions.reduce((s, q) => s + q.timeLimitSeconds, 0) / 60
      ),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      authorId: 'u1',
      questions: data.questions.map((q, i) => ({
        ...q,
        id: uid(),
        quizId: '',
        order: i + 1,
      })),
    };
    MOCK_QUIZZES.push(quiz);
    try { localStorage.setItem('sb_mock_quizzes', JSON.stringify(MOCK_QUIZZES)); } catch {  }
    return quiz;
  },
  async updateQuiz(id: string, data: UpdateQuizRequest): Promise<Quiz> {
    await delay(500);
    const idx = MOCK_QUIZZES.findIndex((q) => q.id === id);
    if (idx === -1) throw new Error('Quiz not found');
    const { questions: newQuestions, ...rest } = data;
    MOCK_QUIZZES[idx] = {
      ...MOCK_QUIZZES[idx],
      ...rest,
      questions: newQuestions
        ? newQuestions.map((q, i) => ({ ...q, id: uid(), quizId: id, order: i + 1 }))
        : MOCK_QUIZZES[idx].questions,
      updatedAt: new Date().toISOString(),
    };
    return MOCK_QUIZZES[idx];
  },
  async deleteQuiz(id: string) {
    await delay();
    const idx = MOCK_QUIZZES.findIndex((q) => q.id === id);
    if (idx !== -1) MOCK_QUIZZES.splice(idx, 1);
  },
};

const sessionsApi: ISessionApi = {
  async createSession(quizId: string, mode?: string): Promise<GameSession> {
    await delay(500);
    const pin = String(Math.floor(100000 + Math.random() * 900000));
    const quiz = MOCK_QUIZZES.find((q) => q.id === quizId);

    
    const existingSessions = getSessions().filter((s) =>
      s.status !== 'finished' && s.quizId !== quizId
    );
    saveSessions(existingSessions);

    const session: GameSession = {
      id: uid(),
      quizId,
      pin,
      status: 'waiting',
      mode: (mode as 'teacher_paced' | 'student_paced' | undefined) ?? 'teacher_paced',
      currentQuestionIndex: 0,
      totalQuestions: quiz?.questionCount ?? 0,
      participants: [],
    };
    const sessions = getSessions();
    sessions.push(session);
    saveSessions(sessions);
    return session;
  },
  async joinSession({ pin, nickname }: JoinSessionRequest): Promise<JoinSessionResponse> {
    await delay(400);
    const session = findSession((s) => s.pin === pin);
    if (!session) throw new Error('Сессия с таким PIN не найдена');
    const quiz = MOCK_QUIZZES.find((q) => q.id === session.quizId);
    if (!quiz) throw new Error('Квиз не найден');
    const participantId = uid();
    
    const color = ['#7c3aed','#2563eb','#16a34a','#dc2626','#ea580c','#0891b2','#be185d'][
      Math.floor(Math.random() * 7)
    ];
    updateSession(session.id, (s) => ({
      ...s,
      participants: [
        ...s.participants,
        {
          id: participantId,
          nickname,
          avatarInitials: nickname.slice(0, 2).toUpperCase(),
          avatarColor: color,
          score: 0,
          answeredCount: 0,
        },
      ],
    }));
    return {
      sessionId: session.id,
      participantId,
      quiz: { id: quiz.id, title: quiz.title, mode: quiz.mode },
    };
  },
  async getSession(sessionId: string): Promise<GameSession> {
    await delay(300);
    const session = findSession((s) => s.id === sessionId);
    if (!session) throw new Error('Session not found');
    return session;
  },
  async startSession(sessionId: string) {
    await delay(300);
    updateSession(sessionId, (s) => ({ ...s, status: 'question_active' }));
  },
  async endSession(sessionId: string) {
    await delay(300);
    updateSession(sessionId, (s) => ({ ...s, status: 'finished' }));
    writeSignal({ event: 'session_finished', payload: {}, sessionId, ts: Date.now() });
  },
};

const analyticsApi: IAnalyticsApi = {
  async getReports() {
    await delay();
    return [...MOCK_REPORTS];
  },
  async getReportById(id: string) {
    await delay();
    const report = MOCK_REPORTS.find((r) => r.id === id);
    if (!report) throw new Error('Report not found');
    return report;
  },
  async exportReportCsv(id: string): Promise<Blob> {
    await delay(800);
    const report = MOCK_REPORTS.find((r) => r.id === id);
    if (!report) throw new Error('Report not found');

    
    const rows: string[] = [
      `Квиз:,${report.quizTitle}`,
      `Дата:,${new Date(report.playedAt).toLocaleString('ru-RU')}`,
      `Участников:,${report.participantCount}`,
      `Средний балл:,${report.avgScore}`,
      '',
      
      'Вопрос,% верных,Среднее время (сек),Частая ошибка',
      ...report.questionReports.map((qr) =>
        [
          `"${qr.questionText}"`,
          `${qr.correctPercent}%`,
          (qr.avgResponseTimeMs / 1000).toFixed(1),
          qr.mostCommonWrongOptionText ?? '—',
        ].join(',')
      ),
      '',
      
      'Место,Участник,Счёт,Ответов',
      ...report.leaderboard.map((p) =>
        [p.rank, `"${p.nickname}"`, p.score, p.answeredCount].join(',')
      ),
    ];

    const csv = rows.join('\n');
    return new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
  },
};

export const mockApiService: IApiService = {
  auth: authApi,
  quizzes: quizzesApi,
  sessions: sessionsApi,
  analytics: analyticsApi,
};
