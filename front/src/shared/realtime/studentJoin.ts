import { createBus } from "./mockBus";

function makeClientId() {
  return `s-${Math.random().toString(16).slice(2)}-${Date.now().toString(16)}`;
}

export async function studentJoin(pin: string, nickname: string) {
  const clientId = makeClientId();
  const bus = createBus(pin, clientId, "student");

  const joined = await new Promise<any>((resolve, reject) => {
    let done = false;

    const stop = () => {
      done = true;
      offJoined();
      window.clearTimeout(timeout);
      window.clearInterval(retry);
      bus.close();
    };

    const offJoined = bus.on("session:joined", (msg) => {
      stop();
      resolve(msg.payload);
    });

    const timeout = window.setTimeout(() => {
      if (done) return;
      stop();
      reject(new Error("NO_TEACHER"));
    }, 10000);

    const retry = window.setInterval(() => {
      if (done) return;
      bus.send("session:join", { pin, nickname });
    }, 300);

    bus.send("session:join", { pin, nickname });
  });

  return { joined, clientId };
}
