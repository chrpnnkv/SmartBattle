import React from "react";
import ReactDOM from "react-dom/client";
import { Provider } from "react-redux";
import { RouterProvider } from "react-router-dom";

import { store } from "./app/store";
import { router } from "./app/router";

import "./shared/styles/globals.css";
import "./shared/styles/theme.css";

async function enableMocks() {
  if (import.meta.env.DEV) {
    try {
      const { worker } = await import("./shared/mocks/browser");
      await worker.start({ onUnhandledRequest: "bypass" });
    } catch (e) {
      console.warn("MSW failed to start, continuing without mocks", e);
    }
  }
}

enableMocks().finally(() => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <Provider store={store}>
        <RouterProvider router={router} />
      </Provider>
    </React.StrictMode>
  );
});

