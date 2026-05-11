import { describe, it, expect, vi, beforeEach } from 'vitest';
import { RealWebSocketService } from '../api/wsService';

class MockWS {
  static OPEN = 1;
  readyState = MockWS.OPEN;
  send = vi.fn();
  close = vi.fn();
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  onclose: (() => void) | null = null;
}

let fakeWs!: MockWS;

describe('RealWebSocketService', () => {
  let service: RealWebSocketService;

  beforeEach(() => {
    vi.stubGlobal('WebSocket', class extends MockWS {
      constructor(..._args: unknown[]) {
        super();
        fakeWs = this;
      }
    });
    service = new RealWebSocketService();
    service.connect('sess1', { roomCode: 'ABCD12', name: 'Alice', token: 'jwt123' });
    fakeWs.onopen?.();
  });

  describe('join handshake', () => {
    it('sends join message immediately after open', () => {
      expect(fakeWs.send).toHaveBeenCalled();
      const msg = JSON.parse(fakeWs.send.mock.calls[0][0] as string);
      expect(msg).toEqual({ type: 'join', room_code: 'ABCD12', name: 'Alice', token: 'jwt123' });
    });
  });

  describe('send', () => {
    it('merges payload into flat object', () => {
      service.send('answer', { question_id: 'q1', answer_id: 'a1' });
      const msg = JSON.parse(fakeWs.send.mock.calls[fakeWs.send.mock.calls.length - 1][0] as string);
      expect(msg).toEqual({ type: 'answer', question_id: 'q1', answer_id: 'a1' });
    });

    it('sends only type when payload is omitted', () => {
      service.send('next_question');
      const msg = JSON.parse(fakeWs.send.mock.calls[fakeWs.send.mock.calls.length - 1][0] as string);
      expect(msg).toEqual({ type: 'next_question' });
    });

    it('sends only type for empty object payload', () => {
      service.send('start_session', {});
      const msg = JSON.parse(fakeWs.send.mock.calls[fakeWs.send.mock.calls.length - 1][0] as string);
      expect(msg).toEqual({ type: 'start_session' });
    });
  });

  describe('message routing', () => {
    it('routes incoming message to registered handler', () => {
      const handler = vi.fn();
      service.on('question_started', handler);
      fakeWs.onmessage?.({ data: JSON.stringify({ type: 'question_started', payload: { questionIndex: 0 } }) });
      expect(handler).toHaveBeenCalledWith({ questionIndex: 0 });
    });

    it('passes null when payload is absent', () => {
      const handler = vi.fn();
      service.on('joined', handler);
      fakeWs.onmessage?.({ data: JSON.stringify({ type: 'joined' }) });
      expect(handler).toHaveBeenCalledWith(null);
    });

    it('ignores messages with no registered handler', () => {
      expect(() =>
        fakeWs.onmessage?.({ data: JSON.stringify({ type: 'unknown_event' }) })
      ).not.toThrow();
    });
  });

  describe('disconnect', () => {
    it('closes the socket', () => {
      service.disconnect();
      expect(fakeWs.close).toHaveBeenCalled();
    });

    it('isConnected returns false after disconnect', () => {
      fakeWs.readyState = 3;
      service.disconnect();
      expect(service.isConnected()).toBe(false);
    });

    it('does not deliver messages after disconnect', () => {
      const handler = vi.fn();
      service.on('some_event', handler);
      service.disconnect();
      fakeWs.onmessage?.({ data: JSON.stringify({ type: 'some_event', payload: {} }) });
      expect(handler).not.toHaveBeenCalled();
    });
  });

  describe('send before socket is open', () => {
    it('does not throw and does not send when readyState is CONNECTING', () => {
      fakeWs.readyState = 0;
      expect(() => service.send('start_session', {})).not.toThrow();
      const sentAfterOpen = fakeWs.send.mock.calls.filter(
        (c) => JSON.parse(c[0] as string).type === 'start_session'
      );
      expect(sentAfterOpen).toHaveLength(0);
    });
  });
});