"use client";

import { useChat } from "@ai-sdk/react";
import { useState } from "react";

export default function Chat() {
  const { messages, sendMessage } = useChat({
    // api:"/counterspell/api/chat?secret=dev-token",
  });
  const [input, setInput] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      sendMessage({ text: input });
      setInput("");
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <div style={{ flex: 1, overflowY: "auto", padding: "1rem" }}>
        {messages.map((m) => (
          <div key={m.id}>
            <strong>{m.role === "user" ? "User: " : "AI: "}</strong>
            {m.parts[0].type === "text" && m.parts[0].text}
          </div>
        ))}
      </div>

      <form
        onSubmit={handleSubmit}
        style={{
          padding: "1rem",
          borderTop: "1px solid var(--mantine-color-dark-4)",
        }}
      >
        <input
          value={input}
          placeholder="Say something..."
          onChange={(e) => setInput(e.target.value)}
          style={{
            width: "100%",
            padding: "0.5rem",
            borderRadius: "var(--mantine-radius-md)",
            border: "1px solid var(--mantine-color-dark-4)",
          }}
        />
      </form>
    </div>
  );
}
