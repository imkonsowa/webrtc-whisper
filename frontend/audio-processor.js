class AudioProcessor extends AudioWorkletProcessor {
    constructor() {
        super();
        this.bufferSize = 4800;
        this.buffer = new Float32Array(this.bufferSize);
        this.bufferIndex = 0;
    }

    process(inputs, outputs, parameters) {
        const input = inputs[0];
        if (!input || !input[0]) return true;

        const samples = input[0];

        for (let i = 0; i < samples.length; i++) {
            this.buffer[this.bufferIndex++] = samples[i];

            if (this.bufferIndex >= this.bufferSize) {
                const pcm = new Int16Array(this.bufferSize);
                for (let j = 0; j < this.bufferSize; j++) {
                    pcm[j] = Math.max(-32768, Math.min(32767, Math.floor(this.buffer[j] * 32767)));
                }
                this.port.postMessage(pcm.buffer, [pcm.buffer]);
                this.buffer = new Float32Array(this.bufferSize);
                this.bufferIndex = 0;
            }
        }

        return true;
    }
}

registerProcessor('audio-processor', AudioProcessor);
