import { createSlice } from "@reduxjs/toolkit";
import type { PayloadAction } from "@reduxjs/toolkit";

type AuthState = {
  teacherToken: string | null;
};

const initialState: AuthState = {
  teacherToken: null,
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    setTeacherToken(state, action: PayloadAction<string>) {
      state.teacherToken = action.payload;
    },
    clearTeacherToken(state) {
      state.teacherToken = null;
    },
  },
});

export const { setTeacherToken, clearTeacherToken } = authSlice.actions;
export default authSlice.reducer;
