class TranscriptionApp {
    constructor() {
        this.ws = null;
        this.peer = null;
        this.stream = null;
        this.sessionId = this.generateId();
        this.isRecording = false;

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
            console.error('mic access failed:', err);
            this.setStatus('Mic denied', 'red');
        }
    }

    connectWebSocket() {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(protocol + '//' + location.host + '/ws');

        this.ws.onopen = () => {
            this.ws.send(JSON.stringify({
                type: 'join',
                sessionId: this.sessionId
            }));
        };

        this.ws.onmessage = (e) => {
            const msg = JSON.parse(e.data);
            this.handleMessage(msg);
        };

        this.ws.onclose = () => {
            this.setStatus('Disconnected', 'gray');
            this.cleanup();
        };

        this.ws.onerror = (err) => {
            console.error('ws error:', err);
            this.setStatus('Error', 'red');
        };
    }

    handleMessage(msg) {
        switch (msg.type) {
            case 'ready':
                this.setupPeer();
                break;

            case 'answer':
                if (this.peer) {
                    this.peer.signal({ type: 'answer', sdp: msg.sdp });
                }
                break;

            case 'candidate':
                if (this.peer && msg.candidate) {
                    this.peer.signal({ candidate: msg.candidate });
                }
                break;

            case 'transcription':
                this.addTranscription(msg.text);
                break;

            case 'error':
                console.error('server error:', msg.message);
                this.setStatus('Error: ' + msg.code, 'red');
                break;
        }
    }

    setupPeer() {
        this.peer = new SimplePeer({
            initiator: true,
            stream: this.stream,
            config: {
                iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
            },
            offerOptions: {
                offerToReceiveAudio: false,
                offerToReceiveVideo: false
            }
        });

        this.peer.on('signal', (data) => {
            if (data.type === 'offer') {
                this.ws.send(JSON.stringify({
                    type: 'offer',
                    sessionId: this.sessionId,
                    sdp: data.sdp
                }));
            } else if (data.candidate) {
                this.ws.send(JSON.stringify({
                    type: 'candidate',
                    sessionId: this.sessionId,
                    candidate: data.candidate
                }));
            }
        });

        this.peer.on('connect', () => {
            this.setStatus('Recording', 'green');
            this.isRecording = true;
            this.startBtn.disabled = true;
            this.stopBtn.disabled = false;
        });

        this.peer.on('error', (err) => {
            console.error('peer error:', err);
            this.setStatus('Connection failed', 'red');
        });

        this.peer.on('close', () => {
            this.cleanup();
        });
    }

    addTranscription(text) {
        if (!text || !text.trim()) return;

        const segment = document.createElement('span');
        segment.className = 'new-segment';
        segment.textContent = text + ' ';
        this.transcriptionEl.appendChild(segment);
        this.transcriptionEl.scrollTop = this.transcriptionEl.scrollHeight;
    }

    stop() {
        this.cleanup();
        this.setStatus('Stopped', 'gray');
    }

    cleanup() {
        this.isRecording = false;
        this.startBtn.disabled = false;
        this.stopBtn.disabled = true;

        if (this.peer) {
            this.peer.destroy();
            this.peer = null;
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
