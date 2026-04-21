<script setup lang="ts">
import { data as reviews } from './reviews.data.js'
</script>

<template>
  <section v-if="reviews.length > 0" class="testimonials">
    <div class="testimonials-inner">
      <p class="testimonials-label">From the App Store</p>
      <h2 class="testimonials-heading">Don't take our word for it.</h2>
      <div class="testimonials-grid">
        <div v-for="review in reviews" :key="review.author" class="testimonial-card">
          <div class="stars">&#9733;&#9733;&#9733;&#9733;&#9733;</div>
          <p class="review-body">{{ review.body }}</p>
          <div class="review-footer">
            <span class="review-title">{{ review.title }}</span>
            <span class="review-author">{{ review.author }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.testimonials {
  padding: 5rem 1.5rem;
  background: var(--vp-c-bg-soft);
}

.testimonials-inner {
  max-width: 1152px;
  margin: 0 auto;
  text-align: center;
}

.testimonials-label {
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--vp-c-text-3);
  margin: 0 0 0.75rem;
}

.testimonials-heading {
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 700;
  color: var(--vp-c-text-1);
  margin: 0 0 3rem;
  border: none;
  letter-spacing: -0.02em;
}

.testimonials-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.25rem;
  text-align: left;
}

.testimonial-card {
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

.testimonial-card:hover {
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.dark .testimonial-card:hover {
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.stars {
  color: #FF9F0A;
  font-size: 1rem;
  letter-spacing: 0.1em;
}

.review-body {
  font-size: 0.95rem;
  color: var(--vp-c-text-2);
  line-height: 1.65;
  margin: 0;
  flex: 1;
}

.review-footer {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--vp-c-divider);
}

.review-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
}

.review-author {
  font-size: 0.8rem;
  color: var(--vp-c-text-3);
}

@media (max-width: 640px) {
  .testimonials {
    padding: 3.5rem 1.25rem;
  }
}
</style>
