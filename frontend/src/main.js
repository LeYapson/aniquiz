import { createApp } from 'vue'
import './style.css'
import AppRoot from './AppRoot.vue'
import router from './router'

createApp(AppRoot).use(router).mount('#app')
