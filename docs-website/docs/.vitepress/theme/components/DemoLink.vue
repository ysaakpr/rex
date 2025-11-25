<template>
  <a :href="demoUrl" class="VPButton medium brand" target="_blank" rel="noopener">
    <span class="text">Try Demo</span>
  </a>
</template>

<script setup>
import { computed } from 'vue'

// Get the demo URL from the global define
const demoUrl = computed(() => {
  if (typeof __DEMO_URL__ !== 'undefined') {
    return __DEMO_URL__
  }
  return 'http://localhost/demo'
})
</script>

<style scoped>
.VPButton {
  display: inline-block;
  border: 1px solid transparent;
  text-align: center;
  font-weight: 600;
  white-space: nowrap;
  transition: color 0.25s, border-color 0.25s, background-color 0.25s;
  text-decoration: none;
}

.VPButton.medium {
  border-radius: 20px;
  padding: 0 20px;
  line-height: 38px;
  font-size: 14px;
}

.VPButton.brand {
  border-color: var(--vp-button-brand-border);
  color: var(--vp-button-brand-text);
  background-color: var(--vp-button-brand-bg);
}

.VPButton.brand:hover {
  border-color: var(--vp-button-brand-hover-border);
  color: var(--vp-button-brand-hover-text);
  background-color: var(--vp-button-brand-hover-bg);
}
</style>

