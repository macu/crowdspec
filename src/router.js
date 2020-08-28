import VueRouter from 'vue-router';
import store from './store.js';
import {idsEq} from './utils.js';

import IndexPage from './pages/index.vue';
import SpecPage from './pages/spec.vue';
import SubspecPage from './pages/subspec.vue';
import AjaxErrorPage from './pages/ajax-error.vue';
import NotFoundPage from './pages/not-found.vue';

export const router = new VueRouter({
	mode: 'history',
	routes: [
		{name: 'index', path: '/', component: IndexPage},
		{name: 'spec', path: '/spec/:specId', component: SpecPage},
		{name: 'subspec', path: '/spec/:specId/subspec/:subspecId', component: SubspecPage},
		{name: 'ajax-error', path: '/ajax-error/:code', component: AjaxErrorPage},
		{path: '*', component: NotFoundPage},
	],
	scrollBehavior(to, from, savedPosition) {
		// scrollBehavior is called after the new route has been rendered.
		// save to allow restoring scroll position following additional DOM updates
		// made during the route mounted hook
		store.commit('setSavedScrollPosition', savedPosition);
		return to.hash ? {selector: to.hash}
			: (savedPosition ? savedPosition : {x: 0, y: 0});
	},
});

router.beforeEach((to, from, next) => {
	// beforeEach is called before navigation is confirmed.
	if (from) {
		if (
			!!store.state.currentSpecId &&
			(!to.params.specId || !idsEq(to.params.specId, store.state.currentSpecId))
		) {
			// Leaving spec context for another context
			store.commit('clearCurrentSpec');
		} else if (from.name === 'spec') {
			// Leaving spec page for another page in same spec
			store.commit('saveCurrentSpecScrollTop', from.params.specId);
		}
	}
	next();
});

router.afterEach((to, from) => {
	// afterEach is called after navigation is confirmed,
	// but before the new route has been rendered.
	// clear saved scroll position of previous route
	store.commit('clearSavedScrollPosition');
});

export default router;
