import { createBus } from "./mockBus";

export type PublicQuestion = {
  question_id: string;
  text: string;
  type: "single_choice" | "multiple_choice" | "free_text";
  options: { id: string; text: string }[];
  time_limit_sec: number;
  started_at: string;
  media?: { id: string; kind: "image" | "video" | "file"; title?: string; url: string }[];
};

export type JoinOk = {
  session_id: string;
  participant_token: string;
  player_id: string;
  current_state: "waiting_lobby" | string;
};

export type StudentAnswer = {
  question_id: string;
  selected_option_ids: string[];
  text_answer: string;
};

export type StudentResult = {
  session_id: string;
  quiz_id: string;
  nickname: string;
  total_score: number;
  items: Array<{
    question_id: string;
    question_text: string;
    type: "single_choice" | "multiple_choice" | "free_text";
    your_selected_option_ids: string[];
    your_text_answer: string;
    correct_option_ids: string[];
    correct_text_answers: string[];
    is_correct: boolean;
    score_awarded: number;
  }>;
};

function makeClientId() {
  return `s-${Math.random().toString(16).slice(2)}-${Date.now().toString(16)}`;
}

export class StudentRealtimeClient {
  private pin: string;
  private nickname: string;
  private clientId: string;
  private bus: ReturnType<typeof createBus> | null = null;

  constructor(pin: string, nickname: string, clientId?: string) {
    this.pin = pin;
    this.nickname = nickname;
    this.clientId = clientId ?? makeClientId();
  }

  getClientId() {
    return this.clientId;
  }

  connect() {
    if (this.bus) return;
    this.bus = createBus(this.pin, this.clientId, "student");
  }

  disconnect() {
    this.bus?.close();
    this.bus = null;
  }

  async join(): Promise<JoinOk> {
    this.connect();
    const bus = this.mustBus();

    return new Promise((resolve, reject) => {
      let done = false;

      const off = bus.on("session:joined", (msg) => {
        done = true;
        cleanup();
        resolve(msg.payload as JoinOk);
      });

      const t = window.setTimeout(() => {
        if (done) return;
        cleanup();
        reject(new Error("NO_TEACHER"));
      }, 10000);

      const retry = window.setInterval(() => {
        if (done) return;
        bus.send("session:join", { pin: this.pin, nickname: this.nickname });
      }, 300);

      const cleanup = () => {
        off();
        window.clearTimeout(t);
        window.clearInterval(retry);
      };

      bus.send("session:join", { pin: this.pin, nickname: this.nickname });
    });
  }

  onQuestionStart(handler: (q: PublicQuestion) => void) {
    this.connect();
    const bus = this.mustBus();
    return bus.on("game:question_start", (msg) => handler(msg.payload as PublicQuestion));
  }

  onFinished(handler: () => void) {
    this.connect();
    const bus = this.mustBus();
    return bus.on("game:finished", () => handler());
  }

  onReset(handler: () => void) {
    this.connect();
    const bus = this.mustBus();
    return bus.on("session:reset", () => handler());
  }

  sendAnswer(answer: StudentAnswer) {
    this.connect();
    this.mustBus().send("game:answer", answer);
  }

  requestResults() {
    this.connect();
    this.mustBus().send("game:results_request", {});
  }

  onResults(handler: (r: StudentResult) => void) {
    this.connect();
    const bus = this.mustBus();
    return bus.on("game:results", (msg) => handler(msg.payload as StudentResult));
  }

  private mustBus() {
    if (!this.bus) throw new Error("StudentRealtimeClient is not connected");
    return this.bus;
  }
}
