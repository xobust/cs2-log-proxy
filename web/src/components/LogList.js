import React, { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import { Box, Typography, Paper, CircularProgress, Alert } from '@mui/material';
import { useWebSocket } from '../WebSocketContext';

export default function LogList({ selectedToken, onSelect }) {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetch('/api/listlogs')
      .then((res) => {
        if (!res.ok) throw new Error('Failed to fetch logs');
        return res.json();
      })
      .then((data) => {
        setLogs(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  const { subscribe, unsubscribe } = useWebSocket();

  useEffect(() => {
    // Handler for new_log events
    const handleNewLog = (data) => {
      if (data.type === 'new_log') {
        setLogs((prev) => [...prev, data.payload]);
      }
    };
    // Subscribe on mount
    const unsub = subscribe('new_log', handleNewLog, { token: '*' });
    return () => {
      unsub();
    };
  }, [subscribe]);

  const columns = [
    { field: 'token', headerName: 'Server Instance', flex: 1, minWidth: 180 },
    { field: 'log_start_time', headerName: 'Log Start Time', flex: 1, minWidth: 180 },
    { field: 'game_map', headerName: 'Game Map', flex: 1, minWidth: 120 },
    { field: 'steam_id', headerName: 'Steam ID', flex: 1, minWidth: 120 },
    { field: 'server_addr', headerName: 'Server Address', flex: 1, minWidth: 150 },
    { field: 'last_activity', headerName: 'Last Activity', flex: 1, minWidth: 180 },
    { field: 'log_id', headerName: 'Log ID', display: 'none' },
  ];

  const rows = logs.map((log, idx) => ({
    id: log.log_id,
    log_id: log.log_id,
    log_start_time: log.log_start_time,
    game_map: log.metadata.game_map,
    token: log.server_instance_token,
    steam_id: log.metadata.steam_id,
    server_addr: log.metadata.server_addr,
    last_activity: log.last_activity,
  }));

  return (
    <Box sx={{ mt: 4 }}>
      <Typography variant="h4" sx={{ mb: 2, fontWeight: 600, color: 'primary.main' }}>
        Log List
      </Typography>
      <Paper elevation={3} sx={{ p: 2, borderRadius: 3 }}>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : error ? (
          <Alert severity="error">{error}</Alert>
        ) : logs.length === 0 ? (
          <Alert severity="info">No logs found.</Alert>
        ) : (
          <div style={{ height: 500, width: '100%' }}>
            <DataGrid
              rows={rows}
              columns={columns}
              pageSize={10}
              rowsPerPageOptions={[10, 25, 50]}
              onRowClick={(params) => onSelect && onSelect(params.row.log_id)}
              getRowClassName={(params) =>
                params.row.log_id === selectedToken ? 'Mui-selected' : ''
              }
              sx={{
                backgroundColor: 'background.paper',
                '& .MuiDataGrid-columnHeaders': { backgroundColor: '#f5f5f5', fontWeight: 700 },
                '& .MuiDataGrid-row:hover': { backgroundColor: '#f0f7ff' },
                '& .Mui-selected': {
                  background: 'linear-gradient(90deg, #2193b0 0%, #6dd5ed 100%)',
                  color: '#fff',
                },
              }}
              disableSelectionOnClick
            />
          </div>
        )}
      </Paper>
    </Box>
  );
}
