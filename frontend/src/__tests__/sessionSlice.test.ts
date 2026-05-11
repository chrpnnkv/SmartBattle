import { describe, it, expect } from 'vitest';
import reducer, {
  resetSession,
  participantJoined,
  participantLeft,
  questionStarted,
  questionEnded,
  tickTimer,
  leaderboardUpdated,
  sessionFinished,
  clearSession,
} from '../store/slices/sessionSlice';
import type { QuestionReport } from '../types';
import type { GameSession, SessionParticipant, Question } from '../types';

const mockParticipant: SessionParticipant = {
  id: 'p1',
  nickname: 'Alice',
  avatarInitials: 'A',
  avatarColor: '#7c3aed',
  score: 0,
  answeredCount: 0,
};

const mockSession: GameSession = {
  id: 'sess1',
  quizId: 'quiz1',
  pin: '123456',
  status: 'waiting',
  mode: 'teacher_paced',
  currentQuestionIndex: 0,
  totalQuestions: 3,
  participants: [],
};

const mockQuestion: Question = {
  id: 'q1',
  quizId: 'quiz1',
  type: 'multiple_choice',
  text: 'Что такое TypeScript?',
  timeLimitSeconds: 30,
  order: 0,
  options: [
    { id: 'a1', text: 'Язык', isCorrect: true, color: 'red' },
    { id: 'a2', text: 'Фреймворк', isCorrect: false, color: 'blue' },
  ],
};

const initialState = {
  session: null,
  isLoading: false,
  error: null,
  joinData: null,
  currentQuestion: null,
  selectedAnswerId: null,
  answerSubmitted: false,
  timeLeftSeconds: 0,
  questionReport: null,
};

describe('sessionSlice', () => {
  describe('resetSession', () => {
    it('clears all state fields', () => {
      const dirty = {
        ...initialState,
        session: mockSession,
        error: 'some error',
        currentQuestion: mockQuestion,
        timeLeftSeconds: 15,
        answerSubmitted: true,
        selectedAnswerId: 'a1',
      };
      expect(reducer(dirty, resetSession())).toEqual(initialState);
    });
  });

  describe('participantJoined', () => {
    it('adds new participant to session', () => {
      const state = { ...initialState, session: { ...mockSession, participants: [] } };
      const result = reducer(state, participantJoined(mockParticipant));
      expect(result.session!.participants).toHaveLength(1);
      expect(result.session!.participants[0]).toEqual(mockParticipant);
    });

    it('does not add duplicate participant', () => {
      const state = { ...initialState, session: { ...mockSession, participants: [mockParticipant] } };
      const result = reducer(state, participantJoined(mockParticipant));
      expect(result.session!.participants).toHaveLength(1);
    });

    it('does nothing when session is null', () => {
      const result = reducer(initialState, participantJoined(mockParticipant));
      expect(result.session).toBeNull();
    });
  });

  describe('participantLeft', () => {
    it('removes participant by id', () => {
      const state = { ...initialState, session: { ...mockSession, participants: [mockParticipant] } };
      const result = reducer(state, participantLeft('p1'));
      expect(result.session!.participants).toHaveLength(0);
    });

    it('ignores unknown id', () => {
      const state = { ...initialState, session: { ...mockSession, participants: [mockParticipant] } };
      const result = reducer(state, participantLeft('unknown'));
      expect(result.session!.participants).toHaveLength(1);
    });
  });

  describe('questionStarted', () => {
    it('sets question and resets answer state', () => {
      const state = {
        ...initialState,
        session: mockSession,
        selectedAnswerId: 'a1',
        answerSubmitted: true,
        timeLeftSeconds: 10,
      };
      const result = reducer(state, questionStarted({ question: mockQuestion, timeLimit: 30 }));
      expect(result.currentQuestion).toEqual(mockQuestion);
      expect(result.selectedAnswerId).toBeNull();
      expect(result.answerSubmitted).toBe(false);
      expect(result.timeLeftSeconds).toBe(30);
      expect(result.questionReport).toBeNull();
    });

    it('updates session status to question_active', () => {
      const state = { ...initialState, session: mockSession };
      const result = reducer(state, questionStarted({ question: mockQuestion, timeLimit: 20 }));
      expect(result.session!.status).toBe('question_active');
    });
  });

  describe('tickTimer', () => {
    it('decrements timeLeftSeconds by 1', () => {
      const result = reducer({ ...initialState, timeLeftSeconds: 5 }, tickTimer());
      expect(result.timeLeftSeconds).toBe(4);
    });

    it('does not go below 0', () => {
      const result = reducer({ ...initialState, timeLeftSeconds: 0 }, tickTimer());
      expect(result.timeLeftSeconds).toBe(0);
    });
  });

  describe('sessionFinished', () => {
    it('sets session status to finished', () => {
      const state = { ...initialState, session: { ...mockSession, status: 'question_active' as const } };
      const result = reducer(state, sessionFinished());
      expect(result.session!.status).toBe('finished');
    });

    it('does nothing when session is null', () => {
      const result = reducer(initialState, sessionFinished());
      expect(result.session).toBeNull();
    });
  });

  describe('questionEnded', () => {
    const mockReport: QuestionReport = {
      questionId: 'q1',
      questionText: 'Что такое TypeScript?',
      correctPercent: 75,
      avgResponseTimeMs: 2300,
      distribution: [],
      fastestCorrectParticipants: [],
    };

    it('saves report to state', () => {
      const result = reducer(initialState, questionEnded(mockReport));
      expect(result.questionReport).toEqual(mockReport);
    });

    it('sets session status to question_results', () => {
      const state = { ...initialState, session: { ...mockSession, status: 'question_active' as const } };
      const result = reducer(state, questionEnded(mockReport));
      expect(result.session!.status).toBe('question_results');
    });

    it('does nothing to session when session is null', () => {
      const result = reducer(initialState, questionEnded(mockReport));
      expect(result.session).toBeNull();
    });
  });

  describe('leaderboardUpdated', () => {
    it('replaces session participants', () => {
      const updated = [{ ...mockParticipant, score: 1200 }];
      const state = { ...initialState, session: { ...mockSession, participants: [mockParticipant] } };
      const result = reducer(state, leaderboardUpdated(updated));
      expect(result.session!.participants).toEqual(updated);
    });

    it('does nothing when session is null', () => {
      const result = reducer(initialState, leaderboardUpdated([mockParticipant]));
      expect(result.session).toBeNull();
    });
  });

  describe('clearSession', () => {
    it('returns initial state', () => {
      const dirty = { ...initialState, session: mockSession, error: 'err', timeLeftSeconds: 10 };
      expect(reducer(dirty, clearSession())).toEqual(initialState);
    });
  });
});