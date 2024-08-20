document.addEventListener('DOMContentLoaded', () => {
    const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Menlo, Monaco, "Courier New", monospace',
        theme: {
            background: '#000814'
        },
        allowTransparency: true,
        scrollback: 1000
    });

    const terminalElement = document.getElementById('terminal');
    term.open(terminalElement);

    // ターミナルのリサイズ関数
    const handleResize = () => {
        const { width, height } = terminalElement.getBoundingClientRect();

        const fontSize = 14;
        const lineHeight = fontSize * 1.2;
        const charWidth = fontSize * 0.6;

        const cols = Math.floor((width - 20) / charWidth);
        const rows = Math.floor((height - 20) / lineHeight);

        term.resize(cols, rows);
    };


    const resizeObserver = new ResizeObserver(() => {
        requestAnimationFrame(handleResize);
    });

    resizeObserver.observe(terminalElement);
    window.addEventListener('resize', () => {
        requestAnimationFrame(handleResize);
    });

    const socket = new WebSocket(`ws://${window.location.host}/terminal`);

    socket.onopen = () => {
        console.log('WebSocket connection established');
        term.write('Connected to server.\r\n');
    };

    socket.onmessage = (event) => {
        term.write(event.data);
    };

    socket.onclose = (event) => {
        if (event.wasClean) {
            console.log(`Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            console.log('Connection died');
        }
        term.write('\r\nConnection to the server closed.\r\n');
    };

    socket.onerror = (error) => {
        console.error(`WebSocket error: ${error.message}`);
        term.write(`\r\nWebSocket error: ${error.message}\r\n`);
    };

    let commandBuffer = '';
    term.onData((data) => {
        socket.send(data);
        commandBuffer += data;
        if (data === '\r') {
            if (commandBuffer.trim() === 'exit') {
                socket.send('exit\r');
                document.getElementById('message').innerText = 'Session ended. You can now close this window.';
            }
            commandBuffer = '';
        }
    });

    term.focus();

    // 初回リサイズ
    handleResize();
});