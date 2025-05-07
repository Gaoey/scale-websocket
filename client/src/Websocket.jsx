import React, { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";

const WebSocketClient = () => {
  const [searchParams] = useSearchParams();
  const [isConnected, setIsConnected] = useState(false);
  const token = searchParams.get("token");
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const socketRef = useRef(null);

  useEffect(() => {
    // Connect to the WebSocket server
    console.log("Connecting websocket server...");
    const socket = new WebSocket("ws://localhost:8080/ws?token=" + token);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      console.log("Message from server:", event.data);
      setMessages((prev) => [...prev, event.data]);
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    socket.onclose = () => {
      console.log("WebSocket closed");
    };

    return () => {
      socket.close();
    };
  }, []);

  const sendMessage = () => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(input);
      setInput("");
    }
  };

  return (
    <div style={{ padding: "2rem" }}>
      <h2>WebSocket Client</h2>
      {isConnected && (
        <button
          onClick={() => {
            if (socketRef.current && isConnected) {
              socketRef.current.close();
              setIsConnected(false);
            }
          }}
        >
          close
        </button>
      )}
      <div>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message"
        />
        <button onClick={sendMessage}>Send</button>
      </div>
      <h3>Messages:</h3>
      <ul>
        {messages.map((msg, idx) => (
          <li key={idx}>{msg}</li>
        ))}
      </ul>
    </div>
  );
};

export default WebSocketClient;
