import DefaultTheme from 'vitepress/theme'
import './custom.css'
import DemoLink from './components/DemoLink.vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // Register the DemoLink component globally
    app.component('DemoLink', DemoLink)
  }
}

