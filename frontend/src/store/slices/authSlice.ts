import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { api } from '../../api';
import type { User, LoginRequest, RegisterRequest, ChangePasswordRequest } from '../../types';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  isInitialized: boolean;
}

// Безопасное чтение токена при инициализации стора. localStorage может быть недоступен
// в test-окружении (jsdom при определённых конфигурациях), в Safari Private Mode или в SSR.
// Без этого guard'a сам импорт authSlice падает в любом окружении без window.localStorage.
function readInitialToken(): string | null {
  try {
    if (typeof localStorage === 'undefined') return null;
    return localStorage.getItem('accessToken');
  } catch {
    return null;
  }
}

const initialState: AuthState = {
  user: null,
  token: readInitialToken(),
  isLoading: false,
  error: null,
  isInitialized: false,
};

export const initAuth = createAsyncThunk('auth/init', async () => {
  const token = localStorage.getItem('accessToken');
  if (!token) return null;
  const user = await api.auth.getMe();
  return user;
});

export const login = createAsyncThunk(
  'auth/login',
  async (data: LoginRequest, { rejectWithValue }) => {
    try {
      const res = await api.auth.login(data);
      localStorage.setItem('accessToken', res.tokens.accessToken);
      return res;
    } catch (e: unknown) {
      const err = e as Error;
      return rejectWithValue(err.message);
    }
  }
);

export const register = createAsyncThunk(
  'auth/register',
  async (data: RegisterRequest, { rejectWithValue }) => {
    try {
      const res = await api.auth.register(data);
      localStorage.setItem('accessToken', res.tokens.accessToken);
      return res;
    } catch (e: unknown) {
      const err = e as Error;
      return rejectWithValue(err.message);
    }
  }
);

// При смене пароля бэкенд отдаёт свежий токен. Сохраняем его и в localStorage,
// и в Redux — иначе клиент продолжит ходить со старым JWT, который тикает к истечению.
export const changePassword = createAsyncThunk(
  'auth/changePassword',
  async (data: ChangePasswordRequest, { rejectWithValue }) => {
    try {
      const res = await api.auth.changePassword(data);
      if (res?.tokens?.accessToken) {
        localStorage.setItem('accessToken', res.tokens.accessToken);
      }
      return res;
    } catch (e: unknown) {
      return rejectWithValue((e as Error)?.message ?? 'Не удалось сменить пароль');
    }
  }
);

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    logout(state) {
      state.user = null;
      state.token = null;
      state.error = null;
      localStorage.removeItem('accessToken');
    },
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    
    builder
      .addCase(initAuth.fulfilled, (state, action) => {
        state.user = action.payload ?? null;
        state.isInitialized = true;
      })
      .addCase(initAuth.rejected, (state) => {
        state.user = null;
        state.token = null;
        state.isInitialized = true;
        localStorage.removeItem('accessToken');
      });

    
    builder
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.token = action.payload.tokens.accessToken;
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    
    builder
      .addCase(register.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(register.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.token = action.payload.tokens.accessToken;
      })
      .addCase(register.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      });

    // Смена пароля → ротация токена.
    builder
      .addCase(changePassword.fulfilled, (state, action) => {
        if (action.payload?.user) state.user = action.payload.user;
        if (action.payload?.tokens?.accessToken) {
          state.token = action.payload.tokens.accessToken;
        }
      });
  },
});

export const { logout, clearError } = authSlice.actions;
export default authSlice.reducer;
