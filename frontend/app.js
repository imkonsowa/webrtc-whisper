class TranscriptionApp {
    constructor() {
        this.ws = null;
        this.audioContext = null;
        this.workletNode = null;
        this.stream = null;
        this.sessionId = this.generateId();

        this.startBtn = document.getElementById('startBtn');
        this.stopBtn = document.getElementById('stopBtn');
        this.clearBtn = document.getElementById('clearBtn');
        this.statusDot = document.getElementById('statusDot');
        this.statusText = document.getElementById('statusText');
        this.sessionIdEl = document.getElementById('sessionId');
        this.transcriptionEl = document.getElementById('transcription');

        this.sessionIdEl.textContent = this.sessionId;

        this.startBtn.onclick = () => this.start();
        this.stopBtn.onclick = () => this.stop();
        this.clearBtn.onclick = () => this.clear();
    }

    generateId() {
        return 'sess_' + Math.random().toString(36).substr(2, 9);
    }

    setStatus(status, color) {
        this.statusText.textContent = status;
        this.statusDot.className = 'w-3 h-3 rounded-full bg-' + color + '-400';
    }

    async start() {
        try {
            this.stream = await navigator.mediaDevices.getUserMedia({
                audio: {
                    echoCancellation: true,
                    noiseSuppression: true,
                    sampleRate: 48000
                },
                video: false
            });

            this.setStatus('Connecting...', 'yellow');
            this.connectWebSocket();
        } catch (err) {
            console.error('mic error:', err);
            this.setStatus('Mic denied', 'red');
        }
    }

    connectWebSocket() {
        const url = (location.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + location.host + '/ws';
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
            this.ws.send(JSON.stringify({ type: 'join', sessionId: this.sessionId }));
        };

        this.ws.onmessage = (e) => {
            const msg = JSON.parse(e.data);
            if (msg.type === 'ready') {
                this.startAudioCapture();
            } else if (msg.type === 'transcription') {
                this.addTranscription(msg.text);
            } else if (msg.type === 'error') {
                console.error('server:', msg.message);
                this.setStatus('Error', 'red');
            }
        };

        this.ws.onclose = () => {
            this.setStatus('Disconnected', 'gray');
            this.cleanup();
        };

        this.ws.onerror = () => this.setStatus('Connection error', 'red');
    }

    async startAudioCapture() {
        try {
            this.audioContext = new AudioContext({ sampleRate: 48000 });
            await this.audioContext.audioWorklet.addModule('audio-processor.js');

            const source = this.audioContext.createMediaStreamSource(this.stream);
            this.workletNode = new AudioWorkletNode(this.audioContext, 'audio-processor');

            this.workletNode.port.onmessage = (e) => {
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(e.data);
                }
            };

            source.connect(this.workletNode);
            this.workletNode.connect(this.audioContext.destination);

            this.setStatus('Recording', 'green');
            this.startBtn.disabled = true;
            this.stopBtn.disabled = false;
        } catch (err) {
            console.error('audio capture error:', err);
            this.setStatus('Audio error', 'red');
        }
    }

    addTranscription(text) {
        if (!text || !text.trim()) return;
        const span = document.createElement('span');
        span.className = 'new-segment';
        span.textContent = text + ' ';
        this.transcriptionEl.appendChild(span);
        this.transcriptionEl.scrollTop = this.transcriptionEl.scrollHeight;
    }

    stop() {
        this.cleanup();
        this.setStatus('Stopped', 'gray');
    }

    cleanup() {
        this.startBtn.disabled = false;
        this.stopBtn.disabled = true;

        if (this.workletNode) {
            this.workletNode.disconnect();
            this.workletNode = null;
        }

        if (this.audioContext) {
            this.audioContext.close();
            this.audioContext = null;
        }

        if (this.stream) {
            this.stream.getTracks().forEach(t => t.stop());
            this.stream = null;
        }

        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.close();
        }
        this.ws = null;

        this.sessionId = this.generateId();
        this.sessionIdEl.textContent = this.sessionId;
    }

    clear() {
        this.transcriptionEl.innerHTML = '';
    }
}

const app = new TranscriptionApp();
