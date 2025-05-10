import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Switch,
  FormControlLabel,
  Stack,
  Tooltip,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import { useWebSocket } from '../WebSocketContext';

function LogViewer({ token }) {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const containerRef = useRef(null);

  // Fetch the full log when token changes
  useEffect(() => {
    if (!token) return;
    setLoading(true);
    fetch(`/api/logs/${token}`)
      .then((res) => (res.ok ? res.text() : Promise.reject('Failed to fetch log')))
      .then((data) => {
        setLogs(data.split('\n'));
        setLoading(false);
      })
      .catch(() => {
        setLogs([]);
        setLoading(false);
      });
  }, [token]);

  const { ws } = useWebSocket();

  useEffect(() => {
    if (!token || !ws) return;
    ws.send(JSON.stringify({ type: 'subscribe', event: 'log_chunk', token }));
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'log_chunk' && data.token === token) {
        setLogs((prev) => [...prev, data.payload]);
      }
    };

    return () => {
      if (ws) {
        ws.send(JSON.stringify({ type: 'unsubscribe', event: 'log_chunk', token }));
      }
    };
  }, [token, ws]);

  const scrollToBottom = () => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  };

  // Scroll to bottom when logs change
  useEffect(() => {
    if (autoScroll) {
      scrollToBottom();
    }
  }, [logs, autoScroll]);

  const handleDownload = () => {
    const blob = new Blob([logs.join('\n')], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `log_${token || 'unknown'}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Paper
      elevation={3}
      sx={{
        mt: 3,
        p: 2,
        minHeight: 400,
        maxHeight: 600,
        background: '#181c24',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <Stack direction="row" alignItems="center" justifyContent="space-between" mb={2} spacing={2}>
        <Typography variant="h6">Log Viewer {token ? `(Token: ${token})` : ''}</Typography>
        <Stack direction="row" spacing={2} alignItems="center">
          <Tooltip title="Download full log">
            <span>
              <Button
                variant="outlined"
                color="primary"
                size="small"
                startIcon={<DownloadIcon />}
                onClick={handleDownload}
                disabled={logs.length === 0}
              >
                Download
              </Button>
            </span>
          </Tooltip>
          <FormControlLabel
            control={
              <Switch
                checked={autoScroll}
                onChange={(e) => setAutoScroll(e.target.checked)}
                color="primary"
              />
            }
            label="Auto-scroll"
          />
        </Stack>
      </Stack>
      <Box
        id="log-container"
        ref={containerRef}
        sx={{
          flex: 1,
          overflowY: 'auto',
          background: '#11151c',
          borderRadius: 1,
          border: '1px solid #23283a',
          p: 2,
          minHeight: 320,
          maxHeight: 500,
        }}
      >
        {logs.length === 0 ? (
          <Typography color="text.secondary" sx={{ mt: 2 }}>
            No logs available.
          </Typography>
        ) : (
          logs.map((line, idx) => (
            <Typography
              key={idx}
              sx={{ fontFamily: 'monospace', fontSize: 14, whiteSpace: 'pre-wrap' }}
            >
              {line}
            </Typography>
          ))
        )}
        {!autoScroll && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 1 }}>
            <Button
              variant="text"
              size="small"
              startIcon={<ArrowDownwardIcon />}
              onClick={scrollToBottom}
            >
              Scroll to bottom
            </Button>
          </Box>
        )}
      </Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          {logs.length} logs
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {loading ? 'Connecting...' : 'Connected'}
        </Typography>
      </Box>
    </Paper>
  );
}

export default LogViewer;
