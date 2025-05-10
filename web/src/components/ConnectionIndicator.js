import React from 'react';
import CircleIcon from '@mui/icons-material/Circle';

const statusColors = {
  connected: '#4caf50', // green
  connecting: '#ff9800', // orange
  disconnected: '#f44336', // red
  error: '#e53935', // dark red
};

export default function ConnectionIndicator({ status }) {
  const color = statusColors[status] || '#757575';
  const label = status.charAt(0).toUpperCase() + status.slice(1);
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14 }}>
      <CircleIcon style={{ color, fontSize: 14 }} />
      <span>{label}</span>
    </div>
  );
}
