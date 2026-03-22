import { defineConfig } from 'vitepress';
import llmstxt from 'vitepress-plugin-llms';

const SITE_URL = 'https://diction.one';
const SITE_NAME = 'Diction';
const SITE_TITLE = 'Diction — Voice Keyboard for iPhone';
const DEFAULT_DESCRIPTION =
  'Voice keyboard for iPhone with open-source server. Tap the mic, speak, text appears. On-device, self-hosted, or cloud. Zero tracking, no word limits, 99 languages.';
const DEFAULT_KEYWORDS =
  'voice keyboard, speech to text, dictation, iPhone keyboard, whisper, self-hosted, on-device, transcription, open source, privacy, voice to text, dictate, offline dictation';
const OG_IMAGE = `${SITE_URL}/og-image.png`;

function getBreadcrumbList(relativePath: string, title: string): string {
  const cleanPath = relativePath.replace(/\.md$/, '').replace(/\/index$/, '');
  const parts = cleanPath.split('/').filter(Boolean);
  const sectionLabels: Record<string, string> = { vs: 'Compare', features: 'Features' };
  const items: object[] = [{ '@type': 'ListItem', position: 1, name: 'Diction', item: SITE_URL }];
  if (parts.length === 1) {
    items.push({ '@type': 'ListItem', position: 2, name: title, item: `${SITE_URL}/${parts[0]}` });
  } else if (parts.length >= 2) {
    items.push({ '@type': 'ListItem', position: 2, name: sectionLabels[parts[0]] || parts[0], item: `${SITE_URL}/${parts[0]}` });
    items.push({ '@type': 'ListItem', position: 3, name: title, item: `${SITE_URL}/${cleanPath}` });
  }
  return JSON.stringify({ '@context': 'https://schema.org', '@type': 'BreadcrumbList', itemListElement: items });
}

export default defineConfig({
  vite: {
    plugins: [llmstxt()],
  },

  title: SITE_NAME,
  titleTemplate: `:title | ${SITE_NAME}`,
  description: DEFAULT_DESCRIPTION,
  sitemap: { hostname: SITE_URL },
  lastUpdated: true,

  // Dynamic page-level SEO via transformPageData
  transformPageData(pageData) {
    // Build canonical URL — match VitePress output format
    const canonicalUrl = `${SITE_URL}/${pageData.relativePath}`
      .replace(/index\.md$/, '')
      .replace(/\.md$/, '');

    // Homepage gets the full branded title, other pages get "Page | Diction"
    const isHomePage = pageData.relativePath === 'index.md';
    const pageTitle = pageData.frontmatter.title || pageData.title;
    const fullTitle = isHomePage
      ? pageTitle || SITE_TITLE
      : pageTitle
        ? `${pageTitle} | ${SITE_NAME}`
        : SITE_NAME;
    const pageDescription =
      pageData.frontmatter.description || pageData.description || DEFAULT_DESCRIPTION;
    const pageKeywords = pageData.frontmatter.keywords || DEFAULT_KEYWORDS;

    // Add dynamic head tags
    pageData.frontmatter.head ??= [];
    pageData.frontmatter.head.push(
      // Canonical URL — critical for SEO
      ['link', { rel: 'canonical', href: canonicalUrl }],
      // Robots
      ['meta', { name: 'robots', content: 'index, follow' }],
      // Keywords
      ['meta', { name: 'keywords', content: pageKeywords }],
      // Dynamic Open Graph
      ['meta', { property: 'og:title', content: fullTitle }],
      ['meta', { property: 'og:description', content: pageDescription }],
      ['meta', { property: 'og:url', content: canonicalUrl }],
      // Dynamic Twitter
      ['meta', { name: 'twitter:title', content: fullTitle }],
      ['meta', { name: 'twitter:description', content: pageDescription }],
    );

    // BreadcrumbList for inner pages
    if (!isHomePage) {
      pageData.frontmatter.head.push(
        ['script', { type: 'application/ld+json' }, getBreadcrumbList(pageData.relativePath, pageTitle || '')],
      );
    }
  },

  head: [
    // Preconnect for performance
    ['link', { rel: 'preconnect', href: 'https://www.googletagmanager.com' }],
    ['link', { rel: 'dns-prefetch', href: 'https://www.googletagmanager.com' }],
    // Google Analytics 4
    ['script', { async: '', src: 'https://www.googletagmanager.com/gtag/js?id=G-PCV64Y7GFM' }],
    [
      'script',
      {},
      `window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', 'G-PCV64Y7GFM');`,
    ],
    // Favicons — generated from app-icon.png
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32.png' }],
    ['link', { rel: 'icon', href: '/favicon.ico', sizes: '48x48' }],
    ['link', { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' }],
    // Apple Smart Banner — shows "Open in App Store" bar on Safari
    ['meta', { name: 'apple-itunes-app', content: 'app-id=6759807364' }],
    // Theme color for mobile browsers
    ['meta', { name: 'theme-color', content: '#007AFF' }],
    // Language
    ['meta', { property: 'og:locale', content: 'en_US' }],
    // Structured data — SoftwareApplication
    [
      'script',
      { type: 'application/ld+json' },
      JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        name: SITE_NAME,
        alternateName: 'diction',
        url: SITE_URL,
        logo: `${SITE_URL}/app-icon.png`,
        image: OG_IMAGE,
        screenshot: OG_IMAGE,
        description: DEFAULT_DESCRIPTION,
        applicationCategory: 'UtilitiesApplication',
        applicationSubCategory: 'Productivity',
        operatingSystem: 'iOS',
        softwareRequirements: 'iOS 17+',
        downloadUrl: 'https://apps.apple.com/app/id6759807364',
        installUrl: 'https://apps.apple.com/app/id6759807364',
        releaseNotes: 'https://github.com/omachala/diction/releases',
        keywords:
          'voice keyboard, speech to text, dictation, self-hosted, whisper, iOS keyboard, transcription, privacy',
        offers: {
          '@type': 'Offer',
          price: '0',
          priceCurrency: 'USD',
          availability: 'https://schema.org/InStock',
        },
        author: {
          '@type': 'Person',
          name: 'Ondrej Machala',
          url: 'https://github.com/omachala',
        },
        maintainer: {
          '@type': 'Person',
          name: 'Ondrej Machala',
          url: 'https://github.com/omachala',
        },
      }),
    ],
    // Structured data — FAQPage
    [
      'script',
      { type: 'application/ld+json' },
      JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'FAQPage',
        mainEntity: [
          {
            '@type': 'Question',
            name: 'Is Diction really free?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'On-device and self-hosted modes are completely free with no word limits, no daily caps, and no restrictions. Diction One cloud requires a paid subscription with a free trial included.',
            },
          },
          {
            '@type': 'Question',
            name: 'What languages does Diction support?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'Diction supports 99 languages via Whisper. The on-device model handles most languages well. Cloud and self-hosted modes use larger models for even better accuracy.',
            },
          },
          {
            '@type': 'Question',
            name: 'Is my voice data stored?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'Never. On-device mode processes audio in memory and discards it immediately. Self-hosted mode sends audio only to your server. Diction One cloud processes and discards — no recordings retained, no model training.',
            },
          },
          {
            '@type': 'Question',
            name: 'What is self-hosting?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'You run a Whisper speech-to-text server on your own hardware — a home server, NAS, or cloud VM. Diction connects to it directly. Your audio never touches any third-party service. One Docker Compose command to start.',
            },
          },
          {
            '@type': 'Question',
            name: 'How do I set up Diction?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'Open the app, grant mic permission, add Diction as a keyboard in iOS Settings, enable Full Access, and you are ready. Full setup takes under a minute.',
            },
          },
          {
            '@type': 'Question',
            name: 'Why does the keyboard need Full Access?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'iOS requires Full Access for any keyboard extension that uses the network. Diction needs it to send audio for transcription. Diction has no QWERTY keys to log, does not read your clipboard, and does not access contacts.',
            },
          },
          {
            '@type': 'Question',
            name: 'How is Diction different from Apple Dictation?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'Diction uses Whisper for higher accuracy, supports 99 languages, has no time limits, and lets you choose where audio is processed — on device, your server, or cloud. Apple Dictation has a 60-second limit and processes audio on Apple servers.',
            },
          },
          {
            '@type': 'Question',
            name: 'What is AI Enhancement?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'After transcription, Diction can optionally clean up your text — removing filler words like "um" and "uh", fixing grammar, adding punctuation, and polishing the result. Only the text is sent to the AI model, never the audio. AI Enhancement is off by default.',
            },
          },
          {
            '@type': 'Question',
            name: 'Does Diction work offline?',
            acceptedAnswer: {
              '@type': 'Answer',
              text: 'Yes. On-device mode works completely offline once the model is downloaded. No internet connection needed. Cloud and self-hosted modes require network access to reach the transcription server.',
            },
          },
        ],
      }),
    ],
    // Structured data — WebSite with SearchAction
    [
      'script',
      { type: 'application/ld+json' },
      JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'WebSite',
        name: SITE_NAME,
        url: SITE_URL,
        description: DEFAULT_DESCRIPTION,
        potentialAction: {
          '@type': 'SearchAction',
          target: {
            '@type': 'EntryPoint',
            urlTemplate: `${SITE_URL}/?q={search_term_string}`,
          },
          'query-input': 'required name=search_term_string',
        },
      }),
    ],
    // Structured data — Organization
    [
      'script',
      { type: 'application/ld+json' },
      JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'Organization',
        name: SITE_NAME,
        url: SITE_URL,
        logo: `${SITE_URL}/app-icon.png`,
        sameAs: ['https://github.com/omachala/diction'],
      }),
    ],
    // Static Open Graph (fallbacks — dynamic ones override these)
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: SITE_NAME }],
    ['meta', { property: 'og:image', content: OG_IMAGE }],
    ['meta', { property: 'og:image:width', content: '1200' }],
    ['meta', { property: 'og:image:height', content: '630' }],
    // Static Twitter (fallbacks)
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: OG_IMAGE }],
  ],

  themeConfig: {
    siteTitle: 'Diction',
    logo: '/app-icon.png',

    search: {
      provider: 'local',
    },

    nav: [
      { text: 'Features', link: '/features' },
      { text: 'Docs', link: '/on-device' },
      { text: 'Support', link: '/support' },
      {
        text: 'Download',
        link: 'https://apps.apple.com/app/id6759807364',
      },
    ],

    sidebar: {
      '/': [
        {
          text: 'Features',
          items: [
            { text: 'Overview', link: '/features/' },
            { text: 'Context-Aware Text Editing', link: '/features/context-aware' },
            { text: 'AI Enhancement', link: '/features/ai-enhancement' },
          ],
        },
        {
          text: 'Transcription Modes',
          items: [
            { text: 'On-Device', link: '/on-device' },
            { text: 'Self-Hosted', link: '/self-hosted' },
            { text: 'Diction One (Cloud)', link: '/cloud' },
          ],
        },
        {
          text: 'Security',
          items: [
            { text: 'Encryption', link: '/encryption' },
          ],
        },
        {
          text: 'Guides',
          items: [
            { text: 'Self-Hosting Setup', link: '/features/self-hosting-setup' },
            { text: 'Use Your Own Model', link: '/features/custom-model' },
          ],
        },
        {
          text: 'Legal',
          items: [
            { text: 'Privacy Policy', link: '/privacy' },
            { text: 'Terms of Service', link: '/terms' },
          ],
        },
        {
          text: 'Compare',
          items: [
            { text: 'vs Wispr Flow', link: '/vs/wispr-flow' },
            { text: 'vs Apple Dictation', link: '/vs/apple-dictation' },
            { text: 'vs Willow', link: '/vs/willow' },
            { text: 'vs Superwhisper', link: '/vs/superwhisper' },
          ],
        },
        {
          text: 'Help',
          items: [
            { text: 'Support', link: '/support' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/omachala/diction' },
    ],

    lastUpdated: {
      text: 'Last updated',
    },
  },
});
