import type {
  IApiService,
  IAuthApi,
  IQuizApi,
  ISessionApi,
  IAnalyticsApi,
} from '../IApiService';
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
  SubmitAnswerRequest,
  User,
} from '../../types';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

async function http<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('accessToken');

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw Object.assign(new Error(err.message ?? 'Request failed'), {
      statusCode: res.status,
      errors: err.errors,
    });
  }

  
  if (res.status === 204) return undefined as T;

  return res.json();
}

const get = <T>(path: string) => http<T>(path, { method: 'GET' });
const post = <T>(path: string, body?: unknown) =>
  http<T>(path, { method: 'POST', body: JSON.stringify(body) });
const put = <T>(path: string, body?: unknown) =>
  http<T>(path, { method: 'PUT', body: JSON.stringify(body) });
const del = <T>(path: string) => http<T>(path, { method: 'DELETE' });

const authApi: IAuthApi = {
  getMe: () => get<User>('/api/me'),
  login: (data: LoginRequest) => post<AuthResponse>('/auth/login', data),
  register: (data: RegisterRequest) => post<AuthResponse>('/auth/register', data),
  changePassword: (data: ChangePasswordRequest) =>
    post<void>('/auth/change-password', data),
  forgotPassword: (data: ForgotPasswordRequest) =>
    post<void>('/auth/forgot-password', data),
  resetPassword: (data: ResetPasswordRequest) =>
    post<void>('/auth/reset-password', data),
};

const quizzesApi: IQuizApi = {
  getMyQuizzes: () => get<Quiz[]>('/api/quizzes'),
  getPublicQuizzes: () => get<Quiz[]>('/api/quizzes/public'),
  getQuizById: (id: string) => get<Quiz>(`/api/quizzes/${id}`),
  createQuiz: (data: CreateQuizRequest) => post<Quiz>('/api/quizzes', data),
  updateQuiz: (id: string, data: UpdateQuizRequest) =>
    put<Quiz>(`/api/quizzes/${id}`, data),
  deleteQuiz: (id: string) => del<void>(`/api/quizzes/${id}`),
};

const sessionsApi: ISessionApi = {
  createSession: (quizId: string) =>
    post<GameSession>('/api/sessions', { quizId }),
  joinSession: (data: JoinSessionRequest) =>
    post<JoinSessionResponse>('/api/sessions/join', data),
  getSession: (sessionId: string) =>
    get<GameSession>(`/api/sessions/${sessionId}`),
  startSession: (sessionId: string) =>
    post<void>(`/api/sessions/${sessionId}/start`),
  nextQuestion: (sessionId: string) =>
    post<void>(`/api/sessions/${sessionId}/next`),
  endSession: (sessionId: string) =>
    post<void>(`/api/sessions/${sessionId}/end`),
  submitAnswer: (data: SubmitAnswerRequest) =>
    post<void>('/api/sessions/answer', data),
};

const analyticsApi: IAnalyticsApi = {
  getReports: () => get<GameReport[]>('/api/reports'),
  getReportById: (id: string) => get<GameReport>(`/api/reports/${id}`),
  exportReportCsv: async (id: string): Promise<Blob> => {
    const token = localStorage.getItem('accessToken');
    const res = await fetch(`${BASE_URL}/api/reports/${id}/export`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) throw new Error('Export failed');
    return res.blob();
  },
};

export const realApiService: IApiService = {
  auth: authApi,
  quizzes: quizzesApi,
  sessions: sessionsApi,
  analytics: analyticsApi,
};
