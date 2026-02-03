export type MediaKind = "image" | "video" | "file";

export type MediaAttachment = {
  id: string;
  kind: MediaKind;
  title?: string;
  url: string;
};

export type JoinedPayload = {
  session_id: string;
  participant_token: string;
  player_id: string;
  current_state: "waiting_lobby";
};

export type PublicQuestion = {
  question_id: string;
  text: string;
  type: "single_choice" | "multiple_choice" | "free_text";
  options: { id: string; text: string }[];
  time_limit_sec: number;
  started_at: string;
  media?: MediaAttachment[];
};

const questionBank: PublicQuestion[] = [
  {
    question_id: "uuid-q1",
    text: 'Что выведет fmt.Println(len("abc"))?',
    type: "single_choice",
    options: [
      { id: "opt-1", text: "3" },
      { id: "opt-2", text: "4" },
      { id: "opt-3", text: "panic" },
    ],
    time_limit_sec: 15,
    started_at: new Date().toISOString(),
    media: [
      {
        id: "m1",
        kind: "image",
        title: "Подсказка (gif/img)",
        url: "https://upload.wikimedia.org/wikipedia/commons/2/23/Golang.png",
      },
    ],
  },
  {
    question_id: "uuid-q2",
    text: "Выберите ссылочные типы в Go",
    type: "multiple_choice",
    options: [
      { id: "opt-4", text: "slice" },
      { id: "opt-5", text: "map" },
      { id: "opt-6", text: "int" },
    ],
    time_limit_sec: 25,
    started_at: new Date().toISOString(),
    media: [],
  },
];

export async function mockJoin(pin: string, nickname: string): Promise<JoinedPayload> {
  if (pin.trim().length < 4) {
    throw new Error("ROOM_NOT_FOUND");
  }
  sessionStorage.setItem("mock_question_index", "0");

  return {
    session_id: `sess-${pin}`,
    participant_token: `token-${pin}-${nickname}`,
    player_id: `player-${nickname}`,
    current_state: "waiting_lobby",
  };
}

export async function mockStartFirstQuestion(): Promise<PublicQuestion> {
  sessionStorage.setItem("mock_question_index", "0");
  return {
    ...questionBank[0],
    started_at: new Date().toISOString(),
  };
}

export async function mockNextQuestion(): Promise<PublicQuestion | null> {
  const idx = Number(sessionStorage.getItem("mock_question_index") ?? "0");
  const nextIdx = idx + 1;
  sessionStorage.setItem("mock_question_index", String(nextIdx));

  const q = questionBank[nextIdx];
  if (!q) return null;

  return { ...q, started_at: new Date().toISOString() };
}

export async function mockSendAnswer(): Promise<{ status: "accepted" }> {
  return { status: "accepted" };
}
