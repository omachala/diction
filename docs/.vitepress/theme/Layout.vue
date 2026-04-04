<script setup>
import DefaultTheme from 'vitepress/theme';
import HeroVideo from './HeroVideo.vue';
import SiteFooter from './SiteFooter.vue';

const { Layout } = DefaultTheme;
</script>

<template>
  <Layout>
    <template #home-hero-image>
      <HeroVideo />
    </template>
    <template #layout-bottom>
      <SiteFooter />
    </template>
  </Layout>
</template>
