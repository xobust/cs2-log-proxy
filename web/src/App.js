import React from 'react';
import LogList from "./components/LogList";
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Container } from '@mui/material';
import LogViewer from './components/LogViewer';
import { useState } from 'react';

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
      <Container maxWidth="lg">
        <LogList selectedToken={selectedToken} onSelect={setSelectedToken} />
        <LogViewer token={selectedToken} />
      </Container>
    </ThemeProvider>
  );
}

export default App;
