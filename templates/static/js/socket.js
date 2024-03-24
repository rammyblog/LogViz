document.addEventListener('alpine:init', () => {
  let mode;

  if (location.protocol != 'https:') {
    mode = 'ws';
  } else {
    mode = 'wss';
  }

  // Handle ws
  const socket = new WebSocket(mode + '://' + document.location.host + '/ws');

  socket.addEventListener('message', (event) => {
    const data = JSON.parse(event.data);
    Alpine.store('state').addToLogs(data);
  });

  socket.addEventListener('open', (event) => {
    console.log('Connected to server');
    Alpine.store('state').toggleLive();
  });

  socket.addEventListener('close', (event) => {
    console.log('Disconnected from server');
    Alpine.store('state').toggleLive();
  });
});
