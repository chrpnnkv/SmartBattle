import type { IWebSocketService, WsEventHandler } from '../api/IApiService';
import type {
  SessionParticipant,
  Question,
  WsQuestionStartedPayload,
  WsQuestionEndedPayload,
  WsParticipantJoinedPayload,
  GameSession,
} from '../types';

const LS_SESSIONS_KEY = 'sb_mock_sessions';
const wsSignalKey = (sessionId: string) => `sb_ws_signal_${sessionId}`;

interface WsSignal {
  event: string;
  payload: unknown;
  sessionId: string;
  ts: number;
}

function readSessions(): GameSession[] {
  try {
    const raw = localStorage.getItem(LS_SESSIONS_KEY);
    return raw ? (JSON.parse(raw) as GameSession[]) : [];
  } catch {
    return [];
  }
}

function readSignal(sessionId: string): WsSignal | null {
  try {
    const raw = localStorage.getItem(wsSignalKey(sessionId));
    return raw ? (JSON.parse(raw) as WsSignal) : null;
  } catch {
    return null;
  }
}

function writeSignal(signal: WsSignal): void {
  try {
    localStorage.setItem(wsSignalKey(signal.sessionId), JSON.stringify(signal));
  } catch {
    
  }
}

class MockWebSocketService implements IWebSocketService {
  private handlers: Map<string, WsEventHandler> = new Map();
  private connected = false;
  private timers: ReturnType<typeof setInterval>[] = [];
  private sentParticipants: Set<string> = new Set();
  private currentSessionId = '';
  private currentQuestion: Question | null = null;

  connect(sessionId: string, _participantId?: string): void {
    
    this.timers.forEach(clearInterval);
    this.timers = [];
    this.sentParticipants.clear();

    this.connected = true;
    this.currentSessionId = sessionId;

    
    const pollParticipants = () => {
      const session = readSessions().find((s) => s.id === sessionId);
      if (!session?.participants) return;
      session.participants.forEach((participant) => {
        if (this.sentParticipants.has(participant.id)) return;
        this.sentParticipants.add(participant.id);
        const payload: WsParticipantJoinedPayload = {
          participant,
          totalCount: this.sentParticipants.size,
        };
        this.emit('participant_joined', payload);
      });
    };

    pollParticipants();
    this.timers.push(setInterval(pollParticipants, 2000));

    
    let lastSignalTs = 0;
    const pollSignals = () => {
      const signal = readSignal(sessionId);
      if (!signal || signal.ts <= lastSignalTs) return;
      lastSignalTs = signal.ts;
      setTimeout(() => this.emit(signal.event, signal.payload), 50);
    };
    this.timers.push(setInterval(pollSignals, 500));
  }

  disconnect(): void {
    this.connected = false;
    this.timers.forEach(clearInterval);
    this.timers = [];
    this.handlers.clear();
    this.sentParticipants.clear();
    this.currentSessionId = '';
    this.currentQuestion = null;
    this.connected = false;
  }

  on<T = unknown>(event: string, handler: WsEventHandler<T>): void {
    this.handlers.set(event, handler as WsEventHandler);
  }

  off(event: string): void {
    this.handlers.delete(event);
  }

  send(event: string, payload: unknown): void {
    if (event === 'start_question') {
      const questionPayload = payload as WsQuestionStartedPayload;
      this.currentQuestion = questionPayload?.question ?? null;

      
      try {
        const now = Date.now();
        const questionState = {
          question: questionPayload.question,
          questionIndex: questionPayload.questionIndex ?? 0,
          totalQuestions: (questionPayload as WsQuestionStartedPayload & { totalQuestions?: number }).totalQuestions ?? 0,
          themeColor: 'purple',
          participantId: '',
          startedAt: now,
        };
        localStorage.setItem(`sb_current_question_${this.currentSessionId}`, JSON.stringify(questionState));
      } catch {  }

      
      const startedAt = Date.now();
      const signalPayload = { ...(payload as WsQuestionStartedPayload), startedAt };
      writeSignal({ event: 'question_started', payload: signalPayload, sessionId: this.currentSessionId, ts: startedAt });
      setTimeout(() => this.emit('question_started', signalPayload), 50);
    }

    if (event === 'end_question') {
      const endPayload = this.buildEndPayload();
      const endedAt = Date.now();
      const endSignalPayload = { ...endPayload, endedAt };
      writeSignal({ event: 'question_ended', payload: endSignalPayload, sessionId: this.currentSessionId, ts: endedAt });
      setTimeout(() => this.emit('question_ended', endSignalPayload), 50);
    }
  }

  isConnected(): boolean {
    return this.connected;
  }

  private buildEndPayload(): WsQuestionEndedPayload {
    const session = readSessions().find((s) => s.id === this.currentSessionId);
    const participants: SessionParticipant[] = session?.participants ?? [];
    const q = this.currentQuestion;
    const qType = q?.type ?? 'multiple_choice';

    
    const distribution = (() => {
      if (qType === 'open_answer') {
        
        const total = Math.max(participants.length, 1);
        const correctCount = Math.round(total * (0.4 + Math.random() * 0.4));
        return [
          { optionId: 'correct', optionText: q?.options[0]?.text ?? 'Верный ответ', count: correctCount, isCorrect: true, color: 'green' as const },
          { optionId: 'wrong', optionText: 'Другое', count: total - correctCount, isCorrect: false, color: 'red' as const },
        ];
      }
      if (q?.options && q.options.length > 0) {
        
        return q.options.map((opt) => ({
          optionId: opt.id,
          optionText: opt.text || opt.id,
          count: Math.round(Math.random() * (opt.isCorrect ? 15 : 10) + (opt.isCorrect ? 3 : 0)),
          isCorrect: opt.isCorrect,
          color: opt.color,
        }));
      }
      
      return [
        { optionId: 'a1', optionText: 'A', count: Math.round(Math.random() * 15 + 5), isCorrect: true,  color: 'red' as const },
        { optionId: 'a2', optionText: 'B', count: Math.round(Math.random() * 10),      isCorrect: false, color: 'blue' as const },
        { optionId: 'a3', optionText: 'C', count: Math.round(Math.random() * 8),       isCorrect: false, color: 'yellow' as const },
        { optionId: 'a4', optionText: 'D', count: Math.round(Math.random() * 12),      isCorrect: false, color: 'green' as const },
      ];
    })();

    const wrongOptions = distribution.filter((d) => !d.isCorrect).sort((a, b) => b.count - a.count);
    const mostCommonWrong = wrongOptions[0]?.optionText ?? '—';
    const correctCount = distribution.filter((d) => d.isCorrect).reduce((s, d) => s + d.count, 0);
    const total = distribution.reduce((s, d) => s + d.count, 0) || 1;

    return {
      questionReport: {
        questionId: q?.id ?? 'mock-q',
        questionText: q?.text ?? '',
        correctPercent: Math.round((correctCount / total) * 100),
        avgResponseTimeMs: Math.round(1500 + Math.random() * 2000),
        mostCommonWrongOptionText: mostCommonWrong,
        distribution,
        fastestCorrectParticipants: participants.slice(0, 5).map((p) => ({
          id: p.id,
          nickname: p.nickname,
        })),
      },
      leaderboard: participants.map((p, i) => ({
        ...p,
        score: Math.round((participants.length - i) * 800 + Math.random() * 400),
        answeredCount: 1,
      })),
    };
  }

  private emit(event: string, payload: unknown): void {
    this.handlers.get(event)?.(payload);
  }
}

class RealWebSocketService implements IWebSocketService {
  private socket: WebSocket | null = null;
  private handlers: Map<string, WsEventHandler> = new Map();
  private readonly WS_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8081';

  connect(sessionId: string, participantId?: string): void {
    const url = `${this.WS_URL}/ws/${sessionId}${participantId ? `?pid=${participantId}` : ''}`;
    this.socket = new WebSocket(url);

    this.socket.onmessage = (e: MessageEvent<string>) => {
      try {
        const { type, payload } = JSON.parse(e.data) as { type: string; payload: unknown };
        this.handlers.get(type)?.(payload);
      } catch {
        
      }
    };

    this.socket.onerror = () => {  };
    this.socket.onclose = () => {  };
  }

  disconnect(): void {
    this.socket?.close();
    this.socket = null;
    this.handlers.clear();
  }

  on<T = unknown>(event: string, handler: WsEventHandler<T>): void {
    this.handlers.set(event, handler as WsEventHandler);
  }

  off(event: string): void {
    this.handlers.delete(event);
  }

  send(event: string, payload: unknown): void {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type: event, payload }));
    }
  }

  isConnected(): boolean {
    return this.socket?.readyState === WebSocket.OPEN;
  }
}

const USE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false';

export const wsService: IWebSocketService = USE_MOCK
  ? new MockWebSocketService()
  : new RealWebSocketService();
