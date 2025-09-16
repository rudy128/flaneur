import { useState, useEffect } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'


function App() {
  const [health, setHealth] = useState<'loading' | 'ok' | 'error'>('loading');
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

    useEffect(() => {

      let ws: WebSocket | null = null;
      let reconnectTimeout: number | null = null;
      let manuallyClosed = false;

      const connect = () => {
        ws = new WebSocket('ws://localhost:8080/health');

        ws.onopen = () => {
          setHealth('loading');
          setErrorMsg(null);
        };
        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.status === 'ok') {
              setHealth('ok');
              setErrorMsg(null);
            } else {
              setHealth('error');
              setErrorMsg('Unexpected health status');
            }
          } catch {
            setHealth('error');
            setErrorMsg('Invalid response from server');
          }
        };
        ws.onerror = () => {
          setHealth('error');
          setErrorMsg('WebSocket error');
        };
        ws.onclose = () => {
          if (!manuallyClosed) {
            setHealth('error');
            setErrorMsg('Connection closed. Reconnecting...');
            reconnectTimeout = window.setTimeout(connect, 2000);
          }
        };
      };

      connect();

      return () => {
        manuallyClosed = true;
        if (ws) ws.close();
        if (reconnectTimeout) clearTimeout(reconnectTimeout);
      };
    }, []);

  return (
    <>
      <div>
        <a href="https://vite.dev" target="_blank">
          <img src={viteLogo} className="logo" alt="Vite logo" />
        </a>
        <a href="https://react.dev" target="_blank">
          <img src={reactLogo} className="logo react" alt="React logo" />
        </a>
      </div>
      <h1>Vite + React</h1>
      <div className="card">
        <h2>API Health Status:</h2>
        {health === 'loading' && <span style={{ color: 'gray' }}>Checking...</span>}
        {health === 'ok' && <span style={{ color: 'green' }}>Healthy</span>}
        {health === 'error' && <span style={{ color: 'red' }}>Unhealthy{errorMsg ? `: ${errorMsg}` : ''}</span>}
      </div>
      <p className="read-the-docs">
        Click on the Vite and React logos to learn more
      </p>
    </>
  )
}

export default App
