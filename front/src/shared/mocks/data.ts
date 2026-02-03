import type { Quiz, QuizListItem } from "../api/types";

export const quizSeed: Quiz = {
  id: "550e8400-e29b-41d4-a716-446655440000",
  title: "Основы Go и микросервисов",
  description: "Проверка знаний по итогам 1 модуля",
  teacher_id: "uuid-teacher-1",
  created_at: "2026-01-21T10:00:00Z",
  settings: { timer_default_sec: 30 },
  questions: [
    {
      id: "uuid-q1",
      type: "single_choice",
      text: 'Что выведет fmt.Println(len("abc"))?',
      timer_sec: 15,
      score: 1000,
      options: [
        { id: "opt-1", text: "3", is_correct: true },
        { id: "opt-2", text: "4", is_correct: false },
        { id: "opt-3", text: "panic", is_correct: false },
      ],
    },
    {
      id: "uuid-q2",
      type: "multiple_choice",
      text: "Выберите ссылочные типы в Go",
      timer_sec: 30,
      score: 1500,
      options: [
        { id: "opt-4", text: "slice", is_correct: true },
        { id: "opt-5", text: "map", is_correct: true },
        { id: "opt-6", text: "int", is_correct: false },
      ],
    },
    {
      id: "uuid-q3",
      type: "free_text",
      text: "Назовите ключевое слово для создания горутины",
      timer_sec: 45,
      score: 2000,
      options: [],
      correct_text_answers: ["go", "Go"],
    },
  ],
};

const quizzesById: Record<string, Quiz> = {
  [quizSeed.id]: structuredClone(quizSeed),
  q2: {
    id: "q2",
    title: "Databases Basics",
    description: "Быстрый квиз по основам БД",
    teacher_id: "uuid-teacher-1",
    created_at: "2026-01-22T12:30:00Z",
    settings: { timer_default_sec: 25 },
    questions: [],
  },
};

export function getQuiz(id: string): Quiz | null {
  const q = quizzesById[id];
  return q ? structuredClone(q) : null;
}

export function saveQuiz(id: string, quiz: Quiz) {
  quizzesById[id] = structuredClone(quiz);
}

export function listQuizzes(): QuizListItem[] {
  return Object.values(quizzesById).map((q) => ({
    id: q.id,
    title: q.title,
    questionCount: q.questions.length,
    updatedAt: q.created_at,
  }));
}
