import './_styles/app.scss';

import Vue from 'vue';

ELEMENT.locale(ELEMENT.lang.en);

import App from './app.vue';

(new Vue(App)).$mount('#app');
