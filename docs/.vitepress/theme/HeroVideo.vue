<template>
  <div class="hero-video-wrapper">
    <div class="hero-video">
      <video
        autoplay
        muted
        loop
        playsinline
        preload="auto"
        poster="/hero-poster.jpg"
      >
        <source src="/hero.webm" type="video/webm" />
        <source src="/hero.mp4" type="video/mp4" />
      </video>
    </div>
  </div>
</template>

<style scoped>
.hero-video-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  isolation: isolate;
}

/* Giant blue glow behind the video */
.hero-video-wrapper::before {
  content: "";
  position: absolute;
  z-index: -1;
  inset: -60px;
  background:
    radial-gradient(
      ellipse at center,
      rgba(120, 120, 120, 0.3) 0%,
      rgba(120, 120, 120, 0.15) 35%,
      rgba(120, 120, 120, 0) 70%
    );
  filter: blur(30px);
  pointer-events: none;
}

.hero-video {
  position: relative;
  width: 260px;
  background: transparent;
}

.hero-video video {
  width: 100%;
  height: auto;
  display: block;
  filter: drop-shadow(0 0 40px rgba(120, 120, 120, 0.35));
}

:root.dark .hero-video-wrapper::before {
  background:
    radial-gradient(
      ellipse at center,
      rgba(200, 200, 200, 0.25) 0%,
      rgba(200, 200, 200, 0.12) 35%,
      rgba(200, 200, 200, 0) 70%
    );
}

:root.dark .hero-video video {
  filter: drop-shadow(0 0 40px rgba(200, 200, 200, 0.3));
}

@media (max-width: 1279px) {
  .hero-video {
    width: 230px;
  }
}

@media (max-width: 767px) {
  .hero-video {
    width: 200px;
  }
  .hero-video-wrapper::before {
    inset: -40px;
  }
}

@media (max-width: 479px) {
  .hero-video {
    width: 170px;
  }
}
</style>
