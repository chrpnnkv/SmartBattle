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
  User,
} from '../../types';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

const NO_REDIRECT_ON_401 = [
  '/auth/login',
  '/auth/register',
  '/auth/forgot-password',
  '/auth/reset-password',
  '/api/sessions/join',
  '/api/public/quizzes',
];

let redirecting = false;

function handleUnauthorized(path: string) {
if (redirecting) return;
  if (NO_REDIRECT_ON_401.some((p) => path.startsWith(p))) return;
  redirecting = true;
  try {
    localStorage.removeItem('accessToken');
  } catch {
  }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('sb:unauthorized'));
    if (window.location.pathname !== '/login') {
        const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password'];
        const currentPath = window.location.pathname;

        if (!publicPaths.some(path => currentPath.startsWith(path))) {
            window.location.href = '/login?reason=expired';
        }
    }
  }
}


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
    if (res.status === 401) {
      handleUnauthorized(path);
    }
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw Object.assign(new Error(err.message ?? err.error ?? 'Request failed'), {
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
    post<AuthResponse & { message?: string }>('/auth/change-password', data),
  forgotPassword: (data: ForgotPasswordRequest) =>
    post<void>('/auth/forgot-password', data),
  resetPassword: (data: ResetPasswordRequest) =>
    post<void>('/auth/reset-password', data),
};

const quizzesApi: IQuizApi = {
  getMyQuizzes: () => get<Quiz[]>('/api/quizzes'),
  getPublicQuizzes: () => get<Quiz[]>('/api/public/quizzes'),
  getQuizById: (id: string) => get<Quiz>(`/api/quizzes/${id}`),
  createQuiz: (data: CreateQuizRequest) => post<Quiz>('/api/quizzes', data),
  updateQuiz: (id: string, data: UpdateQuizRequest) =>
    put<Quiz>(`/api/quizzes/${id}`, data),
  deleteQuiz: (id: string) => del<void>(`/api/quizzes/${id}`),
};

const sessionsApi: ISessionApi = {
  createSession: (quizId: string, mode?: string) =>
    post<GameSession>('/api/sessions', mode ? { quizId, mode } : { quizId }),
  joinSession: (data: JoinSessionRequest) =>
    post<JoinSessionResponse>('/api/sessions/join', data),
  getSession: (sessionId: string) =>
    get<GameSession>(`/api/sessions/${sessionId}`),
  startSession: (sessionId: string) =>
    post<void>(`/api/sessions/${sessionId}/start`),
  endSession: (sessionId: string) =>
    post<void>(`/api/sessions/${sessionId}/end`),
};

const analyticsApi: IAnalyticsApi = {
  getReports: () => get<GameReport[]>('/api/reports'),
  getReportById: (id: string) => get<GameReport>(`/api/reports/${id}`),
  exportReportCsv: async (id: string): Promise<Blob> => {
    const token = localStorage.getItem('accessToken');
    const path = `/api/reports/${id}/export`;
    const res = await fetch(`${BASE_URL}${path}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) {
      if (res.status === 401) handleUnauthorized(path);
      throw new Error('Export failed');
    }
    return res.blob();
  },
};

export const realApiService: IApiService = {
  auth: authApi,
  quizzes: quizzesApi,
  sessions: sessionsApi,
  analytics: analyticsApi,
};
