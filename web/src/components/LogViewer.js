import React, { useState, useEffect, useRef } from 'react';
import { Box, Paper, Typography } from '@mui/material';

function LogViewer({ token }) {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const wsRef = useRef(null);

  // Fetch the full log when token changes
  useEffect(() => {
    if (!token) return;
    setLoading(true);
    fetch(`/api/logs/${token}`)
      .then(res => res.ok ? res.json() : Promise.reject('Failed to fetch log'))
      .then(data => {
        // Support both array and object with .chunks or .logs
        if (Array.isArray(data)) {
          setLogs(data);
        } else if (data.chunks) {
          setLogs(data.chunks);
        } else if (data.logs) {
          setLogs(data.logs);
        } else {
          setLogs([]);
        }
        setLoading(false);
      })
      .catch(() => {
        setLogs([]);
        setLoading(false);
      });
  }, [token]);


  // Listen for live updates over WebSocket
  useEffect(() => {
    if (!token) return;
    const ws = new WebSocket(`ws://localhost:8081/ws/${token}`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      setLogs(prev => [...prev, event.data]);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    return () => {
      ws.close();
    };
  }, [token]);

  const scrollToBottom = () => {
    const container = document.getElementById('log-container');
    if (container) {
      container.scrollTop = container.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [logs]);

  return (
    <Paper sx={{ p: 2, height: '100vh', overflow: 'auto' }}>
      <Box sx={{ mb: 2 }}>
        <Typography variant="h5">CS2 Log Viewer</Typography>
      </Box>
      <Box id="log-container" sx={{ mb: 2 }}>
        {logs.map((log, index) => (
          <Typography key={index} variant="body1" sx={{ color: 'primary.main' }}>
            {log}
          </Typography>
        ))}
      </Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          {logs.length} logs
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {loading ? 'Connecting...' : 'Connected'}
        </Typography>
      </Box>
      <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
        <button
          onClick={() => {
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
              const message = `Test message ${new Date().toISOString()}`;
              wsRef.current.send(message);
              setLogs((prev) => [...prev, message]);
            }
          }}
          style={{ padding: '8px 16px', background: '#1976d2', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
        >
          Send Test Message
        </button>
      </Box>
    </Paper>
  );
}

export default LogViewer;
