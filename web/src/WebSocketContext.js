import React, { createContext, useContext, useEffect, useRef, useState } from 'react';

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const [status, setStatus] = useState('disconnected');
  const wsRef = useRef(null);
  const reconnectTimeout = useRef(null);
  const reconnectAttempts = useRef(0);

  const connect = () => {
    setStatus('connecting');
    const ws = new WebSocket(`ws://localhost:8081/ws`);
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus('connected');
      reconnectAttempts.current = 0;
    };
    ws.onclose = () => {
      setStatus('disconnected');
      scheduleReconnect();
    };
    ws.onerror = () => {
      setStatus('error');
      ws.close(); // triggers onclose
    };

    return () => {
      ws.close();
      wsRef.current = null;
      setStatus('disconnected');
    };
  };

  const scheduleReconnect = () => {
    if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 15000); // exponential backoff, max 15s
    reconnectTimeout.current = setTimeout(() => {
      reconnectAttempts.current += 1;
      connect();
    }, delay);
  };

  useEffect(() => {
    connect();
    return () => {
      if (wsRef.current) wsRef.current.close();
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    };
  }, []);

  // --- Enhanced Subscription/Event API ---
  const listenersRef = useRef({}); // { eventType: Set of callbacks }
  const activeSubsRef = useRef({}); // { eventType: true }

  // Register a callback for a specific event type
  const subscribe = (eventType, callback, options = {}) => {
    if (!listenersRef.current[eventType]) listenersRef.current[eventType] = new Set();
    listenersRef.current[eventType].add(callback);
    // Only send subscribe message if this is the first listener for this event
    if (
      !activeSubsRef.current[eventType] &&
      wsRef.current &&
      wsRef.current.readyState === WebSocket.OPEN
    ) {
      wsRef.current.send(JSON.stringify({ type: 'subscribe', event: eventType, ...options }));
      activeSubsRef.current[eventType] = true;
    }
    return () => unsubscribe(eventType, callback, options);
  };

  const unsubscribe = (eventType, callback, options = {}) => {
    if (listenersRef.current[eventType]) {
      listenersRef.current[eventType].delete(callback);
      if (
        listenersRef.current[eventType].size === 0 &&
        wsRef.current &&
        wsRef.current.readyState === WebSocket.OPEN
      ) {
        wsRef.current.send(JSON.stringify({ type: 'unsubscribe', event: eventType, ...options }));
        activeSubsRef.current[eventType] = false;
      }
    }
  };

  // Centralized message handler
  useEffect(() => {
    if (!wsRef.current) return;
    const handleMessage = (event) => {
      let data;
      try {
        data = JSON.parse(event.data);
      } catch {
        return;
      }
      console.log('WebSocket message:', data);
      const { type } = data;
      if (listenersRef.current[type]) {
        listenersRef.current[type].forEach((cb) => cb(data));
      }
    };
    wsRef.current.addEventListener('message', handleMessage);
    return () => wsRef.current && wsRef.current.removeEventListener('message', handleMessage);
  }, [status]);

  return (
    <WebSocketContext.Provider value={{ ws: wsRef.current, status, subscribe, unsubscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => useContext(WebSocketContext);
