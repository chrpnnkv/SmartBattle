export type Envelope<T = any> = {
  event: string;
  payload: T;
  timestamp: string;
  meta: {
    from: string;
    to?: string; 
    role: "teacher" | "student";
  };
};

type Handler = (msg: Envelope) => void;

export function createBus(pin: string, clientId: string, role: "teacher" | "student") {
  const channel = new BroadcastChannel(`quiz-room-${pin}`);
  const handlers = new Map<string, Set<Handler>>();

  channel.onmessage = (e) => {
    const msg = e.data as Envelope;

    if (msg?.meta?.from === clientId) return;

    if (msg?.meta?.to && msg.meta.to !== clientId) return;

    const set = handlers.get(msg.event);
    if (!set) return;
    for (const h of set) h(msg);
  };

  function send<T>(event: string, payload: T, to?: string) {
    const msg: Envelope<T> = {
      event,
      payload,
      timestamp: new Date().toISOString(),
      meta: { from: clientId, to, role },
    };
    channel.postMessage(msg);
  }

  function on(event: string, handler: Handler) {
    const set = handlers.get(event) ?? new Set<Handler>();
    set.add(handler);
    handlers.set(event, set);
    return () => set.delete(handler);
  }

  function close() {
    handlers.clear();
    channel.close();
  }

  return { send, on, close };
}
