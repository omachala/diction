<script setup>
import { useData } from 'vitepress';
import { computed } from 'vue';

const { isDark } = useData();
const suffix = computed(() => (isDark.value ? 'dark' : 'light'));
</script>

<template>
  <div class="hero-screenshots">
    <img :src="`/slides/v1/raw/keyboard-recording-${suffix}.png`" alt="Diction keyboard recording" class="slide slide-left" />
    <img :src="`/slides/v1/raw/on-device-${suffix}.png`" alt="On-device transcription" class="slide slide-center" />
    <img :src="`/slides/v1/raw/self-hosted-${suffix}.png`" alt="Self-hosted transcription" class="slide slide-right" />
  </div>
</template>

<style scoped>
.hero-screenshots {
  position: relative;
  width: 420px;
  height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.slide {
  position: absolute;
  height: 400px;
  width: auto;
  border-radius: 16px;
  border: 2px solid #1a1a1a;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.slide-left {
  transform: translateX(-120px) rotate(-4deg);
  z-index: 1;
}

.slide-center {
  transform: translateX(0);
  z-index: 2;
}

.slide-right {
  transform: translateX(120px) rotate(4deg);
  z-index: 3;
}

@media (max-width: 960px) {
  .hero-screenshots {
    width: 340px;
    height: 400px;
    margin: 0 auto;
  }
  .slide {
    height: 340px;
  }
  .slide-left {
    transform: translateX(-90px) rotate(-3deg);
  }
  .slide-right {
    transform: translateX(90px) rotate(3deg);
  }
}

@media (max-width: 640px) {
  .hero-screenshots {
    width: 280px;
    height: 340px;
  }
  .slide {
    height: 280px;
  }
  .slide-left {
    transform: translateX(-70px) rotate(-3deg);
  }
  .slide-right {
    transform: translateX(70px) rotate(3deg);
  }
}
</style>
