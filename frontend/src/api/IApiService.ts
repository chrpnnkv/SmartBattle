import type {
  AuthResponse,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  GameReport,
  LoginRequest,
  Quiz,
  CreateQuizRequest,
  UpdateQuizRequest,
  RegisterRequest,
  ResetPasswordRequest,
  User,
  JoinSessionRequest,
  JoinSessionResponse,
  GameSession,
} from '../types';

export interface IAuthApi {
  getMe(): Promise<User>;
  login(data: LoginRequest): Promise<AuthResponse>;
  register(data: RegisterRequest): Promise<AuthResponse>;
  // ChangePassword теперь возвращает свежие токен и user — токен ротируется при смене пароля,
  // и фронт обязан сохранить новый, иначе старый JWT будет тикать к истечению "втихую".
  changePassword(data: ChangePasswordRequest): Promise<AuthResponse & { message?: string }>;
  forgotPassword(data: ForgotPasswordRequest): Promise<void>;
  resetPassword(data: ResetPasswordRequest): Promise<void>;
}

export interface IQuizApi {
  getMyQuizzes(): Promise<Quiz[]>;
  getPublicQuizzes(): Promise<Quiz[]>;
  getQuizById(id: string): Promise<Quiz>;
  createQuiz(data: CreateQuizRequest): Promise<Quiz>;
  updateQuiz(id: string, data: UpdateQuizRequest): Promise<Quiz>;
  deleteQuiz(id: string): Promise<void>;
  uploadImage(file: File): Promise<string>;
}

export interface ISessionApi {
  createSession(quizId: string, mode?: string): Promise<GameSession>;
  joinSession(data: JoinSessionRequest): Promise<JoinSessionResponse>;
  getSession(sessionId: string): Promise<GameSession>;
  // start/end зеркалят статус в БД Core. Прогресс по вопросам и ответам идёт через WebSocket.
  startSession(sessionId: string): Promise<void>;
  endSession(sessionId: string): Promise<void>;
}

export interface IAnalyticsApi {
  getReports(): Promise<GameReport[]>;
  getReportById(id: string): Promise<GameReport>;
  exportReportCsv(id: string): Promise<Blob>;
}

export type WsEventHandler<T = unknown> = (payload: T) => void;

export interface WsConnectOptions {
  roomCode?: string;
  name?: string;
  token?: string;
  participantId?: string;
}

export interface IWebSocketService {
  connect(sessionId: string, options?: WsConnectOptions): void;
  disconnect(): void;
  on<T = unknown>(event: string, handler: WsEventHandler<T>): void;
  off(event: string): void;
  send(event: string, payload?: unknown): void;
  isConnected(): boolean;
}

export interface IApiService {
  auth: IAuthApi;
  quizzes: IQuizApi;
  sessions: ISessionApi;
  analytics: IAnalyticsApi;
}
