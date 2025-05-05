import "./App.css";
import WebSocketClient from "./Websocket";
import { BrowserRouter, Routes, Route } from "react-router";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/websocket" element={<WebSocketClient />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
