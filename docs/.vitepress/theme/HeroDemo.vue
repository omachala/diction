<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue';

const scope = ref(null);
const bubbles = reactive([]);
let bubbleId = 0;
let isMounted = true;
let rafId = null;

// Dot waveform parameters (matches iOS WaveformView.swift frequencies)
const DOT_COUNT = 16;
const dotScales = [];
const dotSpeeds = [];
const dotPhases = [];
const dotFreqs = [];
const dotHeights = new Float32Array(DOT_COUNT);
const dotOffsets = new Float32Array(DOT_COUNT);

for (let i = 0; i < DOT_COUNT; i++) {
  const center = (DOT_COUNT - 1) / 2;
  const dist = Math.abs(i - center) / center;
  dotScales[i] = 0.15 + 0.85 * (1 - dist * dist);
  dotSpeeds[i] = 0.08 + Math.random() * 0.22;
  dotPhases[i] = Math.random() * Math.PI * 2;
  dotFreqs[i] = 1.5 + Math.random() * 2.5;
}

function getAudioLevel(time) {
  const base = 0.3 + 0.2 * Math.sin(time * 1.2);
  const mid = 0.25 * Math.sin(time * 2.7 + 1.0);
  const fast = 0.15 * Math.sin(time * 5.3 + 2.0);
  const burst = 0.1 * Math.sin(time * 0.8);
  return Math.max(0, Math.min(1, base + mid + fast + burst));
}

let mode = 'idle';
let startTime = 0;
let dotEls = [];

function tick() {
  if (!isMounted) return;
  const now = performance.now() / 1000;
  const elapsed = now - startTime;

  if (mode === 'recording') {
    const level = getAudioLevel(elapsed);
    for (let i = 0; i < DOT_COUNT; i++) {
      const sin = Math.sin(elapsed * dotFreqs[i] + dotPhases[i]);
      const target = level * (0.4 + 0.6 * sin) * dotScales[i];
      dotHeights[i] += (target - dotHeights[i]) * dotSpeeds[i];
      dotOffsets[i] += (0 - dotOffsets[i]) * 0.2;
    }
  } else if (mode === 'transcribing') {
    const waveSpeed = 5.0;
    const waveLength = DOT_COUNT * 0.35;
    for (let i = 0; i < DOT_COUNT; i++) {
      dotHeights[i] += (0.08 - dotHeights[i]) * 0.15;
      const phase = (i / waveLength) * Math.PI * 2;
      const targetOffset = Math.sin(phase - elapsed * waveSpeed) * 0.18;
      dotOffsets[i] += (targetOffset - dotOffsets[i]) * 0.2;
    }
  } else {
    for (let i = 0; i < DOT_COUNT; i++) {
      dotHeights[i] += (0 - dotHeights[i]) * 0.1;
      dotOffsets[i] += (0 - dotOffsets[i]) * 0.1;
    }
  }

  // Apply height + offset to DOM (bar animation)
  if (dotEls.length === DOT_COUNT) {
    const containerH = 38;
    const minH = containerH * 0.08;
    for (let i = 0; i < DOT_COUNT; i++) {
      const h = minH + (containerH - minH) * Math.max(0, dotHeights[i]);
      const offsetPx = dotOffsets[i] * containerH;
      dotEls[i].style.height = `${h}px`;
      dotEls[i].style.transform = `translateY(${offsetPx}px)`;
    }
  }

  rafId = requestAnimationFrame(tick);
}

function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

// Varied sizes for natural conversation look
const widths = ['45%', '65%', '50%', '75%', '40%', '60%', '55%', '70%'];
const heights = [78, 90, 74, 88, 82, 86, 76, 84];
let sizeIdx = 0;

function addBubble(type) {
  const w = widths[sizeIdx % widths.length];
  const h = heights[sizeIdx % heights.length];
  sizeIdx++;
  bubbles.push({ id: bubbleId++, type, width: w, height: h + 'px' });
  while (bubbles.length > 12) bubbles.shift();
}

// Seed bubbles — tall enough that ~6 fill the screen
bubbles.push(
  { id: bubbleId++, type: 'incoming', width: '55%', height: '78px' },
  { id: bubbleId++, type: 'outgoing', width: '70%', height: '90px' },
  { id: bubbleId++, type: 'incoming', width: '40%', height: '74px' },
  { id: bubbleId++, type: 'outgoing', width: '60%', height: '88px' },
  { id: bubbleId++, type: 'incoming', width: '75%', height: '82px' },
  { id: bubbleId++, type: 'outgoing', width: '50%', height: '86px' },
  { id: bubbleId++, type: 'incoming', width: '45%', height: '76px' },
  { id: bubbleId++, type: 'outgoing', width: '65%', height: '84px' },
  { id: bubbleId++, type: 'incoming', width: '55%', height: '78px' },
  { id: bubbleId++, type: 'outgoing', width: '70%', height: '80px' },
);
sizeIdx = 10;

async function runTimeline() {
  if (!scope.value || !isMounted) return;

  const barIdle = scope.value.querySelector('.bar-idle');
  const barWaveform = scope.value.querySelector('.bar-waveform');
  dotEls = Array.from(scope.value.querySelectorAll('.bar-waveform .dot'));

  function fadeIn(el, duration = 200) {
    if (!el) return sleep(0);
    el.style.transition = `opacity ${duration}ms ease`;
    el.style.opacity = '1';
    return sleep(duration);
  }
  function fadeOut(el, duration = 200) {
    if (!el) return sleep(0);
    el.style.transition = `opacity ${duration}ms ease`;
    el.style.opacity = '0';
    return sleep(duration);
  }

  const dictionBar = scope.value.querySelector('.diction-bar');

  async function pressButton() {
    if (!dictionBar) return;
    dictionBar.style.transition = 'transform 0.1s ease-in';
    dictionBar.style.transform = 'scale(0.93)';
    await sleep(100);
    dictionBar.style.transition = 'transform 0.15s ease-out';
    dictionBar.style.transform = 'scale(1)';
    await sleep(150);
  }

  dotHeights.fill(0);
  dotOffsets.fill(0);
  if (barIdle) barIdle.style.opacity = '1';
  if (barWaveform) barWaveform.style.opacity = '0';
  dotEls.forEach(d => {
    d.style.height = '3px';
    d.style.transform = 'translateY(0)';
  });

  await sleep(1200);
  if (!isMounted) return;

  while (isMounted) {
    // Tap the button
    await pressButton();
    if (!isMounted) return;

    // Recording
    await fadeOut(barIdle);
    if (!isMounted) return;
    await fadeIn(barWaveform);
    if (!isMounted) return;
    mode = 'recording';
    startTime = performance.now() / 1000;
    await sleep(2500);
    if (!isMounted) return;

    // Transcribing (quick — Diction is fast)
    mode = 'transcribing';
    startTime = performance.now() / 1000;
    await sleep(600);
    if (!isMounted) return;

    // Back to idle
    await fadeOut(barWaveform);
    if (!isMounted) return;
    await fadeIn(barIdle);
    if (!isMounted) return;
    mode = 'idle';

    // New outgoing (blue) bubble — user just dictated this
    await sleep(300);
    if (!isMounted) return;
    addBubble('outgoing');
    await nextTick();

    // Incoming (gray) reply arrives
    await sleep(1200);
    if (!isMounted) return;
    addBubble('incoming');
    await nextTick();

    await sleep(1500);
    if (!isMounted) return;
  }
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
    <div class="phone-frame">
      <div class="phone-screen">
        <div class="dynamic-island"></div>

        <!-- Chat conversation -->
        <TransitionGroup name="bubble" tag="div" class="chat-ui">
          <div
            v-for="b in bubbles"
            :key="b.id"
            class="chat-bubble"
            :class="b.type"
            :style="{ width: b.width, height: b.height }"
          ></div>
        </TransitionGroup>

        <!-- Diction bar -->
        <div class="diction-bar">
          <div class="bar-idle">
            <img src="/mic-fill.svg" alt="" class="bar-mic" />
            <span class="bar-logo">Diction</span>
          </div>
          <div class="bar-waveform">
            <div class="dot" v-for="i in 16" :key="i"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hero-demo {
  position: relative;
  width: 320px;
  height: 664px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.phone-frame {
  position: relative;
  width: 100%;
  height: 100%;
  border: 5px solid #1a1a1a;
  border-radius: 44px;
  overflow: hidden;
  background: #1a1a1a;
}

.phone-screen {
  position: relative;
  width: 100%;
  height: 100%;
  background: #f5f5f5;
  border-radius: 40px;
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

/* Chat UI — spacer pushes bubbles to bottom */
.chat-ui {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 8px;
  padding: 14px;
  padding-bottom: 80px;
  height: 100%;
  overflow: hidden;
  position: relative;
  mask-image: linear-gradient(to bottom, transparent 6%, black 14%);
  -webkit-mask-image: linear-gradient(to bottom, transparent 6%, black 14%);
}

.chat-bubble {
  border-radius: 16px;
  flex-shrink: 0;
  background: #e5e5ea;
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

/* Bubble enter — slide up from below like iMessage */
.bubble-enter-active {
  transition: opacity 0.35s ease-out, transform 0.35s cubic-bezier(0.2, 0.8, 0.3, 1);
}

.bubble-enter-from {
  opacity: 0;
  transform: translateY(40px);
}

/* Existing bubbles shift up smoothly */
.bubble-move {
  transition: transform 0.35s cubic-bezier(0.2, 0.8, 0.3, 1);
}

/* Removed bubbles vanish instantly (clipped off-screen at top) */
.bubble-leave-active {
  position: absolute;
  opacity: 0;
  transition: none;
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

/* Dot waveform — actual circles, not bars */
.bar-waveform {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  opacity: 0;
}

.bar-waveform .dot {
  width: 3px;
  height: 3px;
  background: #fff;
  border-radius: 1.5px;
  will-change: height, transform;
}

/* Responsive */
@media (max-width: 960px) {
  .hero-demo {
    width: 190px;
    height: 400px;
  }

  .phone-frame {
    width: 174px;
    height: 378px;
    border-width: 4px;
    border-radius: 32px;
  }

  .phone-screen {
    border-radius: 28px;
  }

  .dynamic-island {
    width: 70px;
    height: 22px;
  }

  .chat-ui {
    padding: 10px;
    padding-bottom: 60px;
    gap: 6px;
  }

  .chat-bubble {
    border-radius: 12px;
  }

  .chat-bubble.incoming { border-bottom-left-radius: 3px; }
  .chat-bubble.outgoing { border-bottom-right-radius: 3px; }

  .diction-bar {
    height: 42px;
    bottom: 10px;
    left: 10px;
    right: 10px;
    border-radius: 14px;
  }

  .bar-logo {
    font-size: 20px;
  }

  .bar-mic {
    width: 16px;
    height: 16px;
  }
}

@media (max-width: 640px) {
  .hero-demo {
    width: 160px;
    height: 340px;
  }

  .phone-frame {
    width: 146px;
    height: 318px;
    border-width: 3px;
    border-radius: 26px;
  }

  .phone-screen {
    border-radius: 23px;
  }

  .dynamic-island {
    width: 50px;
    height: 16px;
    top: 8px;
  }

  .chat-ui {
    padding: 8px;
    padding-bottom: 48px;
    gap: 5px;
  }

  .chat-bubble {
    border-radius: 10px;
  }

  .chat-bubble.incoming { border-bottom-left-radius: 3px; }
  .chat-bubble.outgoing { border-bottom-right-radius: 3px; }

  .diction-bar {
    height: 34px;
    bottom: 8px;
    left: 8px;
    right: 8px;
    border-radius: 10px;
  }

  .bar-logo {
    font-size: 16px;
  }

  .bar-mic {
    width: 12px;
    height: 12px;
  }

  .bar-waveform .dot {
    width: 2px;
    height: 2px;
    border-radius: 1px;
  }

  .bar-waveform {
    gap: 3px;
  }
}
</style>
