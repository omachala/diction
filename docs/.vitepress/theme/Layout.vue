<script setup>
import DefaultTheme from 'vitepress/theme';
import HeroDemo from './HeroDemo.vue';
import SiteFooter from './SiteFooter.vue';

const { Layout } = DefaultTheme;
</script>

<template>
  <Layout>
    <template #home-hero-image>
      <HeroDemo />
    </template>
    <template #layout-bottom>
      <SiteFooter />
    </template>
  </Layout>
</template>
