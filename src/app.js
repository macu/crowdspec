import './_styles/app.scss';

import {createApp} from 'vue';
import ElementPlus from 'element-plus';
import ElementPlusLocaleEn from 'element-en';

import App from './app.vue';
import router from './router.js';
import store, {registerApp} from './store.js';

export const app = createApp(App);

registerApp(app); // register for programmatically instantiated components

window.app = app; // required for appContext in mounted blocks

app.use(ElementPlus, {
	locale: ElementPlusLocaleEn,
});
app.use(router);
app.use(store);

app.mount('#app');
