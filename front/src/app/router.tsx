import { createBrowserRouter } from "react-router-dom";
import App from "../App";

import LoginPage from "../features/pages/auth/LoginPage";
import TeacherDashboardPage from "../features/pages/teacher/TeacherDashboardPage";
import QuizBuilderPage from "../features/pages/teacher/QuizBuilderPage";
import TeacherSessionPage from "../features/pages/teacher/TeacherSessionPage";

import HomePage from "../features/pages/HomePage";
import JoinPage from "../features/pages/student/JoinPage";
import StudentLobbyPage from "../features/pages/student/StudentLobbyPage";
import StudentPlayPage from "../features/pages/student/StudentPlayPage";
import StudentFinishPage from "../features/pages/student/StudentFinishPage";
import StudentWaitNextPage from "../features/pages/student/StudentWaitNextPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <HomePage /> },

      { path: "join", element: <JoinPage /> }, // QR / PIN -> сюда

      // student flow
      { path: "s/:sessionId/lobby", element: <StudentLobbyPage /> },
      { path: "s/:sessionId/play", element: <StudentPlayPage /> },
      { path: "s/:sessionId/finish", element: <StudentFinishPage /> },
      { path: "s/:sessionId/wait", element: <StudentWaitNextPage /> },

      // teacher
      { path: "login", element: <LoginPage /> },
      { path: "teacher", element: <TeacherDashboardPage /> },
      { path: "teacher/quizzes/:quizId", element: <QuizBuilderPage /> },
     { path: "teacher/session/:pin", element: <TeacherSessionPage /> },
    ],
  },
]);
