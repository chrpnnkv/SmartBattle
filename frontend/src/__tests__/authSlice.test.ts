import { describe, it, expect, beforeEach, vi } from 'vitest';
function makeMemoryStorage(): Storage {
  const store = new Map<string, string>();
  return {
    get length() {
      return store.size;
    },
    clear() {
      store.clear();
    },
    getItem(key: string) {
      return store.has(key) ? store.get(key)! : null;
    },
    setItem(key: string, value: string) {
      store.set(key, String(value));
    },
    removeItem(key: string) {
      store.delete(key);
    },
    key(index: number) {
      return Array.from(store.keys())[index] ?? null;
    },
  };
}

vi.stubGlobal('localStorage', makeMemoryStorage());
import reducer, { changePassword, logout } from '../store/slices/authSlice';
import type { User } from '../types';

const mockUser: User = {
  id: 'u1',
  email: 'a@b.c',
  name: 'Alice',
  role: 'teacher',
};

const initialState = {
  user: mockUser,
  token: 'old-token',
  isLoading: false,
  error: null,
  isInitialized: true,
};

describe('authSlice.changePassword', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('replaces token in state when fulfilled with new token', () => {
    const fulfilledAction = {
      type: changePassword.fulfilled.type,
      payload: {
        message: 'password changed successfully',
        user: { ...mockUser, name: 'Alice Updated' },
        tokens: { accessToken: 'new-rotated-token' },
      },
    };

    const next = reducer(initialState, fulfilledAction);

    expect(next.token).toBe('new-rotated-token');
    expect(next.user?.name).toBe('Alice Updated');
  });

  it('keeps state intact if fulfilled payload has no token', () => {
    const fulfilledAction = {
      type: changePassword.fulfilled.type,
      payload: { message: 'ok' },
    };

    const next = reducer(initialState, fulfilledAction);

    expect(next.token).toBe('old-token');
    expect(next.user).toEqual(mockUser);
  });

  it('keeps state intact if fulfilled payload is undefined', () => {
    const fulfilledAction = {
      type: changePassword.fulfilled.type,
      payload: undefined,
    };

    const next = reducer(initialState, fulfilledAction);

    expect(next.token).toBe('old-token');
    expect(next.user).toEqual(mockUser);
  });
});

describe('authSlice.logout', () => {
  it('clears user, token and localStorage', () => {
    localStorage.setItem('accessToken', 'some-token');
    const next = reducer(initialState, logout());
    expect(next.user).toBeNull();
    expect(next.token).toBeNull();
    expect(localStorage.getItem('accessToken')).toBeNull();
  });
});
