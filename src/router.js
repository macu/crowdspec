import VueRouter from 'vue-router';

import IndexPage from './pages/index.vue';
import NewSpecPage from './pages/new-spec.vue';
import SpecPage from './pages/spec.vue';
import AjaxErrorPage from './pages/ajax-error.vue';
import NotFoundPage from './pages/not-found.vue';

export default new VueRouter({
	mode: 'history',
	routes: [
		{name: 'index', path: '/', component: IndexPage},
		{name: 'new-spec', path: '/new-spec', component: NewSpecPage},
		{name: 'spec', path: '/spec/:specId', component: SpecPage},
		{name: 'ajax-error', path: '/ajax-error/:code', component: AjaxErrorPage},
		{path: '*', component: NotFoundPage},
	],
});
