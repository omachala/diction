<script setup>
const props = defineProps({
  name: { type: String, required: true },
  size: { type: String, default: '1.2em' },
})
</script>

<template>
  <img
    :src="`/icon-${props.name}.svg`"
    :alt="props.name"
    class="inline-icon"
    :style="{ width: props.size, height: props.size }"
  />
</template>

<style scoped>
.inline-icon {
  display: inline-block;
  vertical-align: middle;
  margin: 0 0.15em;
}
</style>

<style>
.dark .inline-icon {
  filter: invert(1);
}
</style>
