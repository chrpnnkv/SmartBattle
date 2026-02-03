import { createBus } from "./mockBus";
import type { Quiz } from "../api/types";
import { toPublicQuestion } from "./toPublic";

type AnswerIn = {
  question_id: string;
  selected_option_ids?: string[];
  text_answer?: string;
};

export type Participant = {
  client_id: string;
  nickname: string;
  player_id: string;
  joined_at: string;
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

export type LeaderboardEntry = {
  rank: number;
  nickname: string;
  total_score: number;
};

function makeClientId() {
  return `t-${Math.random().toString(16).slice(2)}-${Date.now().toString(16)}`;
}

function norm(s: string) {
  return s.trim().toLowerCase();
}
function arraysEqualAsSets(a: string[], b: string[]) {
  const A = new Set(a);
  const B = new Set(b);
  if (A.size !== B.size) return false;
  for (const x of A) if (!B.has(x)) return false;
  return true;
}

export class TeacherRealtimeHost {
  private pin: string;
  private clientId: string;
  private bus: ReturnType<typeof createBus> | null = null;

  private participants: Participant[] = [];
  private answersByClient: Record<string, Record<string, AnswerIn>> = {};
  private resultsByClient: Record<string, StudentResult> = {};

  constructor(pin: string, clientId?: string) {
    this.pin = pin;
    this.clientId = clientId ?? makeClientId();
  }

  getClientId() {
    return this.clientId;
  }

  connect() {
    if (this.bus) return;
    this.bus = createBus(this.pin, this.clientId, "teacher");
  }

  disconnect() {
    this.bus?.close();
    this.bus = null;
  }

  getLeaderboard(): LeaderboardEntry[] {
    const rows = Object.values(this.resultsByClient).map((r) => ({
      nickname: r.nickname,
      total_score: r.total_score,
    }));

    rows.sort((a, b) => b.total_score - a.total_score);

    return rows.map((r, i) => ({
      rank: i + 1,
      nickname: r.nickname,
      total_score: r.total_score,
    }));
  }

  onJoin(handler: (p: Participant) => void) {
    this.connect();
    const bus = this.mustBus();

    return bus.on("session:join", (msg) => {
      const { pin: joinPin, nickname } = msg.payload as any;
      if (String(joinPin) !== String(this.pin)) return;

      const client_id = msg.meta.from;
      const player_id = `player-${client_id.slice(-6)}`;

      const p: Participant = {
        client_id,
        nickname,
        player_id,
        joined_at: new Date().toISOString(),
      };

      if (!this.participants.some((x) => x.client_id === client_id)) {
        this.participants.push(p);
      }
      if (!this.answersByClient[client_id]) this.answersByClient[client_id] = {};

      bus.send(
        "session:joined",
        {
          session_id: `sess-${this.pin}`,
          participant_token: `token-${this.pin}-${client_id}`,
          player_id,
          current_state: "waiting_lobby",
        },
        client_id
      );

      handler(p);
    });
  }

  onAnswer(handler: (clientId: string, payload: AnswerIn) => void) {
    this.connect();
    const bus = this.mustBus();

    return bus.on("game:answer", (msg) => {
      const client_id = msg.meta.from;
      const payload = msg.payload as AnswerIn;

      if (!this.answersByClient[client_id]) this.answersByClient[client_id] = {};
      this.answersByClient[client_id][payload.question_id] = payload;

      bus.send("game:answer_ack", { status: "accepted" }, client_id);
      handler(client_id, payload);
    });
  }

  onResultsRequest(handler?: (clientId: string) => void) {
    this.connect();
    const bus = this.mustBus();

    return bus.on("game:results_request", (msg) => {
      const client_id = msg.meta.from;
      handler?.(client_id);

      const r = this.resultsByClient[client_id];
      if (r) bus.send("game:results", r, client_id);
    });
  }

  startQuiz(quiz: Quiz, index = 0) {
    this.connect();
    this.mustBus().send("game:question_start", toPublicQuestion(quiz.questions[index]));
  }

  sendQuestion(quiz: Quiz, index: number) {
    this.connect();
    this.mustBus().send("game:question_start", toPublicQuestion(quiz.questions[index]));
  }

  finishAndSendResults(quiz: Quiz) {
    this.connect();
    const bus = this.mustBus();

    for (const p of this.participants) {
      const byQ = this.answersByClient[p.client_id] ?? {};
      let total = 0;

      const items: StudentResult["items"] = quiz.questions.map((q) => {
        const ans = byQ[q.id];
        const yourSelected = ans?.selected_option_ids ?? [];
        const yourText = ans?.text_answer ?? "";

        const correctOptionIds = (q.options ?? []).filter((o) => o.is_correct).map((o) => o.id);
        const correctTextAnswers = q.correct_text_answers ?? [];

        let isCorrect = false;

        if (q.type === "free_text") {
          const n = norm(yourText);
          isCorrect = n.length > 0 && correctTextAnswers.some((x) => norm(x) === n);
        } else if (q.type === "single_choice") {
          isCorrect = yourSelected.length === 1 && correctOptionIds.includes(yourSelected[0]);
        } else {
          isCorrect = arraysEqualAsSets(yourSelected, correctOptionIds) && correctOptionIds.length > 0;
        }

        const scoreAwarded = isCorrect ? q.score : 0;
        total += scoreAwarded;

        return {
          question_id: q.id,
          question_text: q.text,
          type: q.type,
          your_selected_option_ids: yourSelected,
          your_text_answer: yourText,
          correct_option_ids: correctOptionIds,
          correct_text_answers: correctTextAnswers,
          is_correct: isCorrect,
          score_awarded: scoreAwarded,
        };
      });

      const payload: StudentResult = {
        session_id: `sess-${this.pin}`,
        quiz_id: quiz.id,
        nickname: p.nickname,
        total_score: total,
        items,
      };

      this.resultsByClient[p.client_id] = payload;
      bus.send("game:results", payload, p.client_id);
    }

    bus.send("game:finished", { reason: "end_of_quiz" });
  }

  reset() {
    this.connect();
    this.participants = [];
    this.answersByClient = {};
    this.resultsByClient = {};
    this.mustBus().send("session:reset", {});
  }

  private mustBus() {
    if (!this.bus) throw new Error("TeacherRealtimeHost is not connected");
    return this.bus;
  }
}
