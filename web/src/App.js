import React from 'react';
import LogList from "./components/LogList";
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Container, Box } from '@mui/material';
import LogViewer from './components/LogViewer';
import { useState } from 'react';
import { WebSocketProvider, useWebSocket } from './WebSocketContext';
import ConnectionIndicator from './components/ConnectionIndicator';

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#1976d2',
    },
    background: {
      default: '#121212',
      paper: '#1e1e1e',
    },
  },
});

function App() {
  const [selectedToken, setSelectedToken] = useState(null);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <WebSocketProvider>
        <Container maxWidth="lg" style={{ position: 'relative' }}>
          <Box sx={{ position: 'absolute', top: 16, right: 24, zIndex: 10 }}>
            <WebSocketStatusDisplay />
          </Box>
          <LogList selectedToken={selectedToken} onSelect={setSelectedToken} />
          <LogViewer token={selectedToken} />
        </Container>
      </WebSocketProvider>
    </ThemeProvider>
  );
}

function WebSocketStatusDisplay() {
  const { status } = useWebSocket() || { status: 'disconnected' };
  return <ConnectionIndicator status={status} />;
}

export default App;
