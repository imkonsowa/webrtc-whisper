const { Room, RoomEvent, Track } = LivekitClient;

class TranscriptionApp {
    constructor() {
        this.room = null;
        this.currentRoom = null;

        this.joinBtn = document.getElementById('joinBtn');
        this.leaveBtn = document.getElementById('leaveBtn');
        this.clearBtn = document.getElementById('clearBtn');
        this.roomInput = document.getElementById('roomInput');
        this.nameInput = document.getElementById('nameInput');
        this.statusDot = document.getElementById('statusDot');
        this.statusText = document.getElementById('statusText');
        this.roomNameEl = document.getElementById('roomName');
        this.participantsEl = document.getElementById('participants');
        this.transcriptionEl = document.getElementById('transcription');

        this.joinBtn.onclick = () => this.join();
        this.leaveBtn.onclick = () => this.leave();
        this.clearBtn.onclick = () => this.clear();
    }

    setStatus(status, color) {
        this.statusText.textContent = status;
        this.statusDot.className = 'w-3 h-3 rounded-full bg-' + color + '-400';
    }

    async join() {
        const roomName = this.roomInput.value.trim() || undefined;
        const participantName = this.nameInput.value.trim() || 'User';

        this.setStatus('Connecting...', 'yellow');

        try {
            const res = await fetch('/api/rooms/' + (roomName || 'default') + '/join', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ participantName })
            });

            if (!res.ok) {
                throw new Error('Failed to get token');
            }

            const { token, livekitUrl } = await res.json();

            this.room = new Room({
                adaptiveStream: true,
                dynacast: true,
            });

            this.setupRoomEvents();

            await this.room.connect(livekitUrl, token);

            await this.room.localParticipant.setMicrophoneEnabled(true);

            this.currentRoom = roomName || 'default';
            this.roomNameEl.textContent = this.currentRoom;
            this.setStatus('Connected', 'green');
            this.joinBtn.disabled = true;
            this.leaveBtn.disabled = false;
            this.roomInput.disabled = true;
            this.nameInput.disabled = true;

            this.updateParticipants();
        } catch (err) {
            console.error('join error:', err);
            this.setStatus('Connection failed', 'red');
        }
    }

    setupRoomEvents() {
        this.room.on(RoomEvent.ParticipantConnected, () => {
            this.updateParticipants();
        });

        this.room.on(RoomEvent.ParticipantDisconnected, () => {
            this.updateParticipants();
        });

        this.room.on(RoomEvent.DataReceived, (payload, participant) => {
            try {
                const msg = JSON.parse(new TextDecoder().decode(payload));
                if (msg.type === 'transcription') {
                    this.addTranscription(msg.participantName, msg.text);
                }
            } catch (e) {
                console.error('data parse error:', e);
            }
        });

        this.room.on(RoomEvent.Disconnected, () => {
            this.handleDisconnect();
        });

        this.room.on(RoomEvent.ActiveSpeakersChanged, (speakers) => {
            this.updateSpeakingState(speakers);
        });
    }

    updateParticipants() {
        if (!this.room) {
            this.participantsEl.innerHTML = '<p class="text-xs text-gray-400">No participants</p>';
            return;
        }

        const participants = [this.room.localParticipant, ...this.room.remoteParticipants.values()]
            .filter(p => p.identity !== 'transcription-bot');

        if (participants.length === 0) {
            this.participantsEl.innerHTML = '<p class="text-xs text-gray-400">No participants</p>';
            return;
        }

        this.participantsEl.innerHTML = participants.map(p => {
            const isLocal = p === this.room.localParticipant;
            return `
                <div class="participant flex items-center gap-2 p-2 rounded border" data-sid="${p.sid}">
                    <span class="w-2 h-2 rounded-full bg-green-400"></span>
                    <span class="text-sm">${p.name || p.identity}${isLocal ? ' (you)' : ''}</span>
                </div>
            `;
        }).join('');
    }

    updateSpeakingState(speakers) {
        document.querySelectorAll('.participant').forEach(el => {
            el.classList.remove('speaking');
        });

        speakers.forEach(speaker => {
            const el = document.querySelector(`.participant[data-sid="${speaker.sid}"]`);
            if (el) {
                el.classList.add('speaking');
            }
        });
    }

    addTranscription(speakerName, text) {
        if (!text || !text.trim()) return;

        const div = document.createElement('div');
        div.className = 'new-segment mb-2';
        div.innerHTML = `<span class="text-xs text-gray-500 ml-2">[${speakerName}]</span> ${text}`;
        this.transcriptionEl.appendChild(div);
        this.transcriptionEl.scrollTop = this.transcriptionEl.scrollHeight;
    }

    leave() {
        if (this.room) {
            this.room.disconnect();
        }
        this.handleDisconnect();
    }

    handleDisconnect() {
        this.room = null;
        this.currentRoom = null;
        this.roomNameEl.textContent = '-';
        this.setStatus('Disconnected', 'gray');
        this.joinBtn.disabled = false;
        this.leaveBtn.disabled = true;
        this.roomInput.disabled = false;
        this.nameInput.disabled = false;
        this.updateParticipants();
    }

    clear() {
        this.transcriptionEl.innerHTML = '';
    }
}

const app = new TranscriptionApp();
