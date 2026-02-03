import { http, HttpResponse } from "msw";
import { getQuiz, listQuizzes, saveQuiz } from "./data";
import type { Quiz } from "../api/types";

export const handlers = [
  http.get("/api/quizzes", () => {
    return HttpResponse.json({ items: listQuizzes() });
  }),

  http.get("/api/quizzes/:quizId", ({ params }) => {
    const quizId = String(params.quizId);
    const quiz = getQuiz(quizId);
    if (!quiz) return new HttpResponse(null, { status: 404 });
    return HttpResponse.json(quiz);
  }),

  http.put("/api/quizzes/:quizId", async ({ params, request }) => {
    const quizId = String(params.quizId);
    const incoming = (await request.json()) as Quiz;

    if (!incoming?.id || incoming.id !== quizId) {
      return HttpResponse.json({ message: "Bad quiz payload" }, { status: 400 });
    }

    saveQuiz(quizId, incoming);
    return HttpResponse.json({ status: "ok" });
  }),
];
