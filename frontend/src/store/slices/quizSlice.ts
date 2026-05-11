import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { api } from '../../api';
import type { Quiz, CreateQuizRequest, UpdateQuizRequest } from '../../types';

interface QuizState {
  quizzes: Quiz[];
  currentQuiz: Quiz | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: QuizState = {
  quizzes: [],
  currentQuiz: null,
  isLoading: false,
  error: null,
};

export const fetchMyQuizzes = createAsyncThunk(
  'quiz/fetchMy',
  async (_arg, { rejectWithValue }) => {
    try {
      return await api.quizzes.getMyQuizzes();
    } catch (e: unknown) {
      return rejectWithValue((e as Error)?.message ?? 'Не удалось загрузить квизы');
    }
  }
);

export const fetchQuizById = createAsyncThunk(
  'quiz/fetchById',
  async (id: string) => api.quizzes.getQuizById(id)
);

export const createQuiz = createAsyncThunk(
  'quiz/create',
  async (data: CreateQuizRequest, { rejectWithValue }) => {
    try {
      return await api.quizzes.createQuiz(data);
    } catch (e: unknown) {
      return rejectWithValue((e as Error).message);
    }
  }
);

export const updateQuiz = createAsyncThunk(
  'quiz/update',
  async ({ id, data }: { id: string; data: UpdateQuizRequest }, { rejectWithValue }) => {
    try {
      return await api.quizzes.updateQuiz(id, data);
    } catch (e: unknown) {
      return rejectWithValue((e as Error).message);
    }
  }
);

export const deleteQuiz = createAsyncThunk(
  'quiz/delete',
  async (id: string) => {
    await api.quizzes.deleteQuiz(id);
    return id;
  }
);

const quizSlice = createSlice({
  name: 'quiz',
  initialState,
  reducers: {
    setCurrentQuiz(state, action) {
      state.currentQuiz = action.payload;
    },
    clearCurrentQuiz(state) {
      state.currentQuiz = null;
    },
  },
  extraReducers: (builder) => {
    const setLoading = (state: QuizState) => { state.isLoading = true; state.error = null; };
    const setError = (state: QuizState, action: { payload?: unknown }) => {
      state.isLoading = false;
      state.error = (action.payload as string) ?? 'Ошибка';
    };

    builder
      .addCase(fetchMyQuizzes.pending, setLoading)
      .addCase(fetchMyQuizzes.fulfilled, (state, action) => {
        state.isLoading = false;
        state.quizzes = action.payload;
      })
      .addCase(fetchMyQuizzes.rejected, setError)

      .addCase(fetchQuizById.pending, setLoading)
      .addCase(fetchQuizById.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentQuiz = action.payload;
      })
      .addCase(fetchQuizById.rejected, setError)

      .addCase(createQuiz.fulfilled, (state, action) => {
        state.quizzes.unshift(action.payload);
        state.currentQuiz = action.payload;
      })
      .addCase(createQuiz.rejected, setError)

      .addCase(updateQuiz.fulfilled, (state, action) => {
        const idx = state.quizzes.findIndex((q) => q.id === action.payload.id);
        if (idx !== -1) state.quizzes[idx] = action.payload;
        state.currentQuiz = action.payload;
      })

      .addCase(deleteQuiz.fulfilled, (state, action) => {
        state.quizzes = state.quizzes.filter((q) => q.id !== action.payload);
      });
  },
});

export const { setCurrentQuiz, clearCurrentQuiz } = quizSlice.actions;
export default quizSlice.reducer;
