import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAppSelector } from '../hooks/redux';

import LandingPage from '../pages/LandingPage';
import LoginPage from '../pages/LoginPage';
import RegisterPage from '../pages/RegisterPage';
import ForgotPasswordPage from '../pages/ForgotPasswordPage';
import ResetPasswordPage from '../pages/ResetPasswordPage';
import ProfilePage from '../pages/ProfilePage';
import DashboardPage from '../pages/DashboardPage';
import QuizBuilderPage from '../pages/QuizBuilderPage';
import JoinPage from '../pages/JoinPage';
import WaitingRoomPage from '../pages/WaitingRoomPage';
import QuestionPage from '../pages/QuestionPage';
import FinishedPage from '../pages/FinishedPage';
import AnalyticsPage from '../pages/AnalyticsPage';
import ReportsPage from '../pages/ReportsPage';
import NotFoundPage from '../pages/NotFoundPage';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, isInitialized } = useAppSelector((s) => s.auth);
  if (!isInitialized) return null;
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function PublicOnlyRoute({ children }: { children: React.ReactNode }) {
  const { user, isInitialized } = useAppSelector((s) => s.auth);
  if (!isInitialized) return null;
  if (user) return <Navigate to="/dashboard" replace />;
  return <>{children}</>;
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/join" element={<JoinPage />} />
        <Route path="/join/:pin" element={<JoinPage />} />
        <Route path="/login"    element={<PublicOnlyRoute><LoginPage /></PublicOnlyRoute>} />
        <Route path="/register" element={<PublicOnlyRoute><RegisterPage /></PublicOnlyRoute>} />
        <Route path="/forgot-password" element={<PublicOnlyRoute><ForgotPasswordPage /></PublicOnlyRoute>} />
        <Route path="/reset-password"  element={<PublicOnlyRoute><ResetPasswordPage /></PublicOnlyRoute>} />
        <Route path="/session/:sessionId/waiting"  element={<WaitingRoomPage />} />
        <Route path="/session/:sessionId/question" element={<QuestionPage />} />
        <Route path="/session/:sessionId/finished" element={<FinishedPage />} />
        <Route path="/dashboard"  element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        <Route path="/profile"    element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
        <Route path="/quiz/new"   element={<ProtectedRoute><QuizBuilderPage /></ProtectedRoute>} />
        <Route path="/quiz/:id/edit" element={<ProtectedRoute><QuizBuilderPage /></ProtectedRoute>} />
        <Route path="/session/:sessionId/analytics" element={<ProtectedRoute><AnalyticsPage /></ProtectedRoute>} />
        <Route path="/reports"    element={<ProtectedRoute><ReportsPage /></ProtectedRoute>} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}