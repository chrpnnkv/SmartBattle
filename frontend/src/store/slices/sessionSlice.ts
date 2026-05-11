import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { api } from '../../api';
import type {
  GameSession,
  SessionParticipant,
  Question,
  QuestionReport,
  JoinSessionResponse,
} from '../../types';

interface SessionState {
  
  session: GameSession | null;
  isLoading: boolean;
  error: string | null;

  
  joinData: JoinSessionResponse | null;

  
  currentQuestion: Question | null;
  selectedAnswerId: string | null;
  answerSubmitted: boolean;
  timeLeftSeconds: number;

  
  questionReport: QuestionReport | null;
}

const initialState: SessionState = {
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

export const fetchSession = createAsyncThunk(
  'session/fetch',
  async (sessionId: string, { rejectWithValue }) => {
    try {
      return await api.sessions.getSession(sessionId);
    } catch (e: unknown) {
      return rejectWithValue((e as Error).message);
    }
  }
);

export const createSession = createAsyncThunk(
  'session/create',
  async ({ quizId, mode }: { quizId: string; mode?: string }, { rejectWithValue }) => {
    try {
      // mode передаётся в Core: ему важно знать, как идёт сессия
      return await api.sessions.createSession(quizId, mode);
    } catch (e: unknown) {
      return rejectWithValue((e as Error).message);
    }
  }
);

export const joinSession = createAsyncThunk(
  'session/join',
  async ({ pin, nickname }: { pin: string; nickname: string }, { rejectWithValue }) => {
    try {
      return await api.sessions.joinSession({ pin, nickname });
    } catch (e: unknown) {
      return rejectWithValue((e as Error).message);
    }
  }
);

export const startSession = createAsyncThunk(
  'session/start',
  async (sessionId: string) => api.sessions.startSession(sessionId)
);

export const endSession = createAsyncThunk(
  'session/end',
  async (sessionId: string) => api.sessions.endSession(sessionId)
);

const sessionSlice = createSlice({
  name: 'session',
  initialState,
  reducers: {
    resetSession(state) {
      state.session = null;
      state.joinData = null;
      state.currentQuestion = null;
      state.timeLeftSeconds = 0;
      state.answerSubmitted = false;
      state.selectedAnswerId = null;
      state.error = null;
    },
    
    participantJoined(state, action: PayloadAction<SessionParticipant>) {
      if (state.session) {
        const exists = state.session.participants.find((p) => p.id === action.payload.id);
        if (!exists) state.session.participants.push(action.payload);
      }
    },
    participantLeft(state, action: PayloadAction<string>) {
      if (state.session) {
        state.session.participants = state.session.participants.filter(
          (p) => p.id !== action.payload
        );
      }
    },
    questionStarted(state, action: PayloadAction<{ question: Question; timeLimit: number }>) {
      state.currentQuestion = action.payload.question;
      state.selectedAnswerId = null;
      state.answerSubmitted = false;
      state.timeLeftSeconds = action.payload.timeLimit;
      state.questionReport = null;
      if (state.session) state.session.status = 'question_active';
    },
    questionEnded(state, action: PayloadAction<QuestionReport>) {
      state.questionReport = action.payload;
      if (state.session) state.session.status = 'question_results';
    },
    tickTimer(state) {
      if (state.timeLeftSeconds > 0) state.timeLeftSeconds -= 1;
    },
    leaderboardUpdated(state, action: PayloadAction<SessionParticipant[]>) {
      if (state.session) state.session.participants = action.payload;
    },
    sessionFinished(state) {
      if (state.session) state.session.status = 'finished';
    },
    clearSession() {
      return initialState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchSession.fulfilled, (state, action) => {
        state.session = action.payload;
      })

      .addCase(createSession.pending, (state) => { state.isLoading = true; state.error = null; })
      .addCase(createSession.fulfilled, (state, action) => {
        state.isLoading = false;
        
        state.session = action.payload;
        state.currentQuestion = null;
        state.timeLeftSeconds = 0;
        state.answerSubmitted = false;
        state.selectedAnswerId = null;
        state.error = null;
      })
      .addCase(createSession.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      })

      .addCase(joinSession.pending, (state) => { state.isLoading = true; state.error = null; })
      .addCase(joinSession.fulfilled, (state, action) => {
        state.isLoading = false;
        state.joinData = action.payload;
      })
      .addCase(joinSession.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });
  },
});

export const {
  resetSession,
  participantJoined,
  participantLeft,
  questionStarted,
  questionEnded,
  tickTimer,
  leaderboardUpdated,
  sessionFinished,
  clearSession,
} = sessionSlice.actions;

export default sessionSlice.reducer;
