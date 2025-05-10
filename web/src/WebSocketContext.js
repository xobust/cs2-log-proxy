import React, { createContext, useContext, useEffect, useRef, useState } from 'react';

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const wsRef = useRef(null);
  const [status, setStatus] = useState('disconnected'); // 'connecting', 'connected', 'disconnected', 'error'

  useEffect(() => {
    setStatus('connecting');
    const ws = new WebSocket(`ws://localhost:8081/ws`);
    wsRef.current = ws;

    ws.onopen = () => setStatus('connected');
    ws.onclose = () => setStatus('disconnected');
    ws.onerror = () => setStatus('error');

    return () => {
      ws.close();
      wsRef.current = null;
      setStatus('disconnected');
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ ws: wsRef.current, status }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => useContext(WebSocketContext);
