import type { QuestionType, QuizOption, QuizQuestion } from "./types";

export function makeId(prefix: string) {
  return `${prefix}-${Math.random().toString(16).slice(2)}-${Date.now().toString(16)}`;
}

export function createEmptyOption(): QuizOption {
  return { id: makeId("opt"), text: "Новый вариант", is_correct: false };
}

export function createEmptyQuestion(type: QuestionType, timerDefaultSec: number): QuizQuestion {
  const base: QuizQuestion = {
    id: makeId("q"),
    type,
    text: "Новый вопрос",
    timer_sec: timerDefaultSec,
    score: 1000,
    options: [],
    media: [],
  };

  if (type === "single_choice") {
    const o1 = createEmptyOption();
    const o2 = createEmptyOption();
    o1.text = "Вариант 1";
    o2.text = "Вариант 2";
    return { ...base, options: [o1, o2] };
  }

  if (type === "multiple_choice") {
    const o1 = createEmptyOption();
    const o2 = createEmptyOption();
    o1.text = "Вариант 1";
    o2.text = "Вариант 2";
    return { ...base, options: [o1, o2] };
  }

  return { ...base, options: [], correct_text_answers: [""] };
}

export function cloneQuestion(q: QuizQuestion): QuizQuestion {
  return {
    ...q,
    id: makeId("q"),
    options: q.options.map((o) => ({ ...o, id: makeId("opt") })),
    correct_text_answers: q.correct_text_answers ? [...q.correct_text_answers] : undefined,
    media: q.media ? q.media.map((m) => ({ ...m, id: makeId("m") })) : [],
  };
}

export function coerceQuestionToType(q: QuizQuestion, nextType: QuestionType, timerDefaultSec: number): QuizQuestion {
  if (q.type === nextType) return q;
  const kept = {
    id: q.id,
    text: q.text,
    timer_sec: q.timer_sec ?? timerDefaultSec,
    score: q.score ?? 1000,
  };

  if (nextType === "free_text") {
    return {
      ...kept,
      type: "free_text",
      options: [],
      correct_text_answers: q.correct_text_answers?.length ? [...q.correct_text_answers] : [""],
    };
  }

  const existingTexts = q.options?.map((o) => o.text).filter(Boolean) ?? [];
  const options =
    existingTexts.length >= 2
      ? existingTexts.slice(0, 6).map((t) => ({ id: makeId("opt"), text: t, is_correct: false }))
      : [createEmptyOption(), createEmptyOption()].map((o, idx) => ({ ...o, text: `Вариант ${idx + 1}` }));

  return {
    ...kept,
    type: nextType,
    options,
  };
}

export function setSingleCorrect(options: QuizOption[], optionId: string, value: boolean): QuizOption[] {
  return options.map((o) => {
    if (o.id === optionId) return { ...o, is_correct: value };
    return value ? { ...o, is_correct: false } : o;
  });
}
