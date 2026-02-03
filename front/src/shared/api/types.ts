export type ISODateString = string;

export type QuestionType = "single_choice" | "multiple_choice" | "free_text";

export type QuizSettings = {
  timer_default_sec: number;
};

export type QuizOption = {
  id: string;
  text: string;
  is_correct: boolean; 
};

export type MediaKind = "image" | "video" | "file";

export type MediaAttachment = {
  id: string;
  kind: MediaKind;
  title?: string;     
  url: string;         
  mime?: string;       
};

export type QuizQuestion = {
  id: string;
  type: QuestionType;
  text: string;
  timer_sec: number;
  score: number;
  options: QuizOption[];
  correct_text_answers?: string[];
  media?: MediaAttachment[]; 
};

export type Quiz = {
  id: string;
  title: string;
  description: string;
  teacher_id: string;
  created_at: ISODateString;
  settings: QuizSettings;
  questions: QuizQuestion[];
};

export type QuizListItem = {
  id: string;
  title: string;
  questionCount: number;
  updatedAt: ISODateString;
};
