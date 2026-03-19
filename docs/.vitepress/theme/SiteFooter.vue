<template>
  <footer class="site-footer">
    <div class="footer-container">
      <div class="footer-columns">
        <div class="footer-col footer-brand">
          <a href="/" class="brand-link">
            <img src="/app-icon.png" alt="Diction" class="brand-icon" width="40" height="40" />
            <span class="brand-name">Diction</span>
          </a>
          <p class="brand-tagline">Voice keyboard for iPhone.</p>
        </div>

        <div class="footer-col">
          <h4>Product</h4>
          <ul>
            <li><a href="/features/">Features</a></li>
            <li><a href="/on-device">On-Device</a></li>
            <li><a href="/self-hosted">Self-Hosted</a></li>
            <li><a href="/cloud">Diction One</a></li>
            <li><a href="/encryption">Encryption</a></li>
          </ul>
        </div>

        <div class="footer-col">
          <h4>Resources</h4>
          <ul>
            <li><a href="/support">Support</a></li>
            <li><a href="https://github.com/omachala/diction" target="_blank" rel="noopener">GitHub</a></li>
          </ul>
        </div>

        <div class="footer-col">
          <h4>Legal</h4>
          <ul>
            <li><a href="/privacy">Privacy Policy</a></li>
            <li><a href="/terms">Terms of Service</a></li>
          </ul>
        </div>
      </div>

      <div class="footer-bottom">
        <span>&copy; {{ year }} Ondrej Machala</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
const year = new Date().getFullYear();
</script>

<style>
/* Footer tokens -- unscoped so :root/.dark work */
:root {
  --footer-bg: #f6f8fa;
  --footer-border: var(--vp-c-divider);
  --footer-heading: #1a1a1a;
  --footer-link: #4a4a4a;
  --footer-link-hover: var(--diction-blue);
  --footer-divider: #d1d5db;
  --footer-muted: #6e6e6e;
}

.dark {
  --footer-bg: #141416;
  --footer-border: rgba(255, 255, 255, 0.06);
  --footer-heading: #e8e8e8;
  --footer-link: #a0a0a0;
  --footer-link-hover: #4DA3FF;
  --footer-divider: rgba(255, 255, 255, 0.08);
  --footer-muted: #707070;
}
</style>

<style scoped>
.site-footer {
  border-top: 1px solid var(--footer-border);
  background: var(--footer-bg);
  padding: 48px 24px 32px;
}

.footer-container {
  max-width: 1152px;
  margin: 0 auto;
}

.footer-columns {
  display: grid;
  grid-template-columns: 1.4fr 1fr 1fr 1fr;
  gap: 40px;
}

/* Brand column */
.brand-link {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.brand-icon {
  border-radius: 22.5%;
}

.brand-name {
  font-family: 'BigShouldersInlineText', sans-serif;
  font-weight: 800;
  font-size: 24px;
  color: var(--footer-heading);
  letter-spacing: 0.5px;
}

.brand-tagline {
  margin: 12px 0 0;
  font-size: 14px;
  color: var(--footer-muted);
  line-height: 1.5;
}

/* Column headings */
.footer-col h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--footer-heading);
  margin: 0 0 16px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Links */
.footer-col ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.footer-col li {
  margin: 0 0 10px;
}

.footer-col a {
  font-size: 14px;
  color: var(--footer-link);
  text-decoration: none;
  transition: color 0.15s;
}

.footer-col a:hover {
  color: var(--footer-link-hover);
}

/* Bottom bar */
.footer-bottom {
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid var(--footer-divider);
  font-size: 13px;
  color: var(--footer-muted);
}

@media (max-width: 768px) {
  .footer-columns {
    grid-template-columns: 1fr 1fr;
    gap: 32px;
  }
}

@media (max-width: 480px) {
  .footer-columns {
    grid-template-columns: 1fr;
    gap: 28px;
  }

  .site-footer {
    padding: 32px 20px 24px;
  }
}
</style>
