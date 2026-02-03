import type { Quiz } from "./types";
import { http } from "../http";

export const coreApi = {

  getQuiz: (id: string) => http<Quiz>(`/api/quizzes/${id}`, { auth: "none" }),
  updateQuiz: (id: string, quiz: Quiz) =>
    http<Quiz>(`/api/quizzes/${id}`, {
      method: "PUT",
      body: JSON.stringify(quiz),
      auth: "teacher",
    }),
};
