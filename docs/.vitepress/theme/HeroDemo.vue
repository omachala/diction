<script setup>
import { ref, onMounted, onUnmounted } from 'vue';

const scope = ref(null);
const typedText = ref('');
const fullText = 'Hey, running 10 mins late. Traffic is crazy!';

let isMounted = true;
let rafId = null;

// Per-bar random parameters (matches iOS WaveformView.swift)
const BAR_COUNT = 16;
const barScales = [];
const barSpeeds = [];
const barPhases = [];
const barFreqs = [];
const barHeights = new Float32Array(BAR_COUNT);
const barOffsets = new Float32Array(BAR_COUNT);

for (let i = 0; i < BAR_COUNT; i++) {
  // Center bars animate more (iOS: 0.15 to 1.0, bell curve)
  const center = (BAR_COUNT - 1) / 2;
  const dist = Math.abs(i - center) / center;
  barScales[i] = 0.15 + 0.85 * (1 - dist * dist);
  // Random easing speed per bar (iOS: 0.08 to 0.30)
  barSpeeds[i] = 0.08 + Math.random() * 0.22;
  // Random phase offset (iOS: 0 to 2pi)
  barPhases[i] = Math.random() * Math.PI * 2;
  // Random oscillation frequency (iOS: 1.5 to 4.0 Hz)
  barFreqs[i] = 1.5 + Math.random() * 2.5;
}

// Simulated audio level -- mimics natural speech patterns
function getAudioLevel(time) {
  // Layered sine waves for organic speech-like amplitude
  const base = 0.3 + 0.2 * Math.sin(time * 1.2);
  const mid = 0.25 * Math.sin(time * 2.7 + 1.0);
  const fast = 0.15 * Math.sin(time * 5.3 + 2.0);
  const burst = 0.1 * Math.sin(time * 0.8);
  return Math.max(0, Math.min(1, base + mid + fast + burst));
}

let mode = 'idle'; // 'idle' | 'recording' | 'transcribing'
let startTime = 0;
let barEls = [];

function tick() {
  if (!isMounted) return;
  const now = performance.now() / 1000;
  const elapsed = now - startTime;

  if (mode === 'recording') {
    const level = getAudioLevel(elapsed);
    for (let i = 0; i < BAR_COUNT; i++) {
      // iOS formula: target = level * (0.4 + 0.6 * sin(time * freq + phase)) * scale
      const sin = Math.sin(elapsed * barFreqs[i] + barPhases[i]);
      const target = level * (0.4 + 0.6 * sin) * barScales[i];
      // Smooth toward target (iOS: barSpeed easing)
      barHeights[i] += (target - barHeights[i]) * barSpeeds[i];
      barOffsets[i] += (0 - barOffsets[i]) * 0.2;
    }
  } else if (mode === 'transcribing') {
    // iOS wave mode: bars collapse, sine wave travels horizontally
    const waveSpeed = 5.0;
    const waveLength = BAR_COUNT * 0.35;
    for (let i = 0; i < BAR_COUNT; i++) {
      // Collapse height to minimum (iOS: 0.08)
      barHeights[i] += (0.08 - barHeights[i]) * 0.15;
      // Traveling sine wave as vertical offset
      const phase = (i / waveLength) * Math.PI * 2;
      const targetOffset = Math.sin(phase - elapsed * waveSpeed) * 0.18;
      barOffsets[i] += (targetOffset - barOffsets[i]) * 0.2;
    }
  }

  // Apply to DOM
  if (barEls.length === BAR_COUNT) {
    const containerHeight = 42; // bar container usable height in px
    const minH = containerHeight * 0.08;
    for (let i = 0; i < BAR_COUNT; i++) {
      const h = minH + (containerHeight - minH) * Math.max(0, barHeights[i]);
      const offsetPx = barOffsets[i] * containerHeight;
      barEls[i].style.height = `${h}px`;
      barEls[i].style.transform = `translateY(${offsetPx}px)`;
    }
  }

  rafId = requestAnimationFrame(tick);
}

function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

function fadeIn(el, duration = 200) {
  el.style.transition = `opacity ${duration}ms ease`;
  el.style.opacity = '1';
  return sleep(duration);
}

function fadeOut(el, duration = 200) {
  el.style.transition = `opacity ${duration}ms ease`;
  el.style.opacity = '0';
  return sleep(duration);
}

async function runTimeline() {
  if (!scope.value || !isMounted) return;

  const barIdle = scope.value.querySelector('.bar-idle');
  const barWaveform = scope.value.querySelector('.bar-waveform');
  const typedBubble = scope.value.querySelector('.typed-bubble');
  barEls = Array.from(scope.value.querySelectorAll('.bar-waveform .bar'));

  // Reset
  typedText.value = '';
  barHeights.fill(0);
  barOffsets.fill(0);
  if (barIdle) barIdle.style.opacity = '1';
  if (barWaveform) barWaveform.style.opacity = '0';
  if (typedBubble) typedBubble.style.opacity = '0';
  barEls.forEach(bar => {
    bar.style.height = '3px';
    bar.style.transform = 'translateY(0)';
  });

  // Phase 1: Idle
  mode = 'idle';
  await sleep(1500);
  if (!isMounted) return;

  // Phase 2: Recording
  await fadeOut(barIdle);
  if (!isMounted) return;
  await fadeIn(barWaveform);
  if (!isMounted) return;
  mode = 'recording';
  startTime = performance.now() / 1000;
  await sleep(3000);
  if (!isMounted) return;

  // Phase 3: Transcribing
  mode = 'transcribing';
  startTime = performance.now() / 1000;
  await sleep(2000);
  if (!isMounted) return;

  // Phase 4: Done -- back to idle
  await fadeOut(barWaveform);
  if (!isMounted) return;
  await fadeIn(barIdle);
  if (!isMounted) return;
  mode = 'idle';

  // Phase 5: Show typed text
  await fadeIn(typedBubble, 300);
  if (!isMounted) return;
  for (let i = 0; i <= fullText.length; i++) {
    if (!isMounted) return;
    typedText.value = fullText.slice(0, i);
    await sleep(30);
  }
  await sleep(2000);
  if (!isMounted) return;
  await fadeOut(typedBubble, 300);
  if (!isMounted) return;

  // Loop
  if (isMounted) runTimeline();
}

onMounted(() => {
  isMounted = true;
  rafId = requestAnimationFrame(tick);
  setTimeout(runTimeline, 100);
});

onUnmounted(() => {
  isMounted = false;
  if (rafId) cancelAnimationFrame(rafId);
});
</script>

<template>
  <div class="hero-demo" ref="scope">
    <!-- Phone frame -->
    <div class="phone-frame">
      <div class="phone-screen">
        <div class="dynamic-island"></div>

        <!-- Chat UI -->
        <div class="chat-ui">
          <div class="chat-bubble incoming short"></div>
          <div class="chat-bubble outgoing long"></div>
          <div class="chat-bubble incoming long"></div>
          <div class="chat-bubble outgoing short"></div>
          <div class="chat-bubble incoming typed-bubble">
            <span class="typed-text">{{ typedText }}<span class="cursor" v-if="typedText.length > 0 && typedText.length < fullText.length">|</span></span>
          </div>
        </div>

        <!-- Diction bar -->
        <div class="diction-bar">
          <!-- Idle state -->
          <div class="bar-idle">
            <img src="/mic-fill.svg" alt="" class="bar-mic" />
            <span class="bar-logo">Diction</span>
          </div>

          <!-- Recording/Transcribing waveform -->
          <div class="bar-waveform">
            <div class="bar" v-for="i in 16" :key="i"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hero-demo {
  position: relative;
  width: 280px;
  height: 500px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.phone-frame {
  position: relative;
  width: 260px;
  height: 480px;
  border: 5px solid #1a1a1a;
  border-radius: 42px;
  overflow: hidden;
  background: #1a1a1a;
}

.phone-screen {
  position: relative;
  width: 100%;
  height: 100%;
  background: #f5f5f5;
  border-radius: 38px;
  overflow: hidden;
}

.dark .phone-screen {
  background: #1c1c1e;
}

.dynamic-island {
  position: absolute;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  width: 90px;
  height: 28px;
  background: #1a1a1a;
  border-radius: 16px;
  z-index: 10;
}

/* Chat UI */
.chat-ui {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 50px 14px 90px;
  height: 100%;
}

.chat-bubble {
  border-radius: 16px;
  background: #e5e5ea;
  min-height: 24px;
}

.dark .chat-bubble {
  background: #3a3a3c;
}

.chat-bubble.incoming {
  align-self: flex-start;
  border-bottom-left-radius: 4px;
}

.chat-bubble.outgoing {
  align-self: flex-end;
  background: #007AFF;
  border-bottom-right-radius: 4px;
}

.chat-bubble.short { width: 45%; flex: 0.6; }
.chat-bubble.long { width: 75%; flex: 1; }

.chat-bubble.typed-bubble {
  align-self: flex-start;
  width: 90%;
  flex: 1.2;
  display: flex;
  align-items: flex-start;
  padding: 10px 12px;
  font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', system-ui, sans-serif;
  font-size: 14px;
  font-weight: 400;
  color: #000;
  line-height: 1.4;
  text-align: left;
  opacity: 0;
}

.dark .chat-bubble.typed-bubble {
  color: #fff;
}

.cursor {
  animation: blink 0.5s step-end infinite;
  color: #007AFF;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* Diction bar */
.diction-bar {
  position: absolute;
  bottom: 12px;
  left: 12px;
  right: 12px;
  height: 54px;
  background: #007AFF;
  border-radius: 18px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bar-idle {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.bar-mic {
  width: 22px;
  height: 22px;
  filter: brightness(0) invert(1);
}

.bar-logo {
  font-family: 'BigShouldersInlineText', sans-serif;
  font-weight: 800;
  font-size: 26px;
  color: #fff;
  letter-spacing: 0.5px;
}

/* Waveform */
.bar-waveform {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  opacity: 0;
}

.bar-waveform .bar {
  width: 4px;
  height: 3px;
  background: #fff;
  border-radius: 2px;
  will-change: height, transform;
}

/* Responsive */
@media (max-width: 960px) {
  .hero-demo {
    width: 240px;
    height: 440px;
  }

  .phone-frame {
    width: 220px;
    height: 400px;
    border-width: 4px;
    border-radius: 36px;
  }

  .phone-screen {
    border-radius: 32px;
  }

  .dynamic-island {
    width: 70px;
    height: 22px;
  }

  .chat-ui {
    padding: 42px 10px 76px;
    gap: 8px;
  }

  .chat-bubble.typed-bubble {
    font-size: 12px;
    padding: 8px 10px;
  }

  .diction-bar {
    height: 46px;
    bottom: 10px;
    left: 10px;
    right: 10px;
    border-radius: 14px;
  }

  .bar-logo {
    font-size: 22px;
  }

  .bar-mic {
    width: 18px;
    height: 18px;
  }
}

@media (max-width: 640px) {
  .hero-demo {
    width: 200px;
    height: 380px;
  }

  .phone-frame {
    width: 180px;
    height: 340px;
    border-width: 3px;
    border-radius: 30px;
  }

  .phone-screen {
    border-radius: 27px;
  }

  .dynamic-island {
    width: 56px;
    height: 18px;
    top: 8px;
  }

  .chat-ui {
    padding: 34px 8px 64px;
    gap: 6px;
  }

  .chat-bubble {
    border-radius: 12px;
    min-height: 18px;
  }

  .chat-bubble.incoming {
    border-bottom-left-radius: 3px;
  }

  .chat-bubble.outgoing {
    border-bottom-right-radius: 3px;
  }

  .chat-bubble.typed-bubble {
    font-size: 10px;
    padding: 6px 8px;
    border-radius: 12px;
    border-bottom-left-radius: 3px;
  }

  .diction-bar {
    height: 38px;
    bottom: 8px;
    left: 8px;
    right: 8px;
    border-radius: 12px;
  }

  .bar-logo {
    font-size: 18px;
  }

  .bar-mic {
    width: 14px;
    height: 14px;
  }

  .bar-waveform .bar {
    width: 3px;
  }

  .bar-waveform {
    gap: 3px;
  }
}
</style>
