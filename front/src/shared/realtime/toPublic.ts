import type { QuizQuestion } from "../api/types";

export type PublicQuestion = {
  question_id: string;
  text: string;
  type: "single_choice" | "multiple_choice" | "free_text";
  options: { id: string; text: string }[];
  time_limit_sec: number;
  started_at: string;
  media?: any[];
};

export function toPublicQuestion(q: QuizQuestion): PublicQuestion {
  return {
    question_id: q.id,
    text: q.text,
    type: q.type,
    options: (q.options ?? []).map((o) => ({ id: o.id, text: o.text })),
    time_limit_sec: q.timer_sec,
    started_at: new Date().toISOString(),
    media: q.media ?? [],
  };
}
