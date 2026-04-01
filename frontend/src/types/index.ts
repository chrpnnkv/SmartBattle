

export interface User {
  id: string;
  email: string;
  name: string;
  role: 'teacher' | 'student';
  createdAt: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

export interface AuthResponse {
  user: User;
  tokens: AuthTokens;
}

export type QuestionType = 'multiple_choice' | 'multiple_select' | 'true_false' | 'open_answer';

export type QuizStatus = 'draft' | 'published';

export type QuizMode = 'teacher_paced' | 'student_paced';

export interface AnswerOption {
  id: string;
  text: string;
  isCorrect: boolean;
  color: 'red' | 'blue' | 'yellow' | 'green';
}

export interface Question {
  id: string;
  quizId: string;
  type: QuestionType;
  text: string;
  imageUrl?: string;
  timeLimitSeconds: number;
  order: number;
  options: AnswerOption[];
}

export interface QuizSettings {
  shuffleQuestions: boolean;
  shuffleAnswers: boolean;
  showLeaderboard: boolean;
  themeColor: 'purple' | 'red' | 'orange' | 'blue';
}

export interface Quiz {
  id: string;
  title: string;
  description?: string;
  status: QuizStatus;
  mode: QuizMode;
  settings: QuizSettings;
  questions: Question[];
  questionCount: number;
  estimatedMinutes: number;
  createdAt: string;
  updatedAt: string;
  authorId: string;
}

export interface CreateQuizRequest {
  title: string;
  description?: string;
  mode: QuizMode;
  settings: QuizSettings;
  questions: Omit<Question, 'id' | 'quizId'>[];
}

export interface UpdateQuizRequest extends Partial<CreateQuizRequest> {}

export type SessionStatus =
  | 'waiting'
  | 'question_active'
  | 'question_results'
  | 'finished';

export interface SessionParticipant {
  id: string;
  nickname: string;
  avatarInitials: string;
  avatarColor: string;
  score: number;
  answeredCount: number;
}

export interface GameSession {
  id: string;
  quizId: string;
  pin: string;
  status: SessionStatus;
  mode: QuizMode;
  currentQuestionIndex: number;
  totalQuestions: number;
  participants: SessionParticipant[];
  startedAt?: string;
  finishedAt?: string;
}

export interface JoinSessionRequest {
  pin: string;
  nickname: string;
}

export interface JoinSessionResponse {
  sessionId: string;
  participantId: string;
  quiz: Pick<Quiz, 'id' | 'title' | 'mode'>;
}

export interface SubmitAnswerRequest {
  sessionId: string;
  participantId: string;
  questionId: string;
  answerId: string;
  timeSpentMs: number;
}

export interface AnswerDistribution {
  optionId: string;
  optionText: string;
  count: number;
  isCorrect: boolean;
  color: AnswerOption['color'];
}

export interface QuestionReport {
  questionId: string;
  questionText: string;
  correctPercent: number;
  avgResponseTimeMs: number;
  mostCommonWrongOptionId?: string;
  mostCommonWrongOptionText?: string;
  distribution: AnswerDistribution[];
  fastestCorrectParticipants: Pick<SessionParticipant, 'id' | 'nickname'>[];
}

export interface GameReport {
  id: string;
  sessionId: string;
  quizId: string;
  quizTitle: string;
  playedAt: string;
  participantCount: number;
  avgScore: number;
  questionReports: QuestionReport[];
  leaderboard: (SessionParticipant & { rank: number })[];
}

export type WsEventType =
  | 'participant_joined'
  | 'participant_left'
  | 'session_started'
  | 'question_started'
  | 'answer_received'
  | 'question_ended'
  | 'session_finished'
  | 'leaderboard_update';

export interface WsEvent<T = unknown> {
  type: WsEventType;
  payload: T;
}

export interface WsQuestionStartedPayload {
  question: Question;
  questionIndex: number;
  totalQuestions: number;
  startedAt: number; 
}

export interface WsQuestionEndedPayload {
  questionReport: QuestionReport;
  leaderboard: SessionParticipant[];
}

export interface WsParticipantJoinedPayload {
  participant: SessionParticipant;
  totalCount: number;
}

export interface ApiError {
  message: string;
  statusCode: number;
  errors?: Record<string, string[]>;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}
