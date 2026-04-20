import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { configureStore } from '@reduxjs/toolkit';
import JoinPage from '../pages/JoinPage';
import sessionReducer from '../store/slices/sessionSlice';

vi.mock('../api', () => ({
  api: {
    sessions: {
      joinSession: vi.fn().mockReturnValue(new Promise(() => {})),
      getSession: vi.fn(),
    },
  },
}));

function createStore() {
  return configureStore({ reducer: { session: sessionReducer } });
}

function renderJoinPage() {
  render(
    <Provider store={createStore()}>
      <MemoryRouter initialEntries={['/join']}>
        <Routes>
          <Route path="/join" element={<JoinPage />} />
          <Route path="/session/:id/waiting" element={<div>Waiting</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>
  );
}

function fillPin(digits: string) {
  const inputs = screen.getAllByRole('textbox').slice(0, 6);
  digits.split('').forEach((digit, i) => {
    fireEvent.change(inputs[i], { target: { value: digit } });
  });
}

describe('JoinPage', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  describe('validation', () => {
    it('shows PIN error when PIN is incomplete', () => {
      renderJoinPage();
      fireEvent.click(screen.getByRole('button', { name: /войти/i }));
      expect(screen.getByText(/6-значный PIN/i)).toBeInTheDocument();
    });

    it('shows nickname error when nickname is empty', () => {
      renderJoinPage();
      fillPin('123456');
      fireEvent.click(screen.getByRole('button', { name: /войти/i }));
      expect(screen.getByText(/введите никнейм/i)).toBeInTheDocument();
    });

    it('shows nickname error when nickname is too short', () => {
      renderJoinPage();
      fillPin('123456');
      fireEvent.change(screen.getByPlaceholderText(/введите ваше имя/i), { target: { value: 'A' } });
      fireEvent.click(screen.getByRole('button', { name: /войти/i }));
      expect(screen.getByText(/минимум 2/i)).toBeInTheDocument();
    });

    it('shows no errors on valid input', () => {
      renderJoinPage();
      fillPin('123456');
      fireEvent.change(screen.getByPlaceholderText(/введите ваше имя/i), { target: { value: 'Alice' } });
      fireEvent.click(screen.getByRole('button', { name: /войти/i }));
      expect(screen.queryByText(/6-значный PIN/i)).toBeNull();
      expect(screen.queryByText(/введите никнейм/i)).toBeNull();
    });
  });

  describe('submit', () => {
    it('saves PIN to sessionStorage on valid submit', () => {
      renderJoinPage();
      fillPin('123456');
      fireEvent.change(screen.getByPlaceholderText(/введите ваше имя/i), { target: { value: 'Alice' } });
      fireEvent.click(screen.getByRole('button', { name: /войти/i }));
      expect(sessionStorage.getItem('sb_pin')).toBe('123456');
    });
  });
});